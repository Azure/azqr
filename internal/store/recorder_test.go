// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/renderers"
)

func newTestSession(t *testing.T, s *Store) *Session {
	t.Helper()
	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	return sess
}

func TestRecorderStageCompletedRecordsStageRun(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rec.StageCompleted(nil, "Initialization", 1, pipeline.StageStatusExecuted, 5*time.Millisecond, nil)

	var status string
	var position int
	err := s.db.QueryRow(
		`SELECT status, position FROM session_stage_runs WHERE session_id = ? AND stage_name = ?`,
		sess.ID, "Initialization",
	).Scan(&status, &position)
	if err != nil {
		t.Fatalf("query session_stage_runs: %v", err)
	}
	if status != string(pipeline.StageStatusExecuted) || position != 1 {
		t.Errorf("status/position = %q/%d, want executed/1", status, position)
	}
}

func TestRecorderStageCompletedRecordsFailure(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rec.StageCompleted(nil, "Graph Scan", 2, pipeline.StageStatusFailed, time.Millisecond, errors.New("boom"))

	var status, errMsg string
	err := s.db.QueryRow(
		`SELECT status, error_message FROM session_stage_runs WHERE session_id = ? AND stage_name = ?`,
		sess.ID, "Graph Scan",
	).Scan(&status, &errMsg)
	if err != nil {
		t.Fatalf("query session_stage_runs: %v", err)
	}
	if status != string(pipeline.StageStatusFailed) || errMsg != "boom" {
		t.Errorf("status/error_message = %q/%q, want failed/boom", status, errMsg)
	}
}

func TestRecorderWritesResources(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rd := &renderers.ReportData{
		Resources: []*models.Resource{
			{
				ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				SubscriptionID: "sub1",
				ResourceGroup:  "rg1",
				Type:           "microsoft.compute/virtualmachines",
				Name:           "vm1",
				Tags:           map[string]string{"env": "prod"},
			},
		},
	}
	ctx := &pipeline.ScanContext{ReportData: rd}

	rec.StageCompleted(ctx, "Resource Discovery", 3, pipeline.StageStatusExecuted, time.Millisecond, nil)

	var name, tagsJSON string
	err := s.db.QueryRow(
		`SELECT name, tags_json FROM resources WHERE session_id = ? AND resource_id = ?`,
		sess.ID, rd.Resources[0].ID,
	).Scan(&name, &tagsJSON)
	if err != nil {
		t.Fatalf("query resources: %v", err)
	}
	if name != "vm1" {
		t.Errorf("name = %q, want vm1", name)
	}
	var tags map[string]string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		t.Fatalf("unmarshal tags_json: %v", err)
	}
	if tags["env"] != "prod" {
		t.Errorf("tags[env] = %q, want prod", tags["env"])
	}
}

func TestRecorderWritesResourcesReplacesOnRewrite(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	first := &pipeline.ScanContext{ReportData: &renderers.ReportData{
		Resources: []*models.Resource{{ID: "res-1", Name: "one"}, {ID: "res-2", Name: "two"}},
	}}
	rec.StageCompleted(first, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)

	second := &pipeline.ScanContext{ReportData: &renderers.ReportData{
		Resources: []*models.Resource{{ID: "res-1", Name: "one-renamed"}},
	}}
	rec.StageCompleted(second, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM resources WHERE session_id = ?`, sess.ID).Scan(&count); err != nil {
		t.Fatalf("count resources: %v", err)
	}
	if count != 1 {
		t.Errorf("resource count after rewrite = %d, want 1 (stage replaces wholesale)", count)
	}
}

func TestRecorderWritesGraphResultsAndCatalog(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rd := &renderers.ReportData{
		Graph: []*models.GraphResult{
			{
				RecommendationID: "rec-1",
				ResourceID:       "res-1",
				SubscriptionID:   "sub1",
				Recommendation:   "Do the thing",
				Category:         models.CategoryHighAvailability,
				Impact:           models.ImpactHigh,
				Source:           "APRL",
			},
		},
	}
	ctx := &pipeline.ScanContext{ReportData: rd}
	rec.StageCompleted(ctx, "Graph Scan", 4, pipeline.StageStatusExecuted, time.Millisecond, nil)

	var catalogCount, resultCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM recommendation_catalog WHERE recommendation_id = ?`, "rec-1").Scan(&catalogCount); err != nil {
		t.Fatalf("count catalog: %v", err)
	}
	if catalogCount != 1 {
		t.Errorf("catalog count = %d, want 1", catalogCount)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM recommendation_results WHERE session_id = ?`, sess.ID).Scan(&resultCount); err != nil {
		t.Fatalf("count results: %v", err)
	}
	if resultCount != 1 {
		t.Errorf("result count = %d, want 1", resultCount)
	}
}

func TestRecorderWritesPluginResults(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rd := &renderers.ReportData{
		PluginResults: []*renderers.PluginResult{
			{
				PluginName: "my-plugin",
				SheetName:  "Findings",
				Table: [][]string{
					{"ResourceId", "Finding"},
					{"res-1", "issue A"},
					{"res-2", "issue B"},
				},
			},
		},
	}
	ctx := &pipeline.ScanContext{ReportData: rd}
	rec.StageCompleted(ctx, "Plugin Execution", 5, pipeline.StageStatusExecuted, time.Millisecond, nil)

	var execCount int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM plugin_executions WHERE session_id = ? AND plugin_name = ?`,
		sess.ID, "my-plugin",
	).Scan(&execCount); err != nil {
		t.Fatalf("count plugin_executions: %v", err)
	}
	if execCount != 1 {
		t.Errorf("plugin_executions count = %d, want 1", execCount)
	}

	rows, err := s.db.Query(
		`SELECT row_ordinal, data FROM plugin_result_rows WHERE session_id = ? AND plugin_name = ? ORDER BY row_ordinal`,
		sess.ID, "my-plugin",
	)
	if err != nil {
		t.Fatalf("query plugin_result_rows: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var got []map[string]string
	for rows.Next() {
		var ordinal int
		var data string
		if err := rows.Scan(&ordinal, &data); err != nil {
			t.Fatalf("scan plugin_result_rows: %v", err)
		}
		var doc map[string]string
		if err := json.Unmarshal([]byte(data), &doc); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
		got = append(got, doc)
	}
	if len(got) != 2 {
		t.Fatalf("plugin_result_rows count = %d, want 2", len(got))
	}
	if got[0]["ResourceId"] != "res-1" || got[0]["Finding"] != "issue A" {
		t.Errorf("row 0 = %+v, want ResourceId=res-1 Finding='issue A'", got[0])
	}
}

func TestRunPersistedScanFinalizesCompleted(t *testing.T) {
	s := openTestStore(t)

	wantSummary := &findings.Summary{}
	summary, err := s.RunPersistedScan(
		SessionKindFull, SessionTriggerCLI, &models.ScanParams{}, "test",
		func(scanner *pipeline.Scanner) (*findings.Summary, error) {
			return wantSummary, nil
		},
		nil,
	)
	if err != nil {
		t.Fatalf("RunPersistedScan() error = %v", err)
	}
	if summary != wantSummary {
		t.Error("RunPersistedScan() did not return scanFn's summary")
	}

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE status = ?`, string(SessionStatusCompleted)).Scan(&count); err != nil {
		t.Fatalf("count completed sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("completed session count = %d, want 1", count)
	}
}

func TestRunPersistedScanFinalizesFailed(t *testing.T) {
	s := openTestStore(t)

	wantErr := errors.New("scan boom")
	_, err := s.RunPersistedScan(
		SessionKindFull, SessionTriggerCLI, &models.ScanParams{}, "test",
		func(scanner *pipeline.Scanner) (*findings.Summary, error) {
			return nil, wantErr
		},
		nil,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunPersistedScan() error = %v, want %v", err, wantErr)
	}

	var status, errMsg string
	err = s.db.QueryRow(`SELECT status, error_message FROM sessions LIMIT 1`).Scan(&status, &errMsg)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if status != string(SessionStatusFailed) {
		t.Errorf("status = %q, want failed", status)
	}
	if errMsg != "scan boom" {
		t.Errorf("error_message = %q, want %q", errMsg, "scan boom")
	}
}

// TestRunPersistedScanInvokesOnSessionCreated confirms a caller can learn the new session's ID
// (and other metadata) before the scan itself runs, without duplicating any create/finalize
// logic -- the capability the Wails desktop app's StartScan binding relies on.
func TestRunPersistedScanInvokesOnSessionCreated(t *testing.T) {
	s := openTestStore(t)

	var createdID string
	var scanFnRanAfterCreate bool
	_, err := s.RunPersistedScan(
		SessionKindFull, SessionTriggerCLI, &models.ScanParams{}, "test",
		func(scanner *pipeline.Scanner) (*findings.Summary, error) {
			scanFnRanAfterCreate = createdID != ""
			return &findings.Summary{}, nil
		},
		func(sess *Session) {
			createdID = sess.ID
		},
	)
	if err != nil {
		t.Fatalf("RunPersistedScan() error = %v", err)
	}
	if createdID == "" {
		t.Fatal("onSessionCreated was not called")
	}
	if !scanFnRanAfterCreate {
		t.Error("scanFn ran before onSessionCreated was invoked, want onSessionCreated first")
	}

	got, err := s.GetSession(createdID)
	if err != nil {
		t.Fatalf("GetSession(%q) error = %v", createdID, err)
	}
	if got.ID != createdID {
		t.Errorf("GetSession(%q).ID = %q, want %q", createdID, got.ID, createdID)
	}
}

// TestRunPersistedScanPersistsSummary confirms the scan's already-computed *findings.Summary is
// written once during finalize, so LoadSessionSummary can retrieve it without reconstructing the
// report from the session's full resource/graph rows.
func TestRunPersistedScanPersistsSummary(t *testing.T) {
	s := openTestStore(t)

	wantSummary := &findings.Summary{Resources: 5, ImpactedResources: 2, RecommendationsFound: 1}
	var createdID string
	_, err := s.RunPersistedScan(
		SessionKindFull, SessionTriggerCLI, &models.ScanParams{}, "test",
		func(scanner *pipeline.Scanner) (*findings.Summary, error) {
			return wantSummary, nil
		},
		func(sess *Session) { createdID = sess.ID },
	)
	if err != nil {
		t.Fatalf("RunPersistedScan() error = %v", err)
	}

	loaded, err := s.LoadSessionSummary(createdID)
	if err != nil {
		t.Fatalf("LoadSessionSummary() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadSessionSummary() = nil, want the summary persisted during RunPersistedScan")
	}
	if loaded.Resources != 5 || loaded.ImpactedResources != 2 || loaded.RecommendationsFound != 1 {
		t.Errorf("LoadSessionSummary() = %+v, want matching counts from scanFn's summary", loaded)
	}
}
