// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package zone

import (
	"encoding/json"
	"testing"

	"github.com/Azure/azqr/internal/plugins"
)

func TestNewZoneMappingScanner(t *testing.T) {
	if NewScanner() == nil {
		t.Fatal("NewZoneMappingScanner returned nil")
	}
}

func TestZoneMappingScanner_GetMetadata(t *testing.T) {
	meta := NewScanner().GetMetadata()

	if meta.Name != "zone-mapping" {
		t.Errorf("Name = %q, want zone-mapping", meta.Name)
	}
	if meta.Version == "" {
		t.Error("Version must not be empty")
	}
	if meta.Type != plugins.PluginTypeInternal {
		t.Errorf("Type = %v, want PluginTypeInternal", meta.Type)
	}
	if len(meta.ColumnMetadata) != 5 {
		t.Errorf("ColumnMetadata len = %d, want 5", len(meta.ColumnMetadata))
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

// TestLocationResponse_Unmarshal verifies the JSON contract that fetchZoneMappings
// relies on: the locations REST response shape, optional (pointer) fields, and
// nested availability zone mappings.
func TestLocationResponse_Unmarshal(t *testing.T) {
	body := []byte(`{
		"value": [
			{
				"name": "eastus",
				"displayName": "East US",
				"availabilityZoneMappings": [
					{"logicalZone": "1", "physicalZone": "eastus-az1"},
					{"logicalZone": "2", "physicalZone": "eastus-az2"}
				]
			},
			{
				"name": "westus",
				"displayName": "West US"
			}
		]
	}`)

	var resp locationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(resp.Value) != 2 {
		t.Fatalf("Value len = %d, want 2", len(resp.Value))
	}

	east := resp.Value[0]
	if east.Name == nil || *east.Name != "eastus" {
		t.Errorf("Value[0].Name = %v, want eastus", east.Name)
	}
	if east.DisplayName == nil || *east.DisplayName != "East US" {
		t.Errorf("Value[0].DisplayName = %v, want East US", east.DisplayName)
	}
	if len(east.AvailabilityZoneMappings) != 2 {
		t.Fatalf("Value[0] zone mappings len = %d, want 2", len(east.AvailabilityZoneMappings))
	}
	m := east.AvailabilityZoneMappings[0]
	if m.LogicalZone == nil || *m.LogicalZone != "1" {
		t.Errorf("LogicalZone = %v, want 1", m.LogicalZone)
	}
	if m.PhysicalZone == nil || *m.PhysicalZone != "eastus-az1" {
		t.Errorf("PhysicalZone = %v, want eastus-az1", m.PhysicalZone)
	}

	// A region without availabilityZoneMappings must unmarshal to an empty slice,
	// which fetchZoneMappings skips.
	if len(resp.Value[1].AvailabilityZoneMappings) != 0 {
		t.Errorf("Value[1] zone mappings len = %d, want 0", len(resp.Value[1].AvailabilityZoneMappings))
	}
}

func TestParseZoneMappings(t *testing.T) {
	body := []byte(`{
		"value": [
			{
				"name": "eastus",
				"displayName": "East US",
				"availabilityZoneMappings": [
					{"logicalZone": "1", "physicalZone": "eastus-az1"},
					{"logicalZone": "2", "physicalZone": "eastus-az2"}
				]
			},
			{
				"name": "westus",
				"displayName": "West US"
			},
			{
				"name": "centralus",
				"availabilityZoneMappings": [
					{"logicalZone": "1"}
				]
			}
		]
	}`)

	results, err := parseZoneMappings(body, "sub-id", "Sub Name")
	if err != nil {
		t.Fatalf("parseZoneMappings returned error: %v", err)
	}

	// eastus contributes 2 rows, westus 0 (no mappings), centralus 1 row.
	if len(results) != 3 {
		t.Fatalf("results len = %d, want 3", len(results))
	}

	first := results[0]
	if first.subscriptionID != "sub-id" || first.subscriptionName != "Sub Name" {
		t.Errorf("subscription fields = (%q,%q), want (sub-id, Sub Name)", first.subscriptionID, first.subscriptionName)
	}
	if first.location != "eastus" || first.displayName != "East US" {
		t.Errorf("location/displayName = (%q,%q), want (eastus, East US)", first.location, first.displayName)
	}
	if first.logicalZone != "1" || first.physicalZone != "eastus-az1" {
		t.Errorf("zones = (%q,%q), want (1, eastus-az1)", first.logicalZone, first.physicalZone)
	}

	// centralus row has a nil physicalZone, which must normalize to "".
	last := results[2]
	if last.location != "centralus" || last.displayName != "" {
		t.Errorf("centralus row = (%q,%q), want (centralus, \"\")", last.location, last.displayName)
	}
	if last.logicalZone != "1" || last.physicalZone != "" {
		t.Errorf("centralus zones = (%q,%q), want (1, \"\")", last.logicalZone, last.physicalZone)
	}
}

func TestParseZoneMappings_InvalidJSON(t *testing.T) {
	if _, err := parseZoneMappings([]byte("{not json"), "sub", "Sub"); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseZoneMappings_Empty(t *testing.T) {
	results, err := parseZoneMappings([]byte(`{"value": []}`), "sub", "Sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results len = %d, want 0", len(results))
	}
}
