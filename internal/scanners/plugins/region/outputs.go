// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azqr/internal/models"
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
	inventory *resourceInventory,
	comparisons []regionComparison,
	costDetails map[string]interface{},
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
	if opts.GenerateCost && costDetails != nil {
		if err := writeCostAnalysisJSONs(opts.OutputDir, costDetails); err != nil {
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
func writeSummaryJSON(outputDir string, inventory *resourceInventory) error {
	// Build summary structure
	summary := make(map[string]interface{})

	// Add resource type summary
	resourceTypeSummary := []map[string]interface{}{}
	for resourceType, count := range inventory.resourceTypes {
		item := map[string]interface{}{
			"ResourceType":  resourceType,
			"ResourceCount": count,
		}

		// Add SKUs for this resource type
		if skus, exists := inventory.skusByType[resourceType]; exists {
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
	for location, count := range inventory.locationCounts {
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
func writeAvailabilityMappingJSON(outputDir string, comparisons []regionComparison) error {
	// Build availability mapping structure
	mapping := make(map[string][]map[string]interface{})

	// Group by source region
	for _, comp := range comparisons {
		sourceRegion := comp.sourceRegion
		if _, exists := mapping[sourceRegion]; !exists {
			mapping[sourceRegion] = []map[string]interface{}{}
		}

		entry := map[string]interface{}{
			"targetRegion":             comp.targetRegion,
			"availableResourceTypes":   comp.availableTypes,
			"unavailableResourceTypes": comp.unavailableTypes,
			"availabilityPercent":      comp.availabilityPercent,
			"missingResourceTypes":     comp.missingResourceTypes,
			"totalSKUsChecked":         comp.totalSKUsChecked,
			"availableSKUs":            comp.availableSKUs,
			"unavailableSKUs":          comp.unavailableSKUs,
			"skuAvailabilityPercent":   comp.skuAvailabilityPercent,
			"missingSKUs":              comp.missingSKUs,
			"avgLatencyMs":             comp.avgLatencyMs,
			"avgCostDifference":        comp.avgCostDifference,
			"recommendationScore":      comp.score,
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
func writeCostAnalysisJSONs(outputDir string, costDetails map[string]interface{}) error {
	// Write region_comparison_inputs.json (meter metadata)
	if inputs, ok := costDetails["inputs"]; ok {
		data, err := json.MarshalIndent(inputs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal cost inputs: %w", err)
		}
		filePath := filepath.Join(outputDir, "region_comparison_inputs.json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write cost inputs file: %w", err)
		}
	}

	// Write region_comparison_prices.json (full pricing matrix)
	if prices, ok := costDetails["prices"]; ok {
		data, err := json.MarshalIndent(prices, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal cost prices: %w", err)
		}
		filePath := filepath.Join(outputDir, "region_comparison_prices.json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write cost prices file: %w", err)
		}
	}

	// Write region_comparison_pricemap.json (summary by meter)
	if pricemap, ok := costDetails["pricemap"]; ok {
		data, err := json.MarshalIndent(pricemap, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal cost pricemap: %w", err)
		}
		filePath := filepath.Join(outputDir, "region_comparison_pricemap.json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write cost pricemap file: %w", err)
		}
	}

	// Write region_comparison_uomerrors.json (unit of measure errors)
	if uomErrors, ok := costDetails["uomErrors"]; ok {
		data, err := json.MarshalIndent(uomErrors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal UOM errors: %w", err)
		}
		filePath := filepath.Join(outputDir, "region_comparison_uomerrors.json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write UOM errors file: %w", err)
		}
	}

	return nil
}
