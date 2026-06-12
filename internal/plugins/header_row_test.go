// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"testing"
)

func TestPluginMetadata_HeaderRow_ReturnsColumnNames(t *testing.T) {
	m := PluginMetadata{
		ColumnMetadata: []ColumnMetadata{
			{Name: "Subscription", DataKey: "subscription", FilterType: FilterTypeSearch},
			{Name: "Location", DataKey: "location", FilterType: FilterTypeDropdown},
			{Name: "Display Name", DataKey: "displayName", FilterType: FilterTypeDropdown},
		},
	}

	got := m.HeaderRow()

	want := []string{"Subscription", "Location", "Display Name"}
	if len(got) != len(want) {
		t.Fatalf("HeaderRow() returned %d elements, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("HeaderRow()[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestPluginMetadata_HeaderRow_EmptyColumnMetadata(t *testing.T) {
	m := PluginMetadata{}
	got := m.HeaderRow()
	if len(got) != 0 {
		t.Errorf("HeaderRow() on empty ColumnMetadata = %v, want []", got)
	}
}

func TestPluginMetadata_HeaderRow_PreservesOrder(t *testing.T) {
	names := []string{"Z", "A", "M", "B"}
	cols := make([]ColumnMetadata, len(names))
	for i, n := range names {
		cols[i] = ColumnMetadata{Name: n}
	}
	m := PluginMetadata{ColumnMetadata: cols}

	got := m.HeaderRow()
	for i, want := range names {
		if got[i] != want {
			t.Errorf("HeaderRow()[%d] = %q, want %q (order must be preserved)", i, got[i], want)
		}
	}
}

// TestPluginMetadata_HeaderRow_DoesNotAliasSlice verifies that HeaderRow returns a
// fresh slice each call so callers cannot mutate ColumnMetadata through the result.
func TestPluginMetadata_HeaderRow_DoesNotAliasSlice(t *testing.T) {
	m := PluginMetadata{
		ColumnMetadata: []ColumnMetadata{
			{Name: "Original"},
		},
	}

	row1 := m.HeaderRow()
	row1[0] = "Mutated"

	row2 := m.HeaderRow()
	if row2[0] != "Original" {
		t.Errorf("HeaderRow is aliasing internal state: second call returned %q, want %q", row2[0], "Original")
	}
}
