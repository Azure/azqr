// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package gate evaluates scan findings against CI/CD policy thresholds.
package gate

import (
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
)

// Criterion describes the minimum severity that fails a scan.
type Criterion struct {
	Threshold string
	rank      int
}

// Result describes findings that matched a gate criterion.
type Result struct {
	Recommendations int
	Impacted        int
}

// Failure is returned when scan findings cross a configured gate.
type Failure struct {
	Threshold       string
	Recommendations int
	Impacted        int
}

// Error implements error.
func (e *Failure) Error() string {
	return fmt.Sprintf(
		"scan gate failed: %d recommendation(s) at or above %s impact %d resource finding(s)",
		e.Recommendations,
		e.Threshold,
		e.Impacted,
	)
}

// Parse creates a severity criterion from a CLI value.
func Parse(value string) (Criterion, error) {
	value = strings.TrimSpace(value)
	rank, ok := models.SeverityRank(value)
	if !ok {
		return Criterion{}, fmt.Errorf("invalid fail-on impact %q: supported values are High, Medium, Low", value)
	}

	threshold := string(models.ImpactLow)
	switch rank {
	case 3:
		threshold = string(models.ImpactHigh)
	case 2:
		threshold = string(models.ImpactMedium)
	}
	return Criterion{Threshold: threshold, rank: rank}, nil
}

// Evaluate counts impacted recommendations at or above the criterion.
func Evaluate(summary *findings.Summary, criterion Criterion) Result {
	var result Result
	if summary == nil {
		return result
	}

	for _, recommendation := range summary.Recommendations {
		if recommendation.ImpactedResources == 0 {
			continue
		}
		rank, ok := models.SeverityRank(recommendation.Impact)
		if !ok || rank < criterion.rank {
			continue
		}
		result.Recommendations++
		result.Impacted += recommendation.ImpactedResources
	}
	return result
}

// Check returns a Failure when findings cross the criterion.
func Check(summary *findings.Summary, criterion Criterion) error {
	result := Evaluate(summary, criterion)
	if result.Recommendations == 0 {
		return nil
	}
	return &Failure{
		Threshold:       criterion.Threshold,
		Recommendations: result.Recommendations,
		Impacted:        result.Impacted,
	}
}
