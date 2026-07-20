// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package types

import (
	"strings"
)

// scoringWeights defines the weights for the scoring algorithm
type scoringWeights struct {
	ResourceAvailability float64 // Weight for resource type availability (default: 0.35)
	SKUAvailability      float64 // Weight for SKU-level availability (default: 0.30)
	Cost                 float64 // Weight for cost difference (default: 0.15)
	Latency              float64 // Weight for network latency (default: 0.20)
}

// DefaultScoringWeights returns the default scoring weights
func DefaultScoringWeights() scoringWeights {
	return scoringWeights{
		ResourceAvailability: 0.35,
		SKUAvailability:      0.30,
		Cost:                 0.15,
		Latency:              0.20,
	}
}

// SKUAvailabilityState is the availability state for an individual SKU check.
type SKUAvailabilityState int

const (
	SKUAvailable      SKUAvailabilityState = iota // SKU confirmed available in target region
	SKURestricted                                 // SKU regionally blocked for this subscription (NotAvailableForSubscription — quota-liftable)
	SKUUnavailable                                // SKU not available in target region (hard block)
	SKUZoneRestricted                             // SKU available in region but restricted in one or more zones (zone-liftable)
)

// SKUAvailability holds the availability state and optional zone-restriction details for one SKU.
type SKUAvailability struct {
	State        SKUAvailabilityState
	BlockedZones []string // logical zones blocked; only populated when State == SKUZoneRestricted
}

// ResourceInventory holds the current resource inventory
type ResourceInventory struct {
	ResourceTypes         map[string]int                       // resource type -> count
	SKUsByType            map[string]map[string]int            // resource type -> sku -> count (used for SvcAvail sheets)
	LocationCounts        map[string]int                       // location -> count
	ResourceTypesByRegion map[string]map[string]int            // sourceRegion -> resourceType -> count
	SKUsByTypeAndRegion   map[string]map[string]map[string]int // resourceType -> region -> sku -> count
}

// RegionComparison holds availability information for a region
type RegionComparison struct {
	SubscriptionID          string
	SubscriptionName        string
	SourceRegion            string // Source region we're comparing FROM
	TargetRegion            string // Target region we're comparing TO
	SourceResourceTypeCount int    // Number of unique resource types in source region
	AvailableTypes          int
	UnavailableTypes        int
	AvailabilityPercent     float64
	AvgCostDifference       float64
	HasCostData             bool    // True when at least one meter was successfully priced in both regions
	AvgLatencyMs            float64 // Average network latency in milliseconds
	LatencyEstimated        bool    // True when AvgLatencyMs is a cluster-based estimate, not a direct measurement
	MissingResourceTypes    []string
	MissingSKUs             []string // Specific SKUs not available in target region (hard block)
	RestrictedSKUs          []string // SKUs regionally restricted for this subscription (NotAvailableForSubscription — quota-liftable)
	ZoneRestrictedSKUs      []string // SKUs available in region but restricted in one or more zones (zone-liftable)
	TotalSKUsChecked        int      // Total number of SKUs checked
	AvailableSKUs           int      // Number of SKUs confirmed available
	UnavailableSKUs         int      // Number of SKUs not available (hard block)
	UnknownSKUs             int      // Number of SKUs where availability could not be determined (API error)
	SKUAvailabilityPercent  float64  // Percentage of SKUs available (raw; excludes unknowns from denominator)
	SourceZoneCount         int               // Number of Availability Zones in source region (0 = no AZ support)
	TargetZoneCount         int               // Number of Availability Zones in target region (0 = no AZ support)
	TargetZoneMappings      map[string]string // Logical → physical AZ mapping for target region (subscription-scoped); nil when unavailable
	Score                   float64
}

// ResourceTypeLocationData caches all provider information for fast lookup.
// The innermost map is a set (struct{} value) for O(1) membership tests.
type ResourceTypeLocationData struct {
	// namespace -> resourceType -> set of locations
	Data map[string]map[string]map[string]struct{}
}

// IsAvailable checks if a resource type is available in a region in O(1).
// Both resourceType and region must already be lowercase (callers normalise them
// before building the inventory, so no ToLower is needed here).
func (rtl *ResourceTypeLocationData) IsAvailable(resourceType, region string) bool {
	// Resource type format: "microsoft.compute/virtualmachines" (already lowercase from inventory).
	namespace, typeName, ok := strings.Cut(resourceType, "/")
	if !ok {
		return false
	}

	locs, exists := rtl.Data[namespace][typeName]
	if !exists {
		return false
	}

	_, found := locs[region]
	return found
}

// SKUAPIResponse is the envelope returned by Azure SKU availability REST APIs
// (e.g. /providers/Microsoft.Compute/skus, /skus for Storage, etc.).
type SKUAPIResponse struct {
	Value []SKUAPIItem `json:"value"`
}

// SKUAPIItem represents one SKU entry in an Azure SKU availability API response.
// Fields are the union of what Microsoft.Compute/skus, Microsoft.Storage/skus,
// and similar endpoints return.
type SKUAPIItem struct {
	Name         string            `json:"name"`
	Tier         string            `json:"tier"`
	Size         string            `json:"size"`
	Locations    []string          `json:"locations"`
	LocationInfo []SKULocationInfo `json:"locationInfo"`
	Restrictions []SKURestriction  `json:"restrictions"`
	Capabilities []SKUCapability   `json:"capabilities"`
}

// SKULocationInfo holds the per-physical-location data returned inside
// SKUAPIItem for Microsoft.Compute SKU responses.
type SKULocationInfo struct {
	Location string   `json:"location"`
	Zones    []string `json:"zones"`
}

// SKURestrictionInfo holds location and zone context for a SKU restriction.
type SKURestrictionInfo struct {
	Locations []string `json:"locations"`
	Zones     []string `json:"zones"`
}

// SKURestriction describes a restriction that may prevent a SKU from being
// used in a region or zone.
type SKURestriction struct {
	Type            string             `json:"type"`
	ReasonCode      string             `json:"reasonCode"`
	RestrictionInfo SKURestrictionInfo `json:"restrictionInfo"`
}

// SKUCapability is a named capability flag reported by some SKU APIs
// (e.g. "available": "false").
type SKUCapability struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MeterCostData holds cost information for a specific meter from Cost Management API
// This follows Get-CostInformation.ps1 approach which queries Cost Management API
type MeterCostData struct {
	MeterID          string
	MeterName        string
	ProductID        string
	ProductName      string // from Retail Prices API — used for CostComparison sheet
	ServiceName      string // from Retail Prices API — used for CostComparison sheet
	SkuName          string
	ArmSkuName       string // armSkuName from Retail Prices API — matches embedded pricing table keys
	ArmRegionName    string
	UnitOfMeasure    string
	TierMinimumUnits float64
	HistoricalCost   float64
}

// UoMError tracks unit of measure mismatches between source and target regions (like PowerShell)
type UoMError struct {
	OrigMeterID   string `json:"origMeterID"`
	OrigUoM       string `json:"origUoM"`
	TargetMeterID string `json:"targetMeterID"`
	TargetUoM     string `json:"targetUoM"`
}

// RetailPriceItem represents an item from Azure Retail Prices API
type RetailPriceItem struct {
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

// RetailPriceResponse represents the response from Azure Retail Prices API
type RetailPriceResponse struct {
	BillingCurrency    string            `json:"BillingCurrency"`
	CustomerEntityID   string            `json:"CustomerEntityId"`
	CustomerEntityType string            `json:"CustomerEntityType"`
	Items              []RetailPriceItem `json:"Items"`
	NextPageLink       string            `json:"NextPageLink"`
	Count              int               `json:"Count"`
}

// CostComparisonData holds structured cost data for Excel output and JSON debug files
type CostComparisonData struct {
	MeterInputs   []MeterCostData               // Meter metadata from Cost Management + Retail API
	RegionPricing map[string]map[string]float64 // meterID → region → retail price
	PriceItems    []RetailPriceItem             // Full retail price items (for JSON debug output)
	UomErrors     []UoMError                    // Unit-of-measure mismatches excluded from comparison
}

// NormalizeRegionName converts region display names to lowercase identifiers
// by removing spaces and converting to lowercase. Use this everywhere a region
// name needs to be compared or stored.
func NormalizeRegionName(region string) string {
	return strings.ToLower(strings.ReplaceAll(region, " ", ""))
}

// nonPhysicalRegions contains Azure meta/logical region identifiers that do not
// correspond to a deployable physical region. Quota and cost APIs are meaningless for these.
var nonPhysicalRegions = map[string]bool{
	"":             true,
	"unassigned":   true,
	"global":       true,
	"europe":       true,
	"unitedstates": true,
	"asia":         true,
	"asiapacific":  true,
	"australia":    true,
	"brazil":       true,
	"canada":       true,
	"france":       true,
	"germany":      true,
	"india":        true,
	"japan":        true,
	"korea":        true,
	"norway":       true,
	"southafrica":  true,
	"switzerland":  true,
	"uae":          true,
	"uk":           true,
}

// IsPhysicalRegion returns true when region is a real deployable Azure region
// (not a meta/logical identifier like "global", "europe", etc.).
func IsPhysicalRegion(region string) bool {
	return !nonPhysicalRegions[NormalizeRegionName(region)]
}
