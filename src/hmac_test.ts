import { getThings } from "./hmac";

const CLIENT_ID = process.env.CLIENT_ID ?? "";
const SECRET_KEY = process.env.SECRET_KEY ?? "";
const SERVER_URL = process.env.SERVER_URL ?? "";

console.info("Querying things...")
const response = await getThings({
    server: SERVER_URL,
    clientId: CLIENT_ID,
    secretKey: SECRET_KEY,
    authMethod: "xCloud",
});
const things = await response.json();
console.log(JSON.stringify(things, null, 2));

// console.info(`Querying datastreams at site ${exampleSite.id}...`);
// const datastreamsResponse = await getDatastreams(exampleSite.id);
// const datastreams = await datastreamsResponse.json();
// const datastreamIds = datastreams.map((datastream) => {
//     return datastream.id
// });
// // console.log(JSON.stringify(datastreams, null, 2));

// console.info(`Querying observations...`);
// const observationsResponse = await getObservations(
//     datastreamIds.slice(0, 1),
//     queryRange
// );
// const observations = await observationsResponse.json();
// console.log(JSON.stringify(observations, null, 2));