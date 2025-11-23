// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed latency.json
var latencyJSONData []byte

// azureRegionLatency contains P50 (median) round-trip time measurements in milliseconds
// between Azure regions. Data sourced from Azure Network Round-trip Latency Statistics
// published by Microsoft (updated every 6-9 months). Loaded from embedded JSON file.
var azureRegionLatency map[string]map[string]float64

// init loads the latency matrix from embedded JSON file
func init() {
	if err := json.Unmarshal(latencyJSONData, &azureRegionLatency); err != nil {
		log.Fatal().Err(err).Msg("Failed to load latency matrix from embedded JSON")
	}
	log.Debug().Msgf("Loaded latency matrix with %d source regions", len(azureRegionLatency))
}

// enrichWithLatencyData calculates weighted average network latency for each target region
// based on current resource distribution. Latency is measured from current resource locations
// to the target region, weighted by resource count in each location.
func (s *RegionSelectorScanner) enrichWithLatencyData(results []regionAvailability, inventory *resourceInventory) {
	log.Debug().Msg("Starting latency calculation using Azure published RTT statistics")

	// Calculate total resources for weighted average
	totalResources := 0
	for _, count := range inventory.locationCounts {
		totalResources += count
	}

	if totalResources == 0 {
		log.Warn().Msg("No resources found for latency calculation")
		return
	}

	// For each target region, calculate weighted average latency from all current locations
	for i := range results {
		targetRegion := normalizeRegionName(results[i].region)
		var weightedLatency float64
		var totalWeight int

		for sourceLocation, resourceCount := range inventory.locationCounts {
			sourceRegion := normalizeRegionName(sourceLocation)

			// Get latency from source to target
			latency := getRegionLatency(sourceRegion, targetRegion)

			// Weight by number of resources in this source location
			weightedLatency += latency * float64(resourceCount)
			totalWeight += resourceCount

			log.Debug().Msgf("Latency from %s to %s: %.1f ms (weight: %d resources)",
				sourceRegion, targetRegion, latency, resourceCount)
		}

		// Calculate weighted average
		if totalWeight > 0 {
			results[i].avgLatencyMs = weightedLatency / float64(totalWeight)
			log.Debug().Msgf("Target region %s: weighted average latency = %.1f ms (across %d resources)",
				targetRegion, results[i].avgLatencyMs, totalWeight)
		}
	}

	log.Debug().Msg("Latency calculation completed")
}

// getRegionLatency retrieves round-trip time in milliseconds between two Azure regions.
// Returns 0 if the latency data is not available for the region pair.
func getRegionLatency(sourceRegion, targetRegion string) float64 {
	// Same region = 0ms latency
	if sourceRegion == targetRegion {
		return 0
	}

	// Look up latency in the matrix
	if sourceLatencies, ok := azureRegionLatency[sourceRegion]; ok {
		if latency, ok := sourceLatencies[targetRegion]; ok {
			return latency
		}
	}

	// Try reverse lookup (target -> source)
	if targetLatencies, ok := azureRegionLatency[targetRegion]; ok {
		if latency, ok := targetLatencies[sourceRegion]; ok {
			return latency
		}
	}

	// No data available - use a reasonable default (100ms for unknown pairs)
	log.Debug().Msgf("No latency data available for %s -> %s, using default 100ms", sourceRegion, targetRegion)
	return 100.0
}

// normalizeRegionName converts region display names to lowercase identifiers matching the latency matrix
func normalizeRegionName(region string) string {
	// Remove spaces and convert to lowercase
	normalized := strings.ToLower(strings.ReplaceAll(region, " ", ""))

	// Handle common variations
	replacements := map[string]string{
		"useast":             "eastus",
		"useast2":            "eastus2",
		"uswest":             "westus",
		"uswest2":            "westus2",
		"uswest3":            "westus3",
		"uscentral":          "centralus",
		"usnorthcentral":     "northcentralus",
		"ussouthcentral":     "southcentralus",
		"europenorth":        "northeurope",
		"europewest":         "westeurope",
		"asiasoutheast":      "southeastasia",
		"asiaeast":           "eastasia",
		"australiasoutheast": "australiasoutheast",
		"australiaeast":      "australiaeast",
	}

	if replacement, ok := replacements[normalized]; ok {
		return replacement
	}

	return normalized
}
