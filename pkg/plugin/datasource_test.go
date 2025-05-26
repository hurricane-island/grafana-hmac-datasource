package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/grafana/grafana-plugin-sdk-go/backend"

	"github.com/hurricane-island/grafana-hmac-datasource/pkg/models"
)

const REFERENCE_DATE = "2025-05-25T13:24:56.789Z"
const REFERENCE_ENDPOINT = "/xcloud/data-export/sites"
const SERVER_URL = "https://cloud.xylem.com"
const FIXED_WIDTH = 94
const AUTH_METHOD = "xCloud"
const DELIMITER = "\n"

func TestIso8061Date(t *testing.T) {
	date, _ := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	isoDate := date.Format(ISO_COMPATIBILITY)
	if isoDate != REFERENCE_DATE {
		t.Fatal("ISO 8061 date not reversible, ", isoDate)
	}
}

func TestClientIdExists(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	byte_length := len(clientId)
	if byte_length == 0 {
		t.Fatal("CLIENT_ID environment variable byte length = ", byte_length)
	}
}

func TestSecretKeyExists(t *testing.T) {
	secret := os.Getenv("SECRET_KEY")
	byte_length := len(secret)
	if byte_length == 0 {
		t.Fatal("SECRET_KEY environment variable byte length = ", byte_length)
	}
}

// Run through signing process
func TestHmacBytes(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	date, err := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	if err != nil {
		t.Fatal("Error parsing date:", err)
	}
	hmacData := hmacStringArray(date, clientId, REFERENCE_ENDPOINT)
	hmacString := strings.Join(hmacData, "\n")
	if FIXED_WIDTH != 0 {
		runeCount := utf8.RuneCountInString(hmacString)
		if runeCount != FIXED_WIDTH {
			t.Fatal("HMAC character length =", runeCount)
		}
	}
	hmac := signedHmacBytes(hmacString, "any-secret-key")
	if len(hmac) == 0 {
		t.Fatal("HMAC byte length = ", len(hmac))
	}
}

func TestQueryThings(t *testing.T) {
	client := http.Client{}
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	req, err := signedGetRequest(SERVER_URL, REFERENCE_ENDPOINT, clientId, secretKey, AUTH_METHOD, DELIMITER)
	if err != nil {
		t.Fatal("Request failed with: ", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Request failed with:", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Error reading body:", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("Request failed with:", string(body))
	}
	var things []models.ThingWithLocation
	err = json.Unmarshal(body, &things)
	if err != nil {
		t.Fatal("Error unmarshaling response:", err)
	}
	if true {
		t.Fatal("Request succeeded with: ", things)
	}
}

func TestQueryDataStreams(t *testing.T) {
	client := http.Client{}
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	datastreamsUrl := "/xcloud/data-export/site/6809170ead845d428de9a636/datastreams"
	req, err := signedGetRequest(SERVER_URL, datastreamsUrl, clientId, secretKey, AUTH_METHOD, DELIMITER)
	if err != nil {
		t.Fatal("Request failed with: ", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Request failed with:", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Error reading body:", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("Request failed with:", string(body))
	}
	var datastreams []models.DataStream
	err = json.Unmarshal(body, &datastreams)
	if err != nil {
		t.Fatal("Error unmarshaling response:", err)
	}
	if true {
		t.Fatal("Request succeeded with:", len(datastreams))
	}
}

func TestQueryObservations(t *testing.T) {
	client := http.Client{}
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	url := "/xcloud/data-export/observations?datastreamIds=2015785,2015786&from=2025-05-20T00:00:00.000Z&until=2025-05-25T00:00:00.000Z"
	req, err := signedGetRequest(SERVER_URL, url, clientId, secretKey, AUTH_METHOD, DELIMITER)
	if err != nil {
		t.Fatal("Request failed with:", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Request failed with:", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Error reading body:", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("Request failed with:", string(body))
	}
	var idMap map[string]json.RawMessage
	err = json.Unmarshal(body, &idMap)
	if err != nil {
		t.Fatal("Error unmarshaling response:", err)
	}
	var observations = make(map[string][]models.Observation)
	for k, v := range idMap {
		var obs []models.Observation
		err = json.Unmarshal(v, &obs)
		if err != nil {
			t.Fatal("Error unmarshaling observation:", err)
		}
		observations[k] = obs
	}
	if true {
		t.Fatal("Request succeeded with:", observations)
	}
}

func TestQueryData(t *testing.T) {
	ds := Datasource{}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}
