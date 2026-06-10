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
	unknownPairs := 0
	for i := range results {
		sourceRegion := normalizeRegionName(results[i].sourceRegion)
		targetRegion := normalizeRegionName(results[i].targetRegion)

		// Get latency from source to target
		latency := getRegionLatency(sourceRegion, targetRegion)
		results[i].avgLatencyMs = latency

		if latency == 0 && sourceRegion != targetRegion {
			unknownPairs++
		}

		log.Debug().Msgf("Latency from %s to %s: %.1f ms",
			sourceRegion, targetRegion, latency)
	}

	if unknownPairs > 0 {
		log.Warn().Msgf("Latency data unavailable for %d of %d region pair(s) — shown as 'N/A' in table, scored as neutral",
			unknownPairs, len(results))
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

	// No data available — return 0 so the caller can show "N/A" and score as neutral.
	log.Debug().Msgf("No latency data for %s -> %s — pair scored as neutral", sourceRegion, targetRegion)
	return 0
}

// normalizeRegionName converts region display names to lowercase identifiers
// by removing spaces and converting to lowercase. Use this everywhere a region
// name needs to be compared or stored.
func normalizeRegionName(region string) string {
	return strings.ToLower(strings.ReplaceAll(region, " ", ""))
}
