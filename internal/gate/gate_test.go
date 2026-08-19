// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package gate

import (
	"errors"
	"testing"

	"github.com/Azure/azqr/internal/findings"
)

func TestParse(t *testing.T) {
	criterion, err := Parse(" medium ")
	if err != nil {
		t.Fatal(err)
	}
	if criterion.Threshold != "Medium" {
		t.Fatalf("got %q, want Medium", criterion.Threshold)
	}
	if _, err := Parse("Critical"); err == nil {
		t.Fatal("expected invalid impact error")
	}
}

func TestCheckUsesSeverityThreshold(t *testing.T) {
	summary := &findings.Summary{Recommendations: []findings.Recommendation{
		{ID: "high", Impact: "High", ImpactedResources: 2},
		{ID: "medium", Impact: "Medium", ImpactedResources: 3},
		{ID: "low", Impact: "Low", ImpactedResources: 4},
		{ID: "zero", Impact: "High", ImpactedResources: 0},
	}}
	criterion, _ := Parse("Medium")

	err := Check(summary, criterion)
	var failure *Failure
	if !errors.As(err, &failure) {
		t.Fatalf("got %v, want gate Failure", err)
	}
	if failure.Recommendations != 2 || failure.Impacted != 5 {
		t.Fatalf("unexpected failure: %+v", failure)
	}
}

func TestCheckPassesWithoutMatchingFindings(t *testing.T) {
	criterion, _ := Parse("High")
	if err := Check(&findings.Summary{Recommendations: []findings.Recommendation{
		{ID: "medium", Impact: "Medium", ImpactedResources: 1},
	}}, criterion); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
