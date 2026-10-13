// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"path/filepath"
	"testing"
)

func TestOpenAppliesMigrationsAndPragmas(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "azqr.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	if s.Path() != dbPath {
		t.Errorf("Path() = %q, want %q", s.Path(), dbPath)
	}

	wantTables := []string{
		"sessions",
		"session_stage_runs",
		"resources",
		"recommendation_catalog",
		"recommendation_results",
		"advisor_results",
		"defender_results",
		"defender_recommendations",
		"azure_policy_results",
		"arc_sql_results",
		"cost_results",
		"plugin_executions",
		"plugin_result_rows",
	}
	for _, table := range wantTables {
		var name string
		err := s.db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", table, err)
		}
	}

	var journalMode string
	if err := s.db.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var foreignKeys int
	if err := s.db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}

	var busyTimeout int
	if err := s.db.QueryRow(`PRAGMA busy_timeout`).Scan(&busyTimeout); err != nil {
		t.Fatalf("query busy_timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busyTimeout)
	}
}

func TestOpenEmptyPathUsesDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	s, err := Open("")
	if err != nil {
		t.Fatalf("Open(\"\") error = %v", err)
	}
	defer func() { _ = s.Close() }()

	wantPath, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if s.Path() != wantPath {
		t.Errorf("Path() = %q, want %q", s.Path(), wantPath)
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "azqr.db")

	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	defer func() { _ = s2.Close() }()
}
