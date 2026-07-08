// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package carbon

import (
	"testing"
	"time"

	"github.com/Azure/azqr/internal/plugins"
)

func TestNewEmissionsScanner(t *testing.T) {
	if NewScanner() == nil {
		t.Fatal("NewEmissionsScanner returned nil")
	}
}

func TestEmissionsScanner_GetMetadata(t *testing.T) {
	meta := NewScanner().GetMetadata()

	if meta.Name != "carbon-emissions" {
		t.Errorf("Name = %q, want carbon-emissions", meta.Name)
	}
	if meta.Version == "" {
		t.Error("Version must not be empty")
	}
	if meta.Type != plugins.PluginTypeInternal {
		t.Errorf("Type = %v, want PluginTypeInternal", meta.Type)
	}
	if len(meta.ColumnMetadata) != 8 {
		t.Errorf("ColumnMetadata len = %d, want 8", len(meta.ColumnMetadata))
	}

	assertDataKeysValid(t, meta)
}

// assertDataKeysValid checks that every column has a non-empty, unique DataKey.
// HeaderRow consistency is covered separately; this guards the DataKeys used by
// the web viewer and filters.
func assertDataKeysValid(t *testing.T, meta plugins.PluginMetadata) {
	t.Helper()
	seen := make(map[string]bool, len(meta.ColumnMetadata))
	for i, col := range meta.ColumnMetadata {
		if col.DataKey == "" {
			t.Errorf("ColumnMetadata[%d] (%q) has empty DataKey", i, col.Name)
		}
		if seen[col.DataKey] {
			t.Errorf("duplicate DataKey %q at index %d", col.DataKey, i)
		}
		seen[col.DataKey] = true
	}
}

func TestBuildEmissionRow(t *testing.T) {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		agg      aggregatedEmissions
		expected []string
	}{
		{
			name: "all fields populated",
			agg:  aggregatedEmissions{latestMonth: 150, previousMonth: 100, monthlyChangeValue: 50},
			expected: []string{
				"2026-03-01", "2026-03-31", "Microsoft.Compute/virtualMachines",
				"150.00", "100.00", "50.00%", "50.00", "kgCO2e",
			},
		},
		{
			name: "no previous month leaves ratio and previous empty",
			agg:  aggregatedEmissions{latestMonth: 75},
			expected: []string{
				"2026-03-01", "2026-03-31", "Microsoft.Compute/virtualMachines",
				"75.00", "", "", "", "kgCO2e",
			},
		},
		{
			name: "negative change ratio",
			agg:  aggregatedEmissions{latestMonth: 80, previousMonth: 100, monthlyChangeValue: -20},
			expected: []string{
				"2026-03-01", "2026-03-31", "Microsoft.Compute/virtualMachines",
				"80.00", "100.00", "-20.00%", "-20.00", "kgCO2e",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := buildEmissionRow(from, to, "Microsoft.Compute/virtualMachines", tt.agg)
			if len(row) != len(tt.expected) {
				t.Fatalf("row len = %d, want %d", len(row), len(tt.expected))
			}
			for i := range tt.expected {
				if row[i] != tt.expected[i] {
					t.Errorf("row[%d] = %q, want %q", i, row[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseAvailableDateRange(t *testing.T) {
	from, to, err := parseAvailableDateRange("2026-01-01", "2026-03-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both returned values must be the end date (latest available month).
	want := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	if !from.Equal(want) || !to.Equal(want) {
		t.Errorf("got (%s, %s), want both %s", from, to, want)
	}
}

func TestParseAvailableDateRange_Invalid(t *testing.T) {
	if _, _, err := parseAvailableDateRange("2026-01-01", "not-a-date"); err == nil {
		t.Error("expected error for invalid end date, got nil")
	}
	if _, _, err := parseAvailableDateRange("bad", "2026-03-31"); err == nil {
		t.Error("expected error for invalid start date, got nil")
	}
}
