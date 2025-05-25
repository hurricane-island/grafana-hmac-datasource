package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type Thing struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    []struct {
		Latitude  float32 `json:"latitude"`
		Longitude float32 `json:"longitude"`
	} `json:"location"`
}

type DataStream struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	UnitOfMeasurement struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	}
}

type Observation struct {
	Value          float32 `json:"value"`
	PhenomenonTime int     `json:"phenomenonTime"`
}

type PluginSettings struct {
	ServerUrl  string                `json:"serverUrl"`
	BasePath   string                `json:"basePath"`
	AuthMethod string                `json:"authMethod"`
	Secrets    *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	SecretKey string `json:"secretKey"`
	ClientId  string `json:"clientId"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}
	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)
	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		SecretKey: source["secretKey"],
		ClientId:  source["clientId"],
	}
}
