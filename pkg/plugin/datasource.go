package plugin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"math"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/hurricane-island/grafana-hmac-datasource/pkg/models"
)

const ISO_COMPATIBILITY = "2006-01-02T15:04:05.000Z"
const INDEX_NAME = "/sites"
const QUERY_PATH = "/observations"

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Array of string data to encode.
func hmacStringArray(
	// Signature time
	date time.Time,
	// API key identifying multi-tenant client
	clientId string,
	// All API path segments as string
	path string,
) []string {
	return []string{
		"GET", // Only GET is supported
		"",    // content type of GET is empty string
		date.Format(ISO_COMPATIBILITY),
		path,
		"", // service headers is empty string
		"", // content checksum is empty string for GET
		clientId,
	}
}

// HMAC-SHA256 signature of the data required for a GET request
func signedHmacBytes(data string, signingKey string) []byte {
	words, _ := base64.StdEncoding.DecodeString(signingKey)
	mac := hmac.New(sha256.New, words)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// Compose valid authorization header with HMAC key.
func authHeader(authMethod string, clientId string, hmac []byte) string {
	encodedClientId := base64.StdEncoding.EncodeToString([]byte(clientId))
	hmacBase64 := base64.StdEncoding.EncodeToString(hmac)
	auth := authMethod + " " + encodedClientId + ":" + hmacBase64
	return auth
}

// Produce a GET request with HMAC signature.
func signedGetRequest(server string, path string, clientId string, secretKey string, authMethod string, delim string) (*http.Request, error) {
	date := time.Now().UTC()
	data := hmacStringArray(date, clientId, path)
	hmac := signedHmacBytes(strings.Join(data, delim), secretKey)
	auth := authHeader(authMethod, clientId, hmac)
	url := server + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return req, err
	}
	isoDate := date.Format(ISO_COMPATIBILITY)
	req.Header.Add("Authorization", auth)
	req.Header.Add("Date", isoDate)
	return req, err
}

// Construct an empty datasource instance.
func NewDatasource(_ context.Context, _ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type QueryModel struct {
	ThingId string `json:"thingId"`
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse
	var qm QueryModel
	instanceSettings := pCtx.DataSourceInstanceSettings
	clientId := instanceSettings.DecryptedSecureJSONData["clientId"]
	secretKey := instanceSettings.DecryptedSecureJSONData["secretKey"]
	config, err := models.LoadPluginSettings(*pCtx.DataSourceInstanceSettings)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("load plugin settings: %v", err.Error()))
	}
	err = json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}
	datastreamsUrl := config.BasePath + "/site" + "/" + qm.ThingId + "/datastreams"
	req, err := signedGetRequest(config.ServerUrl, datastreamsUrl, clientId, secretKey, config.AuthMethod, "\n")
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request: %v", err.Error()))
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request: %v", err.Error()))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("body: %v", err.Error()))
	}
	if resp.StatusCode != 200 {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request: %v", string(body)))
	}
	var datastreams []models.DataStream
	err = json.Unmarshal(body, &datastreams)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unmarshal: %v",err.Error()))
	}
	var datastreamIds []string
	var lookup = make(map[string]string)
	for _, ds := range datastreams {
		datastreamIds = append(datastreamIds, ds.Id)
		lookup[ds.Id] = ds.Name
	}
	from := query.TimeRange.From.Format(ISO_COMPATIBILITY)
	until := query.TimeRange.To.Format(ISO_COMPATIBILITY)
	path := config.BasePath + QUERY_PATH + "?from=" + from + "&until=" + until + "&datastreamIds=" + strings.Join(datastreamIds, ",")
	
	getReq, err := signedGetRequest(config.ServerUrl, path, clientId, secretKey, config.AuthMethod, "\n")
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("signed request: %v", err.Error()))
	}
	resp, err = client.Do(getReq)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request failed: %v", err.Error()))
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("reading body: %v", err.Error()))
	}
	if resp.StatusCode != 200 {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request failed: %v", string(body)))
	}
	var partial map[string]json.RawMessage
	err = json.Unmarshal(body, &partial)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("partial unmarshaling failed: %v", err.Error()))
	}
	for k, v := range partial {
		var obs []models.Observation
		err = json.Unmarshal(v, &obs)
		if err != nil {
			continue
			// return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unmarshaling observation failed: %v", name))
		}
		t := make([]time.Time, len(obs))
		value := make([]float64, len(obs))
		count := 0
		for i, observation := range obs {
			if math.IsNaN(observation.Value) {
				// Skip NaN values
				continue
			}
			t[i] = time.Unix(0, observation.PhenomenonTime * int64(time.Millisecond))
			value[i] = observation.Value
			count += 1
		}
		// https://grafana.com/developers/plugin-tools/introduction/data-frames
		if count == 0 {
			// If no valid observations, skip this datastream
			continue
		}
		name := lookup[k]
		frame := data.NewFrame(name)
		frame.Fields = append(frame.Fields,
			data.NewField("phenomenonTime", nil, t),
			data.NewField("value", nil, value),
		)
		response.Frames = append(response.Frames, frame)
	}
	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)

	res.Status = backend.HealthStatusError
	if err != nil {
		res.Message = "Unable to load settings"
		return res, nil
	}

	if config.Secrets.SecretKey == "" {
		res.Message = "HMAC signing key is missing"
		return res, nil
	}

	if config.Secrets.ClientId == "" {
		res.Message = "Client ID is missing"
		return res, nil
	}

	if config.ServerUrl == "" {
		res.Message = "Server URL is missing"
		return res, nil
	}

	if config.BasePath == "" {
		res.Message = "BasePath is missing"
		return res, nil
	}

	if config.AuthMethod == "" {
		res.Message = "Auth method is missing"
		return res, nil
	}
	path := config.BasePath + INDEX_NAME
	client := http.Client{}
	getReq, err := signedGetRequest(
		config.ServerUrl, path, config.Secrets.ClientId,
		config.Secrets.SecretKey, config.AuthMethod, "\n")
	if err != nil {
		res.Message = "Request failed:" + err.Error()
		return res, nil
	}
	resp, err := client.Do(getReq)
	if err != nil {
		res.Message = "Request failed:" + err.Error()
		return res, nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res.Message = "Error reading body:" + err.Error()
		return res, nil
	}
	if resp.StatusCode != 200 {
		res.Message = "Request failed:" + string(body)
		return res, nil
	}
	var things []models.Thing
	err = json.Unmarshal(body, &things)
	if err != nil {
		res.Message = "Unmarshaling failed:" + err.Error()
		return res, nil
	}
	if len(things) == 0 {
		res.Message = "No root nodes found"
		return res, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
