// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sqleol

import (
	"encoding/json"
	"testing"

	"github.com/Azure/azqr/internal/plugins"
)

func TestNewScanner(t *testing.T) {
	if NewScanner() == nil {
		t.Fatal("NewScanner returned nil")
	}
}

func TestScanner_GetMetadata(t *testing.T) {
	meta := NewScanner().GetMetadata()

	if meta.Name != "sql-eol" {
		t.Errorf("Name = %q, want sql-eol", meta.Name)
	}
	if meta.Version == "" {
		t.Error("Version must not be empty")
	}
	if meta.Type != plugins.PluginTypeInternal {
		t.Errorf("Type = %v, want PluginTypeInternal", meta.Type)
	}
	// sql-eol exposes a wide table; guard against an accidental large drop in columns.
	if len(meta.ColumnMetadata) < 20 {
		t.Errorf("ColumnMetadata len = %d, want >= 20", len(meta.ColumnMetadata))
	}

	seen := make(map[string]bool, len(meta.ColumnMetadata))
	for i, col := range meta.ColumnMetadata {
		if col.Name == "" {
			t.Errorf("ColumnMetadata[%d] has empty Name", i)
		}
		if seen[col.Name] {
			t.Errorf("duplicate Name %q at index %d", col.Name, i)
		}
		seen[col.Name] = true
	}
}

// TestSQLESURow_Unmarshal verifies the JSON tag mapping from the ARG query
// result into sqlEOLRow, including the lower-cased "vCores" tag.
func TestSQLESURow_Unmarshal(t *testing.T) {
	raw := []byte(`{
		"SubscriptionId": "sub-123",
		"Name": "sql-vm-1",
		"ResourceGroup": "rg-sql",
		"Edition": "Enterprise",
		"vCores": "8",
		"EOLStatus": "Out of Support",
		"SQLMIMigrationVerdict": "Recommended",
		"ConsolidationRatio": "2"
	}`)

	var r sqlEOLRow
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if r.SubscriptionID != "sub-123" {
		t.Errorf("SubscriptionID = %q, want sub-123", r.SubscriptionID)
	}
	if r.VCores != "8" {
		t.Errorf("VCores = %q, want 8 (check 'vCores' json tag)", r.VCores)
	}
	if r.Edition != "Enterprise" {
		t.Errorf("Edition = %q, want Enterprise", r.Edition)
	}
	if r.SQLMIMigrationVerdict != "Recommended" {
		t.Errorf("SQLMIMigrationVerdict = %q, want Recommended", r.SQLMIMigrationVerdict)
	}
	if r.ConsolidationRatio != "2" {
		t.Errorf("ConsolidationRatio = %q, want 2", r.ConsolidationRatio)
	}
}

// TestSQLESURow_ToRecord verifies the flattened record preserves field order and
// has one entry per declared column.
func TestSQLESURow_ToRecord(t *testing.T) {
	r := sqlEOLRow{
		Name:                  "sql-vm-1",
		ResourceGroup:         "rg-sql",
		Subscription:          "Prod",
		Location:              "eastus",
		ArcServerName:         "arc-host-1",
		CloudType:             "Arc-enabled (On-Prem)",
		ServiceType:           "Engine",
		Edition:               "Enterprise",
		VCores:                "8",
		EOLStatus:             "Out of Support",
		ESUApplicable:         "Yes",
		ESUEnabled:            "Enabled",
		SQLMIMigrationVerdict: "Recommended",
	}

	record := r.toRecord()

	wantLen := len(NewScanner().GetMetadata().ColumnMetadata)
	if len(record) != wantLen {
		t.Fatalf("record len = %d, want %d (one per column)", len(record), wantLen)
	}

	// Spot-check ordering against the declared column positions.
	checks := []struct {
		idx  int
		want string
		desc string
	}{
		{0, "Prod", "Subscription"},
		{1, "rg-sql", "Resource Group"},
		{2, "sql-vm-1", "Name"},
		{3, "eastus", "Location"},
		{4, "arc-host-1", "Arc Server Name"},
		{5, "Arc-enabled (On-Prem)", "Cloud Type"},
		{6, "Engine", "Service Type"},
		{8, "Enterprise", "Edition"},
		{10, "Yes", "ESU Applicable"},
		{11, "Enabled", "ESU Enabled"},
	}
	for _, c := range checks {
		if record[c.idx] != c.want {
			t.Errorf("record[%d] (%s) = %q, want %q", c.idx, c.desc, record[c.idx], c.want)
		}
	}
	// SQLMIMigrationVerdict is the last column.
	if record[len(record)-1] != "Recommended" {
		t.Errorf("last record = %q, want Recommended", record[len(record)-1])
	}
}
