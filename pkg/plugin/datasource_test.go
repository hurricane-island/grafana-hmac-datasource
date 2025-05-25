package plugin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const REFERENCE_DATE = "2025-05-25T13:24:56.789Z"
const REFERENCE_ENDPOINT = "/xcloud/data-export/sites"

func TestIso8061Date(t *testing.T) {
	date, _ := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	isoDate := date.Format(ISO_COMPATIBILITY)
	// if true {
	// 	today := time.Now().UTC()
	// 	isoToday := today.Format(ISO_COMPATIBILITY)
	// 	t.Fatal("ISO 8061 date = ", isoToday)
	// }
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
	// Check against implementation known to work...
	// encoded := base64.StdEncoding.EncodeToString([]byte(clientId))
	// if true {
	// 	t.Fatal("CLIENT_ID encoded = ", encoded)
	// }
}

func TestSecretKeyExists(t *testing.T) {
	secret := os.Getenv("SECRET_KEY")
	byte_length := len(secret)
	if byte_length == 0 {
		t.Fatal("SECRET_KEY environment variable byte length = ", byte_length)
	}
}

func TestHmacBufferString(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	date, _ := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	buffer := hmacBufferString(date, clientId, REFERENCE_ENDPOINT)
	rune_length := utf8.RuneCountInString(buffer)
	// message will always be 94 characters long
	if rune_length != 94 {
		t.Fatal("HMAC character length =", rune_length)
	}
	secretKeyWords, _ := base64.StdEncoding.DecodeString(secretKey)
	mac := hmac.New(sha256.New, secretKeyWords)
	mac.Write([]byte(buffer))
	checksum := mac.Sum(nil)
	base64checksum := base64.StdEncoding.EncodeToString(checksum)
	if true {
		t.Fatal("HMAC checksum = ", base64checksum)
	}
}

func TestHmacBytes(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	date, _ := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	hmac := SignedHmacBytes(date, clientId, secretKey, "/xcloud/data-export/sites")
	if len(hmac) == 0 {
		t.Fatal("HMAC byte length = ", len(hmac))
	}

}

func TestHmacBase64String(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	date, _ := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	hmac := SignedHmacBase64String(date, clientId, secretKey, "/xcloud/data-export/sites")
	if true {
		t.Fatal("HMAC length = ", len(hmac))
	}
}

func TestAuthHeader(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	date, _ := time.Parse(time.RFC3339Nano, REFERENCE_DATE)
	isoDate := date.Format(ISO_COMPATIBILITY)
	auth := authHeader(date, "xCloud", clientId, secretKey, "/xcloud/data-export/sites")
	// if len(auth) == 0 {
	if true {
		t.Fatal("Authorization =", auth, "Date =", isoDate)
	}
}

func TestQueryThings(t *testing.T) {
	client := http.Client{}
	date := time.Now().UTC()
	clientId := os.Getenv("CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	path := "/xcloud/data-export/sites"
	auth := authHeader(date, "xCloud", clientId, secretKey, path)
	url := "https://cloud.xylem.com" + path
	req, _ := http.NewRequest("GET", url, nil)
	// error handling
	isoDate := date.Format(ISO_COMPATIBILITY)
	// if true {
	// 	t.Fatal("Headers = ", isoDate, auth)
	// }
	req.Header.Add("Authorization", auth)
	req.Header.Add("Date", isoDate)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Request failed with: ", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
	}
	if resp.StatusCode != 200 {
		t.Fatal("Request failed with: ", isoDate, string(body))
	}
	if true {
		t.Fatal("Request succeeded with: ", string(body))
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
