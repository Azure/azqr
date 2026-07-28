// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package skus

import (
	"regexp"
	"sort"
	"strconv"
)

// Recommendation represents an alternative SKU suggestion with a compatibility score.
type Recommendation struct {
	SKU                SKU     `json:"sku"`
	CompatibilityScore float64 `json:"compatibilityScore"`
}

// versionRe extracts the generation suffix _vN (or _vN_...) from an Azure VM SKU name.
var versionRe = regexp.MustCompile(`(?i)_v(\d+)`)

// baseNameRe strips the _vN… suffix to produce the base model name (e.g. "Standard_D16s").
var baseNameRe = regexp.MustCompile(`(?i)_v\d+.*$`)

// extractVersion returns the generation version found in the SKU name, or 0 if absent.
func extractVersion(name string) int {
	m := versionRe.FindStringSubmatch(name)
	if m == nil {
		return 0
	}
	v, _ := strconv.Atoi(m[1])
	return v
}

// extractBaseName returns the SKU name with the _vN… suffix removed
// (e.g. "Standard_D16s_v5" → "Standard_D16s", "Standard_D11_v2_Promo" → "Standard_D11").
func extractBaseName(name string) string {
	return baseNameRe.ReplaceAllString(name, "")
}

// FindAlternatives returns the top n alternative SKUs for the given target SKU,
// ranked by compatibility score using the same algorithm as azure-assessor's SkuAdvisor.
func FindAlternatives(target SKU, n int) []Recommendation {
	var recommendations []Recommendation

	for _, candidate := range ListAll() {
		if candidate.Name == target.Name {
			continue
		}

		score := computeScore(target, candidate)
		if score < 0.1 {
			continue
		}

		recommendations = append(recommendations, Recommendation{
			SKU:                candidate,
			CompatibilityScore: score,
		})
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].CompatibilityScore > recommendations[j].CompatibilityScore
	})

	if n > 0 && len(recommendations) > n {
		return recommendations[:n]
	}
	return recommendations
}

// computeScore computes a compatibility score between 0.0 and 1.0.
//
// Weight budget (total may exceed 1.0 and is capped):
//
//	vCPUs           0.30
//	Memory          0.25
//	Same base model 0.15  (same SKU name ignoring _vN suffix, e.g. Standard_D16s)
//	Family          0.10  (same Azure VM family string)
//	Version         0.05  (same/newer: 0.05; each generation behind costs 0.02, min 0.0)
//	GPU             0.15  (or 0.10 flat for non-GPU workloads)
//	Data disks      0.05
//	Accel. network  0.05
func computeScore(target, candidate SKU) float64 {
	var total float64

	// vCPU match (weight 0.30)
	if target.VCPUs > 0 {
		lo, hi := float64(target.VCPUs), float64(candidate.VCPUs)
		if lo > hi {
			lo, hi = hi, lo
		}
		total += (lo / hi) * 0.30
	}

	// Memory match (weight 0.25)
	if target.MemoryGB > 0 {
		lo, hi := target.MemoryGB, candidate.MemoryGB
		if lo > hi {
			lo, hi = hi, lo
		}
		total += (lo / hi) * 0.25
	}

	// Same base-model bonus (weight 0.15): same SKU series, different generation.
	targetBase := extractBaseName(target.Name)
	candidateBase := extractBaseName(candidate.Name)
	if targetBase != "" && targetBase == candidateBase {
		total += 0.15
	}

	// Same-family bonus (weight 0.10).
	if target.Family != "" && target.Family == candidate.Family {
		total += 0.10
	}

	// Version proximity (weight 0.05).
	total += scoreVersion(extractVersion(target.Name), extractVersion(candidate.Name))

	// GPU match (weight 0.15)
	if target.GPUCount > 0 {
		if candidate.GPUCount >= target.GPUCount {
			total += 0.15
		} else if candidate.GPUCount > 0 {
			total += float64(candidate.GPUCount) / float64(target.GPUCount) * 0.15
		}
	} else {
		total += 0.10 // non-GPU workload
	}

	// Data disk match (weight 0.05)
	if target.MaxDataDisks > 0 {
		lo, hi := float64(target.MaxDataDisks), float64(candidate.MaxDataDisks)
		if lo > hi {
			lo, hi = hi, lo
		}
		if hi > 0 {
			total += (lo / hi) * 0.05
		}
	} else {
		total += 0.05
	}

	// Accelerated networking match (weight 0.05)
	if !target.AcceleratedNetworking || candidate.AcceleratedNetworking {
		total += 0.05
	}

	if total > 1.0 {
		total = 1.0
	}
	return round3(total)
}

// scoreVersion returns a version-proximity score (0.0–0.05).
// When either version is unknown (0), a neutral score of 0.025 is assigned.
// Each generation behind the target costs 0.02, floored at 0.0.
func scoreVersion(targetVer, candidateVer int) float64 {
	if targetVer == 0 || candidateVer == 0 {
		return 0.025
	}
	diff := targetVer - candidateVer
	if diff <= 0 {
		return 0.05
	}
	// Use integer math to avoid float precision drift: score = max(0, 5 - diff*2) / 100
	millis := 5 - diff*2
	if millis < 0 {
		millis = 0
	}
	return float64(millis) / 100.0
}

func round3(v float64) float64 {
	return float64(int(v*1000+0.5)) / 1000
}

