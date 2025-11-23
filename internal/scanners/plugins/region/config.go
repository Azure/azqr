// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/models"
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

// getSKUConfig returns the SKU configuration for a given resource type
func getSKUConfig(resourceType string) *skuConfig {
	resourceType = strings.ToLower(resourceType)
	for i := range skuConfigs {
		if strings.ToLower(skuConfigs[i].ResourceType) == resourceType {
			return &skuConfigs[i]
		}
	}
	return nil
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

// extractSKUFromResource extracts SKU information from a resource based on configuration
func extractSKUFromResource(resource interface{}, config *skuConfig) (skuInfo, error) {
	if config == nil {
		return skuInfo{}, fmt.Errorf("no SKU configuration available")
	}

	// Handle *models.Resource by extracting the Properties field
	var resourceData map[string]interface{}
	switch r := resource.(type) {
	case *models.Resource:
		if r.Properties != nil {
			resourceData = r.Properties
		} else {
			return skuInfo{}, fmt.Errorf("resource has no properties")
		}
	case map[string]interface{}:
		resourceData = r
	default:
		return skuInfo{}, fmt.Errorf("unexpected resource type: %T", resource)
	}

	// Navigate the property path
	value := navigateProperty(resourceData, config.Property)
	if value == nil {
		return skuInfo{}, fmt.Errorf("property %s not found", config.Property)
	}

	// Convert to SKU info
	sku := skuInfo{
		Properties: make(map[string]string),
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Complex SKU object
		// Special handling for VM hardware profile which has vmSize instead of name
		if vmSize, ok := v["vmSize"].(string); ok {
			sku.Name = vmSize
		} else if name, ok := v["name"].(string); ok {
			sku.Name = name
		}

		if tier, ok := v["tier"].(string); ok {
			sku.Tier = tier
		}
		if family, ok := v["family"].(string); ok {
			sku.Family = family
		}
		if capacity, ok := v["capacity"].(float64); ok {
			sku.Capacity = int(capacity)
		}
		if size, ok := v["size"].(string); ok {
			sku.Size = size
		}

		// Store all properties, excluding those in excludeFromReport
		for key, val := range v {
			if !contains(config.ExcludeFromReport, key) {
				sku.Properties[key] = fmt.Sprintf("%v", val)
			}
		}

	case string:
		// Simple string SKU (like VM size or tier)
		sku.Name = v

	default:
		return sku, fmt.Errorf("unexpected SKU value type: %T", v)
	}

	return sku, nil
}

// navigateProperty navigates a nested property path using dot notation
func navigateProperty(obj interface{}, path string) interface{} {
	if obj == nil {
		return nil
	}

	parts := strings.Split(path, ".")
	current := obj

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// contains checks if a string slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
