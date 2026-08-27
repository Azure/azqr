// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package skus

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
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

// gpuCategoryRe matches the GPU workload class prefix in N-series SKU names.
// Groups: 1=category letter (V=visualization, C=compute, D=distributed).
var gpuCategoryRe = regexp.MustCompile(`(?i)^Standard_N([VCD])\d`)

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

// extractGPUCategory returns the GPU workload class for N-series SKU names:
// "V" for visualization (NV), "C" for compute (NC), "D" for distributed (ND).
// Returns "" for non-GPU SKUs.
func extractGPUCategory(name string) string {
	m := gpuCategoryRe.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	return strings.ToUpper(m[1])
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

	slices.SortFunc(recommendations, func(a, b Recommendation) int {
		if a.CompatibilityScore > b.CompatibilityScore {
			return -1
		}
		if a.CompatibilityScore < b.CompatibilityScore {
			return 1
		}
		return 0
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
//	vCPUs               0.30
//	Memory              0.25
//	Same base model     0.15  (same SKU name ignoring _vN suffix, e.g. Standard_D16s)
//	Family              same=-0.10 / cross=+0.10 scaled by generation gap
//	Version             0.05  (same/newer: 0.05; each generation behind costs 0.02, min 0.0)
//	GPU count           0.15  (or 0.10 flat for non-GPU workloads)
//	GPU workload class  0.20  (NV/NC/ND same class bonus, GPU SKUs only)
//	Data disks          0.05
//	Accel. network      0.05
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

	// Cross-family scoring: different families draw from independent capacity pools,
	// making them more useful alternatives when the target family is constrained.
	// Same-family candidates are penalized (-0.10) since they face the same capacity limits.
	// Cross-family bonus: 0.10 base, reduced by 0.02 per generation behind the target, floor 0.0.
	if target.Family != "" && target.Family == candidate.Family {
		total -= 0.10
	} else {
		diff := extractVersion(target.Name) - extractVersion(candidate.Name)
		if diff < 0 {
			diff = 0
		}
		bonus := 0.10 - float64(diff)*0.02
		if bonus > 0 {
			total += bonus
		}
	}

	// Version proximity (weight 0.05).
	total += scoreVersion(extractVersion(target.Name), extractVersion(candidate.Name))

	// GPU count match (weight 0.15)
	if target.GPUCount > 0 {
		if candidate.GPUCount >= target.GPUCount {
			total += 0.15
		} else if candidate.GPUCount > 0 {
			total += (candidate.GPUCount / target.GPUCount) * 0.15
		}

		// GPU workload class bonus (weight 0.20): NV=visualization, NC=compute, ND=distributed.
		// Weighted higher than memory because GPU VM sizing is driven by partition count/workload
		// type, not raw RAM — a same-class candidate is always preferable regardless of memory gap.
		targetCat := extractGPUCategory(target.Name)
		candidateCat := extractGPUCategory(candidate.Name)
		if targetCat != "" && targetCat == candidateCat {
			total += 0.20
		}
	} else {
		total += 0.10 // non-GPU workload flat bonus
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
	centis := 5 - diff*2
	if centis < 0 {
		centis = 0
	}
	return float64(centis) / 100.0
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

