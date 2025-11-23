// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"github.com/Azure/azqr/internal/azhttp"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/spf13/cobra"
)

// RegionSelectorScanner is an internal plugin that analyzes optimal Azure region selection
type RegionSelectorScanner struct {
	skuCache      *skuAvailabilityCache // Cache for SKU availability queries
	targetRegions []string              // Optional: specific regions to analyze (if empty, analyze all)
	httpClient    *azhttp.Client        // Reusable HTTP client with connection pooling and token caching
}

// NewRegionSelectorScanner creates a new region selector scanner
func NewRegionSelectorScanner() *RegionSelectorScanner {
	return &RegionSelectorScanner{
		skuCache:      newSKUAvailabilityCache(),
		targetRegions: []string{}, // Empty means analyze all regions
	}
}

// GetMetadata returns plugin metadata
func (s *RegionSelectorScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "region-selection",
		Version:     "0.1.0",
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
			{Name: "Avg Latency (ms)", DataKey: "avgLatency", FilterType: plugins.FilterTypeNone},
			{Name: "Avg Cost Difference %", DataKey: "avgCostDiff", FilterType: plugins.FilterTypeNone},
			{Name: "Recommendation Score", DataKey: "score", FilterType: plugins.FilterTypeNone},
			{Name: "Missing Resource Types", DataKey: "missingTypes", FilterType: plugins.FilterTypeSearch},
		},
	}
}

// targetRegionsFlag holds the CLI flag value at package level
var targetRegionsFlag []string

// RegisterFlags registers plugin-specific flags (implements FlagProvider interface)
func (s *RegionSelectorScanner) RegisterFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVarP(&targetRegionsFlag, "target-regions", "t", []string{}, "Target regions to analyze (comma-separated, e.g., eastus,westeurope)")
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("region-selection", NewRegionSelectorScanner())
}
