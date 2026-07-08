// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package latency

import (
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/rs/zerolog/log"
)

// clusterPairAverages holds the average RTT (ms) between geographic cluster pairs,
// computed at init time from the measured latency matrix.
// Key format: "srcCluster:tgtCluster" (e.g. "americas:europe").
var clusterPairAverages map[string]float64

func init() {
	clusterPairAverages = computeClusterAverages(azureRegionLatency)
}

// computeClusterAverages builds a cluster→cluster average RTT table from a measured latency matrix.
// Only pairs where both regions appear in regionCluster contribute to the averages.
func computeClusterAverages(matrix map[string]map[string]float64) map[string]float64 {
	type bucket struct {
		sum   float64
		count int
	}
	buckets := make(map[string]*bucket)

	for src, targets := range matrix {
		srcCluster, ok := regionCluster[src]
		if !ok {
			continue
		}
		for tgt, ms := range targets {
			tgtCluster, ok := regionCluster[tgt]
			if !ok {
				continue
			}
			key := srcCluster + ":" + tgtCluster
			if buckets[key] == nil {
				buckets[key] = &bucket{}
			}
			buckets[key].sum += ms
			buckets[key].count++
		}
	}

	averages := make(map[string]float64, len(buckets))
	for key, b := range buckets {
		averages[key] = b.sum / float64(b.count)
	}
	return averages
}

// EnrichWithLatencyData populates avgLatencyMs and latencyEstimated on each result.
func EnrichWithLatencyData(results []types.RegionComparison) {
	log.Debug().Msg("Starting latency calculation using Azure published RTT statistics")

	var unknownPairs, estimatedPairs int
	for i := range results {
		sourceRegion := types.NormalizeRegionName(results[i].SourceRegion)
		targetRegion := types.NormalizeRegionName(results[i].TargetRegion)

		latency, estimated := getRegionLatency(sourceRegion, targetRegion)
		results[i].AvgLatencyMs = latency
		results[i].LatencyEstimated = estimated

		switch {
		case latency == 0 && sourceRegion != targetRegion:
			unknownPairs++
		case estimated:
			estimatedPairs++
		}

		log.Debug().Msgf("Latency from %s to %s: %.1f ms (estimated=%v)",
			sourceRegion, targetRegion, latency, estimated)
	}

	if estimatedPairs > 0 {
		log.Warn().Msgf("Direct latency data missing for %d region pair(s) — using cluster-based estimates (shown as 'X.X (est.)' in table)",
			estimatedPairs)
	}
	if unknownPairs > 0 {
		log.Warn().Msgf("No latency data or estimate for %d region pair(s) — shown as 'N/A', scored as neutral",
			unknownPairs)
	}

	log.Debug().Msg("Latency calculation completed")
}

// getRegionLatency returns the round-trip time in milliseconds between two Azure regions.
// The bool return value is true when the value is a cluster-based estimate rather than
// a direct measurement.
//
// Lookup chain:
//  1. Exact measured pair in the latency matrix
//  2. Reverse pair (RTT is near-symmetric)
//  3. Geographic cluster-pair average computed from the measured matrix at startup
//
// Returns (0, false) when no measured data or cluster estimate is available.
func getRegionLatency(sourceRegion, targetRegion string) (float64, bool) {
	if sourceRegion == targetRegion {
		return 0, false
	}

	// 1. Exact pair
	if sourceLatencies, ok := azureRegionLatency[sourceRegion]; ok {
		if latency, ok := sourceLatencies[targetRegion]; ok {
			return latency, false
		}
	}

	// 2. Reverse lookup (RTT is symmetric)
	if targetLatencies, ok := azureRegionLatency[targetRegion]; ok {
		if latency, ok := targetLatencies[sourceRegion]; ok {
			return latency, false
		}
	}

	// 3. Geographic cluster-pair average
	srcCluster := regionCluster[sourceRegion]
	tgtCluster := regionCluster[targetRegion]
	if srcCluster != "" && tgtCluster != "" {
		key := srcCluster + ":" + tgtCluster
		if avg, ok := clusterPairAverages[key]; ok && avg > 0 {
			log.Debug().Msgf("No direct latency for %s→%s — using %s:%s cluster average (%.0f ms)",
				sourceRegion, targetRegion, srcCluster, tgtCluster, avg)
			return avg, true
		}
	}

	log.Debug().Msgf("No latency data or estimate for %s→%s", sourceRegion, targetRegion)
	return 0, false
}
