// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"os"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/renderers"
)

// testPostgresDSNEnvVar names the environment variable that opts a developer or CI run into
// the Postgres-backed tests in this file. It is deliberately distinct from AZQR_POSTGRESQL (the
// production env var Store.Open reads) so that running `go test` never accidentally points a
// test run at a real operator-configured Postgres store: the two must be set independently.
const testPostgresDSNEnvVar = "AZQR_TEST_POSTGRESQL"

// openTestPostgresStore opens a Store against the Postgres instance named by
// AZQR_TEST_POSTGRESQL, or skips the calling test when that variable is unset/empty. This is
// the default, hermetic-by-default testing bar for dialect-sensitive store behavior: it keeps
// `go test ./...` fast and requiring no external services by default, while still exercising the
// Postgres path end-to-end whenever a developer or CI job opts in by setting the DSN.
func openTestPostgresStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv(testPostgresDSNEnvVar)
	if dsn == "" {
		t.Skipf("skipping Postgres-backed test: set %s to a Postgres connection string to run it", testPostgresDSNEnvVar)
	}

	s, err := openPostgres(dsn)
	if err != nil {
		t.Fatalf("openPostgres(%s) error = %v", testPostgresDSNEnvVar, err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return s
}

func TestPostgresOpenAppliesMigrations(t *testing.T) {
	s := openTestPostgresStore(t)

	if s.dialect != dialectPostgres {
		t.Fatalf("dialect = %q, want %q", s.dialect, dialectPostgres)
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
		err := s.db.QueryRow(
			`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`, table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", table, err)
		}
	}

	var scopeCol, tenantCol string
	err := s.db.QueryRow(
		`SELECT column_name FROM information_schema.columns WHERE table_name = 'sessions' AND column_name = 'scope_id'`,
	).Scan(&scopeCol)
	if err != nil {
		t.Errorf("sessions.scope_id column not found after migration (00001_init.sql): %v", err)
	}
	err = s.db.QueryRow(
		`SELECT column_name FROM information_schema.columns WHERE table_name = 'sessions' AND column_name = 'tenant_id'`,
	).Scan(&tenantCol)
	if err != nil {
		t.Errorf("sessions.tenant_id column not found after migration (00001_init.sql): %v", err)
	}
}

// TestPostgresSessionLifecycle exercises CreateSession/FinalizeSession/GetSession/ListSessions/
// DeleteSession against Postgres, verifying the `?`-placeholder queries in session.go and
// query.go that P01-T03 rebound for Postgres behave like their SQLite counterparts covered
// elsewhere in this package (session_test.go, query_test.go).
func TestPostgresSessionLifecycle(t *testing.T) {
	s := openTestPostgresStore(t)

	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test-version")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	got, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.ID != sess.ID {
		t.Errorf("GetSession().ID = %q, want %q", got.ID, sess.ID)
	}

	if err := s.FinalizeSession(sess.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}

	finalized, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession() after finalize error = %v", err)
	}
	if finalized.Status != SessionStatusCompleted {
		t.Errorf("Status = %q, want %q", finalized.Status, SessionStatusCompleted)
	}
	if finalized.FinishedAt == nil {
		t.Error("FinishedAt is nil after FinalizeSession")
	}

	sessions, err := s.ListSessions(SessionFilter{})
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	found := false
	for _, listed := range sessions {
		if listed.ID == sess.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListSessions() did not include session %q", sess.ID)
	}

	if err := s.DeleteSession(sess.ID); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := s.GetSession(sess.ID); err == nil {
		t.Error("GetSession() after DeleteSession() returned no error, want not-found error")
	}
}

// TestPostgresRecorderWritesPluginResults exercises Recorder's plugin_result_rows writes
// against Postgres. migrations/00001_init.sql intentionally drops its SQLite-only
// `CHECK (json_valid(data))` constraint on this column for Postgres portability (P01-T02), so
// this test guards that recorder.go's rebound INSERT/DELETE statements still round-trip data
// correctly without that defense-in-depth check.
func TestPostgresRecorderWritesPluginResults(t *testing.T) {
	s := openTestPostgresStore(t)
	sess, err := s.CreateSession(SessionKindPluginOnly, SessionTriggerCLI, NewSessionParams(nil), "test-version")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	rec := NewRecorder(s, sess.ID, "test-version")
	rd := &renderers.ReportData{
		PluginResults: []*renderers.PluginResult{
			{PluginName: "demo", SheetName: "Sheet1", Table: [][]string{{"col_a", "col_b"}, {"v1", "v2"}}},
		},
	}
	ctx := &pipeline.ScanContext{ReportData: rd}

	rec.StageCompleted(ctx, "Plugin Execution", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)

	loaded, err := s.LoadReportData(sess.ID, "out.xlsx", false)
	if err != nil {
		t.Fatalf("LoadReportData() error = %v", err)
	}
	if len(loaded.PluginResults) != 1 {
		t.Fatalf("len(PluginResults) = %d, want 1", len(loaded.PluginResults))
	}
	got := loaded.PluginResults[0]
	if got.PluginName != "demo" || got.SheetName != "Sheet1" {
		t.Errorf("PluginResult = %+v, want PluginName=demo SheetName=Sheet1", got)
	}
	if len(got.Table) != 2 {
		t.Fatalf("len(Table) = %d, want 2", len(got.Table))
	}
}
