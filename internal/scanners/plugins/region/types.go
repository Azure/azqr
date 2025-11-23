// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import "strings"

// skuInfo holds detailed SKU information for a resource
type skuInfo struct {
	Name     string `json:"name"`
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

// skuAvailabilityState is a tri-state result for an individual SKU check.
type skuAvailabilityState int

const (
	skuAvailable    skuAvailabilityState = iota // SKU confirmed available in target region
	skuRestricted                               // SKU exists but restricted for this subscription (NotAvailableForSubscription — quota-liftable)
	skuUnavailable                              // SKU not available in target region (hard block)
)

// resourceInventory holds the current resource inventory
type resourceInventory struct {
	resourceTypes         map[string]int                       // resource type -> count
	skusByType            map[string]map[string]int            // resource type -> sku -> count (used for SvcAvail sheets)
	locationCounts        map[string]int                       // location -> count
	resourceTypesByRegion map[string]map[string]int            // sourceRegion -> resourceType -> count
	skusByTypeAndRegion   map[string]map[string]map[string]int // resourceType -> region -> sku -> count
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
	missingSKUs             []string // Specific SKUs not available in target region (hard block)
	restrictedSKUs          []string // SKUs restricted for this subscription (NotAvailableForSubscription — quota-liftable)
	totalSKUsChecked        int      // Total number of SKUs checked
	availableSKUs           int      // Number of SKUs confirmed available
	unavailableSKUs         int      // Number of SKUs not available (hard block)
	unknownSKUs             int      // Number of SKUs where availability could not be determined (API error)
	skuAvailabilityPercent  float64  // Percentage of SKUs available (raw; excludes unknowns from denominator)
	sourceZoneCount         int      // Number of Availability Zones in source region (0 = no AZ support)
	targetZoneCount         int      // Number of Availability Zones in target region (0 = no AZ support)
	score                   float64
}

// resourceTypeLocationData caches all provider information for fast lookup.
// The innermost map is a set (struct{} value) for O(1) membership tests.
type resourceTypeLocationData struct {
	// namespace -> resourceType -> set of locations
	data map[string]map[string]map[string]struct{}
}

// isAvailable checks if a resource type is available in a region in O(1).
func (rtl *resourceTypeLocationData) isAvailable(resourceType, region string) bool {
	// Parse resource type (format: Microsoft.Compute/virtualMachines)
	parts := strings.SplitN(resourceType, "/", 2)
	if len(parts) != 2 {
		return false
	}

	namespace := strings.ToLower(parts[0])
	typeName := strings.ToLower(parts[1])

	locs, exists := rtl.data[namespace][typeName]
	if !exists {
		return false
	}

	_, found := locs[region]
	return found
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

// CostComparisonData holds structured cost data for Excel output and JSON debug files
type CostComparisonData struct {
	MeterInputs   []meterCostData               // Meter metadata from Cost Management + Retail API
	RegionPricing map[string]map[string]float64 // meterID → region → retail price
	PriceItems    []retailPriceItem             // Full retail price items (for JSON debug output)
	UomErrors     []uomError                    // Unit-of-measure mismatches excluded from comparison
}
