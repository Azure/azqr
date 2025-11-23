// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import "strings"

// skuInfo holds detailed SKU information for a resource
type skuInfo struct {
	Name       string            `json:"name"`
	Tier       string            `json:"tier,omitempty"`
	Family     string            `json:"family,omitempty"`
	Capacity   int               `json:"capacity,omitempty"`
	Size       string            `json:"size,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

// resourceWithSKU tracks a resource and its SKU details
type resourceWithSKU struct {
	ResourceID   string
	ResourceType string
	Location     string
	SKU          skuInfo
	MeterIDs     []string // Associated meter IDs for cost tracking
}

// scoringWeights defines the configurable weights for the scoring algorithm
type scoringWeights struct {
	ResourceAvailability float64 // Weight for resource type availability (default: 0.35)
	SKUAvailability      float64 // Weight for SKU-level availability (default: 0.30)
	Cost                 float64 // Weight for cost difference (default: 0.15)
	Latency              float64 // Weight for network latency (default: 0.20)
}

// defaultScoringWeights returns the default scoring weights
func defaultScoringWeights() scoringWeights {
	return scoringWeights{
		ResourceAvailability: 0.35,
		SKUAvailability:      0.30,
		Cost:                 0.15,
		Latency:              0.20,
	}
}

// resourceInventory holds the current resource inventory
type resourceInventory struct {
	resourceTypes         map[string]int                       // resource type -> count
	skusByType            map[string]map[string]int            // resource type -> sku -> count
	locationCounts        map[string]int                       // location -> count
	resourceTypesByRegion map[string]map[string]int            // sourceRegion -> resourceType -> count
	skusByTypeAndRegion   map[string]map[string]map[string]int // resourceType -> region -> sku -> count
	resourcesWithSKUs     []resourceWithSKU                    // Detailed resource list with SKUs and meter IDs
}

// regionComparison holds availability information for a region
type regionComparison struct {
	subscriptionID          string
	subscriptionName        string
	sourceRegion            string // Source region we're comparing FROM
	targetRegion            string // Target region we're comparing TO
	sourceResourceTypeCount int    // Number of unique resource types in source region
	availableTypes          int
	unavailableTypes        int
	availabilityPercent     float64
	avgCostDifference       float64
	avgLatencyMs            float64 // Average network latency in milliseconds
	missingResourceTypes    []string
	missingSKUs             []string // Specific SKUs not available in target region
	totalSKUsChecked        int      // Total number of SKUs checked
	availableSKUs           int      // Number of SKUs available
	unavailableSKUs         int      // Number of SKUs not available
	skuAvailabilityPercent  float64  // Percentage of SKUs available
	score                   float64
}

// resourceTypeLocationData caches all provider information for fast lookup
type resourceTypeLocationData struct {
	// namespace -> resourceType -> []locations
	data map[string]map[string][]string
}

// isAvailable checks if a resource type is available in a region
func (rtl *resourceTypeLocationData) isAvailable(resourceType, region string) bool {
	// Parse resource type (format: Microsoft.Compute/virtualMachines)
	parts := strings.SplitN(resourceType, "/", 2)
	if len(parts) != 2 {
		return false
	}

	namespace := strings.ToLower(parts[0])
	typeName := strings.ToLower(parts[1])

	// Look up in cache
	if rtl.data[namespace] == nil {
		return false
	}

	locations, exists := rtl.data[namespace][typeName]
	if !exists {
		return false
	}

	// Check if region is in the locations list
	for _, loc := range locations {
		if loc == region {
			return true
		}
	}

	return false
}

// meterCostData holds cost information for a specific meter from Cost Management API
// This follows Get-CostInformation.ps1 approach which queries Cost Management API
type meterCostData struct {
	meterID          string
	meterName        string
	productID        string
	skuName          string
	armRegionName    string
	unitOfMeasure    string
	tierMinimumUnits float64
	historicalCost   float64
}

// uomError tracks unit of measure mismatches between source and target regions (like PowerShell)
type uomError struct {
	OrigMeterID   string `json:"origMeterID"`
	OrigUoM       string `json:"origUoM"`
	TargetMeterID string `json:"targetMeterID"`
	TargetUoM     string `json:"targetUoM"`
}

// retailPriceItem represents an item from Azure Retail Prices API
type retailPriceItem struct {
	CurrencyCode         string  `json:"currencyCode"`
	TierMinimumUnits     float64 `json:"tierMinimumUnits"`
	RetailPrice          float64 `json:"retailPrice"`
	UnitPrice            float64 `json:"unitPrice"`
	ArmRegionName        string  `json:"armRegionName"`
	Location             string  `json:"location"`
	EffectiveStartDate   string  `json:"effectiveStartDate"`
	MeterID              string  `json:"meterId"`
	MeterName            string  `json:"meterName"`
	ProductID            string  `json:"productId"`
	SkuID                string  `json:"skuId"`
	ProductName          string  `json:"productName"`
	SkuName              string  `json:"skuName"`
	ServiceName          string  `json:"serviceName"`
	ServiceID            string  `json:"serviceId"`
	ServiceFamily        string  `json:"serviceFamily"`
	UnitOfMeasure        string  `json:"unitOfMeasure"`
	Type                 string  `json:"type"`
	IsPrimaryMeterRegion bool    `json:"isPrimaryMeterRegion"`
	ArmSkuName           string  `json:"armSkuName"`
}

// retailPriceResponse represents the response from Azure Retail Prices API
type retailPriceResponse struct {
	BillingCurrency    string            `json:"BillingCurrency"`
	CustomerEntityID   string            `json:"CustomerEntityId"`
	CustomerEntityType string            `json:"CustomerEntityType"`
	Items              []retailPriceItem `json:"Items"`
	NextPageLink       string            `json:"NextPageLink"`
	Count              int               `json:"Count"`
}
