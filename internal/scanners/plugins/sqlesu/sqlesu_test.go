// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sqlesu

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

	if meta.Name != "sql-esu" {
		t.Errorf("Name = %q, want sql-esu", meta.Name)
	}
	if meta.Version == "" {
		t.Error("Version must not be empty")
	}
	if meta.Type != plugins.PluginTypeInternal {
		t.Errorf("Type = %v, want PluginTypeInternal", meta.Type)
	}
	// sql-esu exposes a wide table; guard against an accidental large drop in columns.
	if len(meta.ColumnMetadata) < 20 {
		t.Errorf("ColumnMetadata len = %d, want >= 20", len(meta.ColumnMetadata))
	}

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

// TestSQLESURow_Unmarshal verifies the JSON tag mapping from the ARG query
// result into sqlESURow, including the lower-cased "vCores" tag.
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

	var r sqlESURow
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
	r := sqlESURow{
		Name:                  "sql-vm-1",
		ResourceGroup:         "rg-sql",
		Subscription:          "Prod",
		Location:              "eastus",
		Edition:               "Enterprise",
		VCores:                "8",
		EOLStatus:             "Out of Support",
		SQLMIMigrationVerdict: "Recommended",
	}

	record := r.toRecord()

	wantLen := len(NewScanner().GetMetadata().ColumnMetadata)
	if len(record) != wantLen {
		t.Fatalf("record len = %d, want %d (one per column)", len(record), wantLen)
	}

	// Spot-check ordering against the first columns and the final column.
	if record[0] != "Prod" {
		t.Errorf("record[0] = %q, want Prod (Subscription)", record[0])
	}
	if record[1] != "rg-sql" {
		t.Errorf("record[1] = %q, want rg-sql (Resource Group)", record[1])
	}
	if record[2] != "sql-vm-1" {
		t.Errorf("record[2] = %q, want sql-vm-1 (Name)", record[2])
	}
	if record[6] != "Enterprise" {
		t.Errorf("record[6] = %q, want Enterprise", record[6])
	}
	// SQLMIMigrationVerdict is the last column (index 26).
	if record[len(record)-1] != "Recommended" {
		t.Errorf("last record = %q, want Recommended", record[len(record)-1])
	}
}
