// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package quota

import (
	"testing"
)

func TestAtRiskSummaries(t *testing.T) {
	tests := []struct {
		name    string
		usages  []UsageEntry
		wantLen int    // expected number of at-risk entries
		wantStr string // expected content of first entry (empty = skip check)
	}{
		{
			name:    "empty input returns nil",
			usages:  nil,
			wantLen: 0,
		},
		{
			name: "no near-limit entries",
			usages: []UsageEntry{
				{ResourceName: "standardDSv3Family", CurrentValue: 10, Limit: 100, Available: 90, HeadroomPct: 90},
			},
			wantLen: 0,
		},
		{
			name: "near-limit entry flagged with counts",
			usages: []UsageEntry{
				{ResourceName: "standardDSv3Family", LocalizedName: "Standard DSv3 Family vCPUs", CurrentValue: 90, Limit: 100, Available: 10, HeadroomPct: 10, IsNearLimit: true},
			},
			wantLen: 1,
			wantStr: "Standard DSv3 Family vCPUs (90/100, 90% used)",
		},
		{
			name: "at-limit entry falls back to ResourceName when no LocalizedName",
			usages: []UsageEntry{
				{ResourceName: "standardDSv3Family", CurrentValue: 100, Limit: 100, Available: 0, HeadroomPct: 0, IsAtOrOverLimit: true},
			},
			wantLen: 1,
			wantStr: "standardDSv3Family (100/100, 100% used)",
		},
		{
			name: "mixed entries — only risky ones returned",
			usages: []UsageEntry{
				{ResourceName: "standardDSv3Family", CurrentValue: 10, Limit: 100, Available: 90, HeadroomPct: 90},
				{ResourceName: "standardEv4Family", LocalizedName: "Standard Ev4 Family vCPUs", CurrentValue: 95, Limit: 100, Available: 5, HeadroomPct: 5, IsNearLimit: true},
			},
			wantLen: 1,
			wantStr: "Standard Ev4 Family vCPUs (95/100, 95% used)",
		},
		{
			name: "network resource near-limit",
			usages: []UsageEntry{
				{ResourceName: "PublicIPAddresses", LocalizedName: "Public IP Addresses", CurrentValue: 48, Limit: 50, Available: 2, HeadroomPct: 4, IsNearLimit: true},
			},
			wantLen: 1,
			wantStr: "Public IP Addresses (48/50, 96% used)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AtRiskSummaries(tt.usages)
			if len(got) != tt.wantLen {
				t.Errorf("AtRiskSummaries() len = %d, want %d; got %v", len(got), tt.wantLen, got)
			}
			if tt.wantStr != "" && (len(got) == 0 || got[0] != tt.wantStr) {
				t.Errorf("AtRiskSummaries()[0] = %q, want %q", got[0], tt.wantStr)
			}
		})
	}
}

func TestUsageEntry_HeadroomCalculation(t *testing.T) {
	limit := 100
	current := 88
	available := limit - current
	headroomPct := float64(available) / float64(limit) * 100

	u := UsageEntry{
		ResourceName:    "standardDSv3Family",
		CurrentValue:    current,
		Limit:           limit,
		Available:       available,
		HeadroomPct:     headroomPct,
		IsNearLimit:     headroomPct < nearLimitThreshold*100,
		IsAtOrOverLimit: available <= 0,
	}

	if !u.IsNearLimit {
		t.Errorf("expected IsNearLimit=true for %.1f%% headroom (threshold %.0f%%)", headroomPct, nearLimitThreshold*100)
	}
	if u.IsAtOrOverLimit {
		t.Errorf("expected IsAtOrOverLimit=false when available=%d", available)
	}
}

