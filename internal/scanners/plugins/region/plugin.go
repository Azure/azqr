// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners/plugins/region/sku"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/spf13/cobra"
)

// RegionSelectorScanner is an internal plugin that analyzes optimal Azure region selection
type RegionSelectorScanner struct {
	skuCache          *sku.Cache             // Cache for SKU availability queries
	targetRegions     []string               // Optional: specific regions to analyze (if empty, analyze all)
	httpClient        *az.HttpClient         // Reusable HTTP client with connection pooling and token caching
	cred              azcore.TokenCredential // Azure credential for typed ARM SDK clients
	clientOpts        *arm.ClientOptions     // ARM client options shared by all typed SDK clients
	costHistoryMonths int                    // Number of full calendar months to include in Cost Management query (default: 1)
}

// NewScanner creates a new region selector scanner
func NewScanner() *RegionSelectorScanner {
	return &RegionSelectorScanner{
		skuCache:          sku.NewCache(),
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
			{Name: "Subscription"},
			{Name: "Source Region"},
			{Name: "Target Region"},
			{Name: "Source Resource Type Count"},
			{Name: "Available Resource Types"},
			{Name: "Unavailable Resource Types"},
			{Name: "Availability %"},
			{Name: "Total SKUs Checked"},
			{Name: "Available SKUs"},
			{Name: "Unavailable SKUs"},
			{Name: "Restricted SKUs"},
			{Name: "Zone-Restricted SKUs"},
			{Name: "Unknown SKUs"},
			{Name: "SKU Availability %"},
			{Name: "Availability Zones"},
			{Name: "Target AZ Mapping"},
			{Name: "Avg Latency (ms)"},
			{Name: "Avg Cost Difference %"},
			{Name: "Recommendation Score"},
			{Name: "Score Quality"},
			{Name: "Recommendation"},
			{Name: "Missing Resource Types"},
			{Name: "Unavailable SKUs (detail)"},
			{Name: "Restricted SKUs (detail)"},
			{Name: "Zone-Restricted SKUs (detail)"},
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
