// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners/plugins/region/cost"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/rs/zerolog/log"
)

// outputOptions defines options for JSON output generation
type outputOptions struct {
	OutputDir         string
	GenerateResources bool
	GenerateSummary   bool
	GenerateMapping   bool
	GenerateCost      bool
}

// generateJSONOutputs generates intermediate JSON files for debugging and analysis
func generateJSONOutputs(
	opts outputOptions,
	resources []*models.Resource,
	inventory *types.ResourceInventory,
	comparisons []types.RegionComparison,
	costData *types.CostComparisonData,
) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate resources.json
	if opts.GenerateResources {
		if err := writeResourcesJSON(opts.OutputDir, resources); err != nil {
			log.Error().Err(err).Msg("Failed to write resources.json")
		} else {
			log.Info().Msgf("Generated %s/resources.json", opts.OutputDir)
		}
	}

	// Generate summary.json
	if opts.GenerateSummary {
		if err := writeSummaryJSON(opts.OutputDir, inventory); err != nil {
			log.Error().Err(err).Msg("Failed to write summary.json")
		} else {
			log.Info().Msgf("Generated %s/summary.json", opts.OutputDir)
		}
	}

	// Generate Availability_Mapping.json
	if opts.GenerateMapping {
		if err := writeAvailabilityMappingJSON(opts.OutputDir, comparisons); err != nil {
			log.Error().Err(err).Msg("Failed to write Availability_Mapping.json")
		} else {
			log.Info().Msgf("Generated %s/Availability_Mapping.json", opts.OutputDir)
		}
	}

	// Generate cost analysis JSONs
	if opts.GenerateCost && costData != nil {
		if err := writeCostAnalysisJSONs(opts.OutputDir, costData); err != nil {
			log.Error().Err(err).Msg("Failed to write cost analysis JSONs")
		} else {
			log.Info().Msgf("Generated cost analysis JSONs in %s/", opts.OutputDir)
		}
	}

	return nil
}

// writeResourcesJSON writes the resources inventory to resources.json
func writeResourcesJSON(outputDir string, resources []*models.Resource) error {
	data, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	filePath := filepath.Join(outputDir, "resources.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// writeSummaryJSON writes the inventory summary to summary.json
func writeSummaryJSON(outputDir string, inventory *types.ResourceInventory) error {
	// Build summary structure
	summary := make(map[string]interface{})

	// Add resource type summary
	resourceTypeSummary := []map[string]interface{}{}
	for resourceType, count := range inventory.ResourceTypes {
		item := map[string]interface{}{
			"ResourceType":  resourceType,
			"ResourceCount": count,
		}

		// Add SKUs for this resource type
		if skus, exists := inventory.SKUsByType[resourceType]; exists {
			implementedSKUs := []map[string]interface{}{}
			for skuName, skuCount := range skus {
				implementedSKUs = append(implementedSKUs, map[string]interface{}{
					"sku":   skuName,
					"count": skuCount,
				})
			}
			item["ImplementedSkus"] = implementedSKUs
		}

		resourceTypeSummary = append(resourceTypeSummary, item)
	}
	summary["ResourceTypes"] = resourceTypeSummary

	// Add location summary
	locationSummary := []map[string]interface{}{}
	for location, count := range inventory.LocationCounts {
		locationSummary = append(locationSummary, map[string]interface{}{
			"location":      location,
			"resourceCount": count,
		})
	}
	summary["ImplementedRegions"] = locationSummary

	// Write to file
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	filePath := filepath.Join(outputDir, "summary.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// writeAvailabilityMappingJSON writes the availability mapping to Availability_Mapping.json
func writeAvailabilityMappingJSON(outputDir string, comparisons []types.RegionComparison) error {
	// Build availability mapping structure
	mapping := make(map[string][]map[string]interface{})

	// Group by source region
	for _, comp := range comparisons {
		sourceRegion := comp.SourceRegion
		if _, exists := mapping[sourceRegion]; !exists {
			mapping[sourceRegion] = []map[string]interface{}{}
		}

		entry := map[string]interface{}{
			"targetRegion":             comp.TargetRegion,
			"availableResourceTypes":   comp.AvailableTypes,
			"unavailableResourceTypes": comp.UnavailableTypes,
			"availabilityPercent":      comp.AvailabilityPercent,
			"missingResourceTypes":     comp.MissingResourceTypes,
			"totalSKUsChecked":         comp.TotalSKUsChecked,
			"availableSKUs":            comp.AvailableSKUs,
			"unavailableSKUs":          comp.UnavailableSKUs,
			"restrictedSKUs":           comp.RestrictedSKUs,
			"zoneRestrictedSKUs":       comp.ZoneRestrictedSKUs,
			"unknownSKUs":              comp.UnknownSKUs,
			"skuAvailabilityPercent":   comp.SKUAvailabilityPercent,
			"missingSKUs":              comp.MissingSKUs,
			"sourceZoneCount":          comp.SourceZoneCount,
			"targetZoneCount":          comp.TargetZoneCount,
			"targetZoneMappings":       comp.TargetZoneMappings,
			"avgLatencyMs":             comp.AvgLatencyMs,
			"avgCostDifference":        comp.AvgCostDifference,
			"recommendationScore":      comp.Score,
		}

		mapping[sourceRegion] = append(mapping[sourceRegion], entry)
	}

	// Write to file
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal availability mapping: %w", err)
	}

	filePath := filepath.Join(outputDir, "Availability_Mapping.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// writeCostAnalysisJSONs writes cost analysis details to multiple JSON files
func writeCostAnalysisJSONs(outputDir string, costData *types.CostComparisonData) error {
	// Reconstruct legacy map format for JSON serialisation
	costDetails := cost.BuildCostDetailsForOutput(costData.MeterInputs, costData.RegionPricing, costData.PriceItems, costData.UomErrors)

	files := []struct {
		key      string
		filename string
		label    string
	}{
		{"inputs", "region_comparison_inputs.json", "cost inputs"},
		{"prices", "region_comparison_prices.json", "cost prices"},
		{"pricemap", "region_comparison_pricemap.json", "cost pricemap"},
		{"uomErrors", "region_comparison_uomerrors.json", "UOM errors"},
	}

	for _, f := range files {
		value, ok := costDetails[f.key]
		if !ok {
			continue
		}
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %w", f.label, err)
		}
		if err := os.WriteFile(filepath.Join(outputDir, f.filename), data, 0644); err != nil {
			return fmt.Errorf("failed to write %s file: %w", f.label, err)
		}
	}

	return nil
}
