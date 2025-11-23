// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed modules/sku.json
var skuConfigData []byte

//go:embed propertymaps/propertyMaps.json
var propertyMapsData []byte

// skuConfig defines how to extract SKU information for a resource type
type skuConfig struct {
	ResourceType                     string   `json:"resourceType"`
	Property                         string   `json:"property"`
	Description                      string   `json:"description"`
	IsContainedInOriginalGraphOutput bool     `json:"isContainedInOriginalGraphOutput"`
	ExcludeFromReport                []string `json:"excludeFromReport"`
}

// propertyMapConfig defines how to query SKU availability for a resource type
type propertyMapConfig struct {
	ResourceType string `json:"resourceType"`
	URI          string `json:"uri"`
	RegionalAPI  bool   `json:"regionalApi"`
	Properties   struct {
		StartPath          []string          `json:"startPath"`
		TopLevelProperties map[string]string `json:"topLevelProperties"`
		ChildProperties    *struct {
			Name  []string          `json:"name"`
			Props map[string]string `json:"props"`
		} `json:"childProperties,omitempty"`
	} `json:"properties"`
}

var (
	skuConfigs         []skuConfig
	propertyMapsConfig []propertyMapConfig
)

// init loads configuration files
func init() {
	// Load SKU configurations
	if err := json.Unmarshal(skuConfigData, &skuConfigs); err != nil {
		log.Fatal().Err(err).Msg("Failed to load SKU configuration from embedded JSON")
	}

	// Load property maps configurations
	if err := json.Unmarshal(propertyMapsData, &propertyMapsConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to load property maps configuration from embedded JSON")
	}
}

// getPropertyMapConfig returns the property map configuration for a given resource type
func getPropertyMapConfig(resourceType string) *propertyMapConfig {
	resourceType = strings.ToLower(resourceType)
	for i := range propertyMapsConfig {
		if strings.ToLower(propertyMapsConfig[i].ResourceType) == resourceType {
			return &propertyMapsConfig[i]
		}
	}
	return nil
}
