// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

// resourceInventory holds the current resource inventory
type resourceInventory struct {
	resourceTypes  map[string]int            // resource type -> count
	skusByType     map[string]map[string]int // resource type -> sku -> count
	locationCounts map[string]int            // location -> count
}

// regionAvailability holds availability information for a region
type regionAvailability struct {
	subscriptionID       string
	subscriptionName     string
	region               string
	availableTypes       int
	unavailableTypes     int
	availabilityPercent  float64
	avgCostDifference    float64
	avgLatencyMs         float64 // Average network latency in milliseconds
	missingResourceTypes []string
	score                float64
}

// providerCache caches all provider information for fast lookup
type providerCache struct {
	// namespace -> resourceType -> []locations
	data map[string]map[string][]string
}

// meterCostData holds cost information for a specific meter from Cost Management API
// This follows Get-CostInformation.ps1 approach which queries Cost Management API
type meterCostData struct {
	meterID          string
	meterName        string
	productID        string
	skuName          string
	armRegionName    string
	historicalCost   float64
	unitOfMeasure    string
	tierMinimumUnits float64
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
