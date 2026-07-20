// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/spf13/cobra"
)

// RegionSelectorScanner is an internal plugin that analyzes optimal Azure region selection
type RegionSelectorScanner struct {
	skuCache          *types.SKUAvailabilityCache // Cache for SKU availability queries
	targetRegions     []string                    // Optional: specific regions to analyze (if empty, analyze all)
	httpClient        *az.HttpClient              // Reusable HTTP client with connection pooling and token caching
	cred              azcore.TokenCredential      // Azure credential for typed ARM SDK clients
	clientOpts        *arm.ClientOptions          // ARM client options shared by all typed SDK clients
	costHistoryMonths int                         // Number of full calendar months to include in Cost Management query (default: 1)
}

// NewScanner creates a new region selector scanner
func NewScanner() *RegionSelectorScanner {
	return &RegionSelectorScanner{
		skuCache:          types.NewSKUAvailabilityCache(),
		targetRegions:     []string{}, // Empty means analyze all regions
		costHistoryMonths: 1,
	}
}

// GetMetadata returns plugin metadata
func (s *RegionSelectorScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "region-selection",
		Version:     "0.2.0-beta",
		Description: "Analyzes optimal Azure region selection based on service availability, network latency, and cost comparison",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription", DataKey: "subscription", FilterType: plugins.FilterTypeDropdown},
			{Name: "Source Region", DataKey: "sourceRegion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Target Region", DataKey: "targetRegion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Source Resource Type Count", DataKey: "sourceResourceTypeCount", FilterType: plugins.FilterTypeNone},
			{Name: "Available Resource Types", DataKey: "availableResources", FilterType: plugins.FilterTypeNone},
			{Name: "Unavailable Resource Types", DataKey: "unavailableResources", FilterType: plugins.FilterTypeNone},
			{Name: "Availability %", DataKey: "availabilityPercent", FilterType: plugins.FilterTypeNone},
			{Name: "Total SKUs Checked", DataKey: "totalSKUsChecked", FilterType: plugins.FilterTypeNone},
			{Name: "Available SKUs", DataKey: "availableSKUs", FilterType: plugins.FilterTypeNone},
			{Name: "Unavailable SKUs", DataKey: "unavailableSKUs", FilterType: plugins.FilterTypeNone},
			{Name: "Restricted SKUs", DataKey: "restrictedSKUs", FilterType: plugins.FilterTypeNone},
			{Name: "Zone-Restricted SKUs", DataKey: "zoneRestrictedSKUs", FilterType: plugins.FilterTypeNone},
			{Name: "Unknown SKUs", DataKey: "unknownSKUs", FilterType: plugins.FilterTypeNone},
			{Name: "SKU Availability %", DataKey: "skuAvailabilityPercent", FilterType: plugins.FilterTypeNone},
			{Name: "Availability Zones", DataKey: "availabilityZones", FilterType: plugins.FilterTypeNone},
			{Name: "Target AZ Mapping", DataKey: "targetAZMapping", FilterType: plugins.FilterTypeNone},
			{Name: "Avg Latency (ms)", DataKey: "avgLatency", FilterType: plugins.FilterTypeNone},
			{Name: "Avg Cost Difference %", DataKey: "avgCostDiff", FilterType: plugins.FilterTypeNone},
			{Name: "Recommendation Score", DataKey: "score", FilterType: plugins.FilterTypeNone},
			{Name: "Score Quality", DataKey: "scoreQuality", FilterType: plugins.FilterTypeDropdown},
			{Name: "Recommendation", DataKey: "recommendation", FilterType: plugins.FilterTypeDropdown},
			{Name: "Missing Resource Types", DataKey: "missingTypes", FilterType: plugins.FilterTypeSearch},
			{Name: "Unavailable SKUs (detail)", DataKey: "unavailableSKUsDetail", FilterType: plugins.FilterTypeSearch},
			{Name: "Restricted SKUs (detail)", DataKey: "restrictedSKUsDetail", FilterType: plugins.FilterTypeSearch},
			{Name: "Zone-Restricted SKUs (detail)", DataKey: "zoneRestrictedSKUsDetail", FilterType: plugins.FilterTypeSearch},
		},
	}
}

// RegisterFlags registers plugin-specific flags (implements FlagProvider interface)
func (s *RegionSelectorScanner) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("target-regions", []string{}, "Target regions to analyze (comma-separated, e.g., eastus,westeurope)")
	cmd.Flags().Int("cost-history-months", 1, "Number of full calendar months of Cost Management history to use for pricing weights (1–12, default: 1)")
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("region-selection", NewScanner())
}
