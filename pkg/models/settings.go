package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// Container data type returned to the frontend for populating
// query editor dropdowns and other UI elements.
type ThingWithDataStreams struct {
	Thing ThingWithLocation `json:"thing"`
	DataStreams []DataStream `json:"dataStreams"`
}

// SensorThings API Thing, with nested Location.
// Schema is determine by the API that the plugin
// integrates with, and propagates to the frontend.
type ThingWithLocation struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    []struct {
		Latitude  float32 `json:"latitude"`
		Longitude float32 `json:"longitude"`
	} `json:"location"`
}

// SensorThings API Datastream.
// Schema is determine by the API that the plugin
// integrates with, and propagates to the frontend.
type DataStream struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	UnitOfMeasurement struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	}
}

// SensorThings API Observation.
// Schema is determine by the API that the plugin
// integrates with, and propagates to the frontend.
type Observation struct {
	Value          float64 `json:"value"`
	PhenomenonTime int64   `json:"phenomenonTime"`
}

// Info set during plugin initialization, including
// plaintext and secure settings.
type PluginSettings struct {
	ServerUrl  string                `json:"serverUrl"`
	BasePath   string                `json:"basePath"`
	AuthMethod string                `json:"authMethod"`
	Secrets    *SecretPluginSettings `json:"-"`
}

// Secrets set in plugin configuration.
type SecretPluginSettings struct {
	SecretKey string `json:"secretKey"`
	ClientId  string `json:"clientId"`
}

// Used in datasource initialization to load
// the plugin settings from the datasource instance settings.
func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}
	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)
	return &settings, nil
}

// Convert unstructured source map to SecretPluginSettings.
func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		SecretKey: source["secretKey"],
		ClientId:  source["clientId"],
	}
}
