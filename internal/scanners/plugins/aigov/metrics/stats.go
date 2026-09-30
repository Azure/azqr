// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package metrics

import "sort"

// Percentile returns the linear-interpolated Percentile (0-100) of values.
func Percentile(values []float64, pct float64) float64 {
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
