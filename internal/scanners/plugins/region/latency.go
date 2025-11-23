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
}

// enrichWithLatencyData calculates network latency from source region to target region
// for each result. Uses Azure published RTT statistics.
func (s *RegionSelectorScanner) enrichWithLatencyData(results []regionComparison) {
	log.Debug().Msg("Starting latency calculation using Azure published RTT statistics")

	// For each source->target pair, get the direct latency
	for i := range results {
		sourceRegion := normalizeRegionName(results[i].sourceRegion)
		targetRegion := normalizeRegionName(results[i].targetRegion)

		// Get latency from source to target
		latency := getRegionLatency(sourceRegion, targetRegion)
		results[i].avgLatencyMs = latency

		log.Debug().Msgf("Latency from %s to %s: %.1f ms",
			sourceRegion, targetRegion, latency)
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
	return normalized
}
