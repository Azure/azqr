// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "sort"

// tpmStats computes the average and p95 total (input+output) tokens-per-
// minute across a set of hourly usage buckets, along with the count of
// "active" buckets (nonzero usage), mirroring ptu.py's separation of
// elapsed/observed/active evidence without its full statistical machinery.
func tpmStats(samples []hourlyTokenUsage) (avgTPM, p95TPM float64, activeBuckets int) {
	tpmValues := make([]float64, 0, len(samples))
	var sum float64

	for _, s := range samples {
		total := s.InputTokens + s.OutputTokens
		if total <= 0 {
			continue
		}
		tpm := total / 60 // one hourly bucket's tokens spread over 60 minutes
		tpmValues = append(tpmValues, tpm)
		sum += tpm
	}

	activeBuckets = len(tpmValues)
	if activeBuckets == 0 {
		return 0, 0, 0
	}

	avgTPM = sum / float64(activeBuckets)
	p95TPM = percentile(tpmValues, 95)
	return avgTPM, p95TPM, activeBuckets
}

// percentile returns the linear-interpolated percentile (0-100) of values.
func percentile(values []float64, pct float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)

	position := float64(len(ordered)-1) * pct / 100
	lower := int(position)
	upper := lower + 1
	if upper > len(ordered)-1 {
		upper = len(ordered) - 1
	}
	fraction := position - float64(lower)
	return ordered[lower] + (ordered[upper]-ordered[lower])*fraction
}
