// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"regexp"
	"strings"
)

// documentedReplacement is one row of the Microsoft-documented replacement VM
// series table. The rows live in zz_generated.replacement_series.go, produced
// by hack/code/replacement_series_gen (`make replacement-series`).
type documentedReplacement struct {
	Sources     []string
	Targets     []string
	Differences []string
}

const (
	replacementNotApplicable = "N/A"
	replacementNoAction      = "No action (already a documented target)"
	replacementReview        = "Review newer candidates (no documented replacement)"
)

// vmSizeParts splits a size such as Standard_NC40ads_H100_v5 into its family
// letters (NC), vCPU count, constrained-vCPU suffix, feature letters (ads),
// optional middle token (H100) and version (5).
var vmSizeParts = regexp.MustCompile(`(?i)^([a-z]+?)(\d+)(?:-\d+)?([a-z]*)(?:_([a-z0-9]+))?$`)

// seriesKeys returns the normalized series identifiers a VM size can be
// matched on, most specific first (for example "dsv3" then "dv3"). Keys are
// lowercase with separators removed, which is how table entries are
// normalized too.
func seriesKeys(vmSize string) []string {
	name := strings.TrimPrefix(strings.ToLower(vmSize), "standard_")

	version := ""
	if m := seriesVersionRe.FindStringSubmatch(name); m != nil {
		version = "v" + m[1]
		name = name[:len(name)-len(m[0])]
	}

	m := vmSizeParts.FindStringSubmatch(name)
	if m == nil {
		return nil
	}
	family, features, middle := m[1], m[3], m[4]

	// Bv1 sizes have no version suffix in their names.
	if version == "" && family == "b" {
		version = "v1"
	}

	keys := []string{family + features + middle + version}
	if version != "" {
		keys = append(keys, family+version)
	} else {
		keys = append(keys, family)
	}
	return keys
}

func normalizeSeries(s string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(s))
}

func containsSeries(keys []string, candidate string) bool {
	candidate = normalizeSeries(candidate)
	for _, k := range keys {
		if k == candidate {
			return true
		}
	}
	return false
}

// documentedTargets lists the series named in a replacement row's targets.
func documentedTargets(row documentedReplacement) []string {
	var series []string
	for _, line := range row.Targets {
		for _, token := range strings.FieldsFunc(line, func(r rune) bool { return r == '/' || r == ' ' }) {
			series = append(series, normalizeSeries(token))
		}
	}
	return series
}

// lookupDocumentedReplacement returns the documented target series and the
// documented specification differences for a VM size. Sizes already on a
// target generation or on a documented target series report no action; sizes
// without a documented replacement are flagged for review.
func lookupDocumentedReplacement(vmSize string, isTargetGeneration bool) (replacement, differences string) {
	if isTargetGeneration {
		return replacementNoAction, replacementNotApplicable
	}

	keys := seriesKeys(vmSize)
	for _, row := range documentedReplacements {
		for _, target := range documentedTargets(row) {
			if containsSeries(keys, target) {
				return replacementNoAction, replacementNotApplicable
			}
		}
	}
	for _, key := range keys {
		for _, row := range documentedReplacements {
			for _, src := range row.Sources {
				if normalizeSeries(src) == key {
					return strings.Join(row.Targets, "; "), strings.Join(row.Differences, "; ")
				}
			}
		}
	}
	return replacementReview, replacementNotApplicable
}
