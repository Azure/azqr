// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
)

func TestAppendReadAndSelect(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	for index, scope := range []string{"a", "b", "a"} {
		record := NewRecord(&findings.Summary{
			Resources:            index + 1,
			ImpactedResources:    index + 2,
			RecommendationsFound: 1,
		}, scope, "test", time.Date(2026, 1, index+1, 0, 0, 0, 0, time.UTC))
		if err := Append(path, record); err != nil {
			t.Fatal(err)
		}
	}

	records, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	selected, scope, err := Select(records, "a", 1)
	if err != nil {
		t.Fatal(err)
	}
	if scope != "a" || len(selected) != 1 || selected[0].Resources != 3 {
		t.Fatalf("unexpected selection: %q %+v", scope, selected)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("history permissions = %v, %v", info.Mode().Perm(), err)
	}
}

func TestReadIgnoresTruncatedTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	valid := `{"schemaVersion":1,"timestamp":"2026-01-01T00:00:00Z","scopeId":"scope"}` + "\n"
	if err := os.WriteFile(path, []byte(valid+`{"schemaVersion":1`), 0600); err != nil {
		t.Fatal(err)
	}
	records, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
}

func TestScopeIDIsOrderIndependent(t *testing.T) {
	first := &models.ScanParams{
		Subscriptions: []string{"B", "a"},
		ScannerKeys:   []string{"vm", "aks"},
		Stages:        models.NewStageConfigsWithDefaults(),
		Filters:       models.NewFilters(),
	}
	second := &models.ScanParams{
		Subscriptions: []string{"A", "b"},
		ScannerKeys:   []string{"aks", "vm"},
		Stages:        models.NewStageConfigsWithDefaults(),
		Filters:       models.NewFilters(),
	}
	firstID, err := ScopeID(first)
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := ScopeID(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstID != secondID {
		t.Fatalf("scope IDs differ: %s != %s", firstID, secondID)
	}
}

func TestRenderTableAndJSON(t *testing.T) {
	records := []Record{{
		SchemaVersion:     SchemaVersion,
		Timestamp:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ScopeID:           "scope",
		Resources:         2,
		ImpactedResources: 1,
		ImpactedByImpact:  map[string]int{"High": 1},
	}}
	table, err := Render(records, "scope", "table")
	if err != nil || !strings.Contains(table, "FINDINGS") {
		t.Fatalf("unexpected table: %q, %v", table, err)
	}
	output, err := Render(records, "scope", "json")
	if err != nil || !strings.Contains(output, `"scopeId": "scope"`) {
		t.Fatalf("unexpected JSON: %q, %v", output, err)
	}
}
