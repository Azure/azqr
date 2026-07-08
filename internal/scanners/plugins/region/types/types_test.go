// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package types

import (
	"testing"
)

func TestDefaultScoringWeights(t *testing.T) {
	w := DefaultScoringWeights()

	if w.ResourceAvailability != 0.35 {
		t.Errorf("ResourceAvailability = %v, want 0.35", w.ResourceAvailability)
	}
	if w.SKUAvailability != 0.30 {
		t.Errorf("SKUAvailability = %v, want 0.30", w.SKUAvailability)
	}
	if w.Cost != 0.15 {
		t.Errorf("Cost = %v, want 0.15", w.Cost)
	}
	if w.Latency != 0.20 {
		t.Errorf("Latency = %v, want 0.20", w.Latency)
	}

	// Weights must sum to 1.0 so the weighted score stays in the 0-100 range.
	sum := w.ResourceAvailability + w.SKUAvailability + w.Cost + w.Latency
	if diff := sum - 1.0; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("weights sum to %v, want 1.0", sum)
	}
}

func TestResourceTypeLocationData_IsAvailable(t *testing.T) {
	rtl := &ResourceTypeLocationData{
		Data: map[string]map[string]map[string]struct{}{
			"microsoft.compute": {
				"virtualmachines": {"eastus": {}, "westeurope": {}},
			},
		},
	}

	tests := []struct {
		name         string
		resourceType string
		region       string
		want         bool
	}{
		{
			name:         "available type and region",
			resourceType: "microsoft.compute/virtualmachines",
			region:       "eastus",
			want:         true,
		},
		{
			name:         "type present but region absent",
			resourceType: "microsoft.compute/virtualmachines",
			region:       "brazilsouth",
			want:         false,
		},
		{
			name:         "namespace present but type absent",
			resourceType: "microsoft.compute/disks",
			region:       "eastus",
			want:         false,
		},
		{
			name:         "unknown namespace",
			resourceType: "microsoft.storage/storageaccounts",
			region:       "eastus",
			want:         false,
		},
		{
			name:         "missing slash separator",
			resourceType: "microsoft.compute",
			region:       "eastus",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rtl.IsAvailable(tt.resourceType, tt.region); got != tt.want {
				t.Errorf("isAvailable(%q, %q) = %v, want %v", tt.resourceType, tt.region, got, tt.want)
			}
		})
	}
}
