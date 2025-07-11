package plugin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/hurricane-island/grafana-hmac-datasource/pkg/models"
)

// Equivalent to JavaScript's Date.toISOString() format.
const ISO_COMPATIBILITY = "2006-01-02T15:04:05.000Z"
// Base path for indexing available resources.
const INDEX_NAME = "sites"
// Path to query for time series data
const QUERY_PATH = "/observations"
// Name of time query in service API.
const QUERY_START = "from"
// Name of time query in service API.
const QUERY_END = "until"
// Name of query parameter for time series tags.
const QUERY_TAGS = "datastreamIds"
// Root path for querying data streams.
const QUERY_ROOT = "site"
// Second path element for querying data streams.
const QUERY_COLLECTION = "datastreams"

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

// Construct an empty datasource instance. Called as Factory method in main.go
// Can pass in the instance settings, which are used to configure the datasource,
// so that secrets can be access from resource calls.
func NewDatasource(_ context.Context, instanceSettings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(instanceSettings)
	if err != nil {
		return nil, err
	}
	return &Datasource{
		Config: config,
		Client: &http.Client{},
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{
	Config *models.PluginSettings
	Client *http.Client
}

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

// Implement a generic resource handler for the datasource.
// This will need a switch statement to handle different paths.
func (d *Datasource) CallResource(
	// Unused
	_ context.Context,
	// Contains API path to query
	req *backend.CallResourceRequest,
	// Response handler
	sender backend.CallResourceResponseSender,
) error {
	path := d.Config.BasePath + "/" + req.Path
	getReq, err := d.request(path)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body: []byte(err.Error()),
		})
	}
	resp, err := d.Client.Do(getReq)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body: []byte(err.Error()),
		})
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body: []byte(err.Error()),
		})
	}
	if resp.StatusCode != 200 {
		return sender.Send(&backend.CallResourceResponse{
			Status: resp.StatusCode,
			Body: []byte(body),
		})
	}
	var things []models.ThingWithLocation
	err = json.Unmarshal(body, &things)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body: []byte(err.Error()),
		})
	}
	resource := make([]models.ThingWithDataStreams, 0, len(things))
	for _, thing := range things {
		parts := []string{d.Config.BasePath, QUERY_ROOT, thing.Id, QUERY_COLLECTION}
		url := strings.Join(parts, "/")
		getReq, err = d.request(url)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body: []byte(err.Error()),
			})
		}
		resp, err = d.Client.Do(getReq)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body: []byte(err.Error()),
			})
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body: []byte(err.Error()),
			})
		}
		if resp.StatusCode != 200 {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body: []byte(body),
			})
		}
		var dataStreams []models.DataStream
		err = json.Unmarshal(body, &dataStreams)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body: []byte(err.Error()),
			})
		}
		resource = append(resource, models.ThingWithDataStreams{
			Thing:        thing,
			DataStreams: dataStreams,
		})
	}
	
	result, err := json.Marshal(resource)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body: []byte(err.Error()),
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   result,
		Headers: map[string][]string{	
			"Content-Type": {"application/json"},
		},
	})
}

// Selection data from the frontend query editor
type QueryModel struct {
	ThingId string `json:"thingId"`
}

// Convenience function to make request with configured secrets and params.
func (d *Datasource) request(path string) (*http.Request, error) {
	return signedGetRequest(d.Config.ServerUrl, path, d.Config.Secrets.ClientId, d.Config.Secrets.SecretKey, d.Config.AuthMethod, "\n")
}

// Handler for a single frontend query.
func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse
	var qm QueryModel
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}
	parts := []string{d.Config.BasePath, QUERY_ROOT, qm.ThingId, QUERY_COLLECTION}
	url := strings.Join(parts, "/")
	req, err := d.request(url)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request: %v", err.Error()))
	}
	resp, err := d.Client.Do(req)
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
	var dataStreams []models.DataStream
	err = json.Unmarshal(body, &dataStreams)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unmarshal: %v", err.Error()))
	}
	var tags []string
	var lookup = make(map[string]string)
	for _, ds := range dataStreams {
		tags = append(tags, ds.Id)
		lookup[ds.Id] = ds.Name
	}
	from := query.TimeRange.From.Format(ISO_COMPATIBILITY)
	until := query.TimeRange.To.Format(ISO_COMPATIBILITY)
	path := d.Config.BasePath + QUERY_PATH + 
		"?" + QUERY_START + "=" + from + 
		"&" + QUERY_END + "=" + until + 
		"&" + QUERY_TAGS + "=" + strings.Join(tags, ",")

	getReq, err := d.request(path)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("signed request: %v", err.Error()))
	}
	resp, err = d.Client.Do(getReq)
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
		}
		t := make([]time.Time, len(obs))
		value := make([]float64, len(obs))
		count := 0
		for i, observation := range obs {
			if math.IsNaN(observation.Value) {
				continue
			}
			t[i] = time.Unix(0, observation.PhenomenonTime*int64(time.Millisecond))
			value[i] = observation.Value
			count += 1
		}
		if count == 0 {
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
	res := &backend.CheckHealthResult{
		Status: backend.HealthStatusError,
	}
	if d.Config.Secrets.SecretKey == "" {
		res.Message = "HMAC signing key is missing"
		return res, nil
	}
	if d.Config.Secrets.ClientId == "" {
		res.Message = "Client ID is missing"
		return res, nil
	}
	if d.Config.ServerUrl == "" {
		res.Message = "Server URL is missing"
		return res, nil
	}
	if d.Config.BasePath == "" {
		res.Message = "BasePath is missing"
		return res, nil
	}
	if d.Config.AuthMethod == "" {
		res.Message = "AuthMethod is missing"
		return res, nil
	}
	path := strings.Join([]string{d.Config.BasePath, INDEX_NAME}, "/")
	getReq, err := d.request(path)
	if err != nil {
		res.Message = "Request failed:" + err.Error()
		return res, nil
	}
	resp, err := d.Client.Do(getReq)
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
	var things []models.ThingWithLocation
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
