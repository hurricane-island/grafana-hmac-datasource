import Crypto from "crypto-js";
const {HmacSHA256, enc} = Crypto;
const {Base64} = enc;

// Partial data used in creating the HMAC key
interface RoutingParams {
    authMethod: string
    clientId: string
    secretKey: string
}
// Headers function
interface HeaderParams extends RoutingParams {
    path: string
}
// Auth header function
interface AuthParams extends HeaderParams {
    date: Date
}
// All requests require the same base set of parameters
interface RequestParams extends RoutingParams {
    server: string
    path: string
}
interface ThingsParams extends Omit<RequestParams, "path"> {}
interface DataStreamsParams extends Omit<RequestParams, "path"> {
    thingId: string
}
interface ObservationsParams extends Omit<RequestParams, "path"> {
    datastreamIds: string[]
    from: Date
    until: Date
}

/**
 * Compose a valid HMAC key and output string to use as Authorization
 * header value. This only works for GET requests, which is
 * the only method supported by the partner data access API.
 * 
 * Not used in plugin, but preserved for testing for regression
 * against the original implementation.
 */
function hmacAuth({
    date,
    path,
    authMethod,
    clientId,
    secretKey
}: AuthParams) {
    const message = [
        "GET", // http method, 
        "", // content type is empty string for GET,
        date.toISOString(), // ISO timestamp
        path, // path
        "", // service-specific headers
        "", // content MD5 is empty string for GET
        clientId // multi-tenant client ID
    ].join("\n");
    const decodedSecretKey = Base64.parse(secretKey)
    const hmac = HmacSHA256(message, decodedSecretKey).toString(Base64);
    return `${authMethod} ${btoa(clientId)}:${hmac}`
}
/**
 * Make a fetch request with the HMAC authorization header.
 */
export function hmacHeaders({
    path,
    ...rest
}: HeaderParams) {
    const date = new Date();
    const headers = {
        Date: date.toISOString(),
        Authorization: hmacAuth({
            date,
            path,
            ...rest
        })
    }
    return headers
}

export function getRequestWithHmac({server, path, ...rest}: RequestParams) {
    const headers = hmacHeaders({
        path,
        ...rest
    });
    return fetch(`${server}${path}`, {
        headers
    });
}
/**
 * Query the production API for all sites associated with our
 * account. These contain display and location information that
 * can be used to construction queries for individual data streams.
 */
export function getThings({...rest}: ThingsParams) {
    const path = `/sites`
    return getRequestWithHmac({
        ...rest,
        path,
    })
}

/**
 * Retrieve all datastreams associate with a single site. Note
 * that the base path is `site` instead of `sites`. Otherwise
 * returns a server error, rather than a 404 error.
 */
export function getDataStreams({
    thingId,
    ...rest
}: DataStreamsParams) {
    const path = `/site/${thingId}/datastreams`;
    return getRequestWithHmac({
        ...rest,
        path,
    })
}

/**
 * Retrieve batch observations. Rather than using a path variable, 
 * this endpoint uses query parameters. You can get observations 
 * from multiple datastreams in a single request.
 */
export function getObservations({
    datastreamIds,
    from,
    until,
    ...rest
}: ObservationsParams) {
    const keyValuePairs = [
        ["from", from.toISOString()],
        ["until", until.toISOString()],
    ]
    for (const datastream of datastreamIds) {
        keyValuePairs.push(["datastreamIds", datastream]);
    }
    const query = new URLSearchParams(keyValuePairs);
    const path = `/observations?${query.toString()}`;
    return getRequestWithHmac({
        ...rest,
        path
    });
}

