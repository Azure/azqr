// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/models"
)

// TestDeriveResourceTypeTotals verifies that aggregating per-subscription
// ResourceTypeCount rows produces the correct per-type totals.
// This is the core of fix #9 — replacing a second Azure Graph round-trip
// with local computation over already-fetched data.
func TestDeriveResourceTypeTotals_AggregatesAcrossSubscriptions(t *testing.T) {
	input := []*models.ResourceTypeCount{
		{Subscription: "sub1", ResourceType: "microsoft.compute/virtualmachines", Count: 3},
		{Subscription: "sub2", ResourceType: "microsoft.compute/virtualmachines", Count: 5},
		{Subscription: "sub1", ResourceType: "microsoft.storage/storageaccounts", Count: 10},
	}

	totals := deriveResourceTypeTotals(input)

	if got := totals["microsoft.compute/virtualmachines"]; got != 8 {
		t.Errorf("VMs: got %.0f, want 8", got)
	}
	if got := totals["microsoft.storage/storageaccounts"]; got != 10 {
		t.Errorf("Storage: got %.0f, want 10", got)
	}
}

func TestDeriveResourceTypeTotals_NormalisesToLowercase(t *testing.T) {
	input := []*models.ResourceTypeCount{
		{ResourceType: "Microsoft.Compute/VirtualMachines", Count: 2},
		{ResourceType: "MICROSOFT.COMPUTE/VIRTUALMACHINES", Count: 3},
	}

	totals := deriveResourceTypeTotals(input)

	if got := totals["microsoft.compute/virtualmachines"]; got != 5 {
		t.Errorf("case-folded total: got %.0f, want 5", got)
	}
	// No uppercase key should remain
	for k := range totals {
		if k != strings.ToLower(k) {
			t.Errorf("key %q is not lowercase — normalisation failed", k)
		}
	}
}

func TestDeriveResourceTypeTotals_EmptyInput(t *testing.T) {
	totals := deriveResourceTypeTotals(nil)
	if len(totals) != 0 {
		t.Errorf("expected empty map for nil input, got %v", totals)
	}
}

func TestDeriveResourceTypeTotals_SingleEntry(t *testing.T) {
	input := []*models.ResourceTypeCount{
		{ResourceType: "microsoft.network/virtualnetworks", Count: 7},
	}
	totals := deriveResourceTypeTotals(input)
	if got := totals["microsoft.network/virtualnetworks"]; got != 7 {
		t.Errorf("got %.0f, want 7", got)
	}
}
