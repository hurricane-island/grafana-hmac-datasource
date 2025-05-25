package plugin

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
	"unicode/utf8"
	"strings"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
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
	hmacString:= strings.Join(hmacData, "\n")
	if FIXED_WIDTH != 0 {
		runeCount := utf8.RuneCountInString(hmacString)
		if runeCount != FIXED_WIDTH{
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
	// if true {
	// 	t.Fatal("Request succeeded with: ", string(body))
	// }
}

func TestQueryDataStreams(t *testing.T) {
	client := http.Client{}
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	req, err := signedGetRequest(SERVER_URL, "/site/6809170ead845d428de9a636/datastreams", clientId, secretKey, AUTH_METHOD, DELIMITER)
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
	// if true {
	// 	t.Fatal("Request succeeded with: ", string(body))
	// }
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
