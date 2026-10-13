// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "azqr.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestCreateSessionInsertsRunningRow(t *testing.T) {
	s := openTestStore(t)

	params := NewSessionParams(&models.ScanParams{
		Subscriptions: []string{"sub-1"},
		Mask:          true,
	})

	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, params, "v1.2.3")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if sess.ID == "" {
		t.Fatal("CreateSession() returned empty ID")
	}
	if sess.Status != SessionStatusRunning {
		t.Errorf("Status = %q, want %q", sess.Status, SessionStatusRunning)
	}

	var status string
	if err := s.db.QueryRow(`SELECT status FROM sessions WHERE id = ?`, sess.ID).Scan(&status); err != nil {
		t.Fatalf("query inserted session: %v", err)
	}
	if status != string(SessionStatusRunning) {
		t.Errorf("stored status = %q, want %q", status, SessionStatusRunning)
	}
}

func TestCreateSessionNilParams(t *testing.T) {
	s := openTestStore(t)

	sess, err := s.CreateSession(SessionKindPluginOnly, SessionTriggerMCP, NewSessionParams(nil), "")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if sess.Kind != SessionKindPluginOnly {
		t.Errorf("Kind = %q, want %q", sess.Kind, SessionKindPluginOnly)
	}
}

func TestFinalizeSessionSetsTerminalState(t *testing.T) {
	s := openTestStore(t)

	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "v1.0.0")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	if err := s.FinalizeSession(sess.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}

	var status string
	var finishedAt, errMsg *string
	var durationMs *int64
	err = s.db.QueryRow(
		`SELECT status, finished_at, duration_ms, error_message FROM sessions WHERE id = ?`,
		sess.ID,
	).Scan(&status, &finishedAt, &durationMs, &errMsg)
	if err != nil {
		t.Fatalf("query finalized session: %v", err)
	}
	if status != string(SessionStatusCompleted) {
		t.Errorf("status = %q, want %q", status, SessionStatusCompleted)
	}
	if finishedAt == nil {
		t.Error("finished_at is nil, want set")
	}
	if durationMs == nil || *durationMs < 0 {
		t.Errorf("duration_ms = %v, want non-negative", durationMs)
	}
	if errMsg != nil {
		t.Errorf("error_message = %q, want nil", *errMsg)
	}
}

func TestFinalizeSessionRecordsErrorMessage(t *testing.T) {
	s := openTestStore(t)

	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "v1.0.0")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if err := s.FinalizeSession(sess.ID, SessionStatusFailed, "graph stage: boom"); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}

	var errMsg string
	if err := s.db.QueryRow(`SELECT error_message FROM sessions WHERE id = ?`, sess.ID).Scan(&errMsg); err != nil {
		t.Fatalf("query error_message: %v", err)
	}
	if errMsg != "graph stage: boom" {
		t.Errorf("error_message = %q, want %q", errMsg, "graph stage: boom")
	}
}

func TestFinalizeSessionUnknownID(t *testing.T) {
	s := openTestStore(t)

	err := s.FinalizeSession("unknown-session", SessionStatusFailed, "")
	if err == nil {
		t.Fatal("FinalizeSession() error = nil, want not-found error")
	}
}

func TestSaveAndLoadSessionSummaryRoundTrips(t *testing.T) {
	s := openTestStore(t)

	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "v1.0.0")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	summary := &findings.Summary{
		Resources:            10,
		ImpactedResources:    4,
		RecommendationsFound: 2,
		ImpactedByImpact:     map[string]int{"High": 3, "Medium": 1},
		ImpactedByCategory:   map[string]int{"Security": 4},
		Recommendations: []findings.Recommendation{
			{ID: "rec-1", ImpactedResources: 3},
		},
	}
	if err := s.SaveSessionSummary(sess.ID, summary); err != nil {
		t.Fatalf("SaveSessionSummary() error = %v", err)
	}

	loaded, err := s.LoadSessionSummary(sess.ID)
	if err != nil {
		t.Fatalf("LoadSessionSummary() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadSessionSummary() = nil, want persisted summary")
	}
	if loaded.Resources != 10 || loaded.ImpactedResources != 4 || loaded.RecommendationsFound != 2 {
		t.Errorf("LoadSessionSummary() = %+v, want matching counts from saved summary", loaded)
	}
	if len(loaded.Recommendations) != 1 || loaded.Recommendations[0].ID != "rec-1" {
		t.Errorf("LoadSessionSummary().Recommendations = %+v, want one entry rec-1", loaded.Recommendations)
	}
}

func TestLoadSessionSummaryReturnsNilWhenNotPersisted(t *testing.T) {
	s := openTestStore(t)

	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "v1.0.0")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	loaded, err := s.LoadSessionSummary(sess.ID)
	if err != nil {
		t.Fatalf("LoadSessionSummary() error = %v, want nil (summary simply not available yet)", err)
	}
	if loaded != nil {
		t.Errorf("LoadSessionSummary() = %+v, want nil", loaded)
	}
}

func TestLoadSessionSummaryMissingSessionReturnsError(t *testing.T) {
	s := openTestStore(t)

	_, err := s.LoadSessionSummary("does-not-exist")
	if err == nil {
		t.Fatal("LoadSessionSummary() error = nil, want not-found error")
	}
}

func TestDeriveStatus(t *testing.T) {
	cases := []struct {
		name           string
		pipelineErr    error
		stagesExecuted int
		want           SessionStatus
	}{
		{"no error", nil, 5, SessionStatusCompleted},
		{"error on first stage", errors.New("boom"), 1, SessionStatusFailed},
		{"error with nothing executed", errors.New("boom"), 0, SessionStatusFailed},
		{"error after earlier stages succeeded", errors.New("boom"), 3, SessionStatusPartial},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DeriveStatus(tc.pipelineErr, tc.stagesExecuted)
			if got != tc.want {
				t.Errorf("DeriveStatus(%v, %d) = %q, want %q", tc.pipelineErr, tc.stagesExecuted, got, tc.want)
			}
		})
	}
}

func TestNewSessionParamsCapturesScope(t *testing.T) {
	stages := models.NewStageConfigsWithDefaults()
	params := &models.ScanParams{
		Subscriptions:    []string{"sub-1", "sub-2"},
		ManagementGroups: []string{"mg-1"},
		ResourceGroups:   []string{"rg-1"},
		ScannerKeys:      []string{"vm"},
		OutputName:       "report",
		Mask:             true,
		Xlsx:             true,
		Stages:           stages,
	}

	sp := NewSessionParams(params)

	if len(sp.Subscriptions) != 2 {
		t.Errorf("Subscriptions = %v, want 2 entries", sp.Subscriptions)
	}
	if !sp.Mask || !sp.Xlsx {
		t.Errorf("Mask/Xlsx not captured: %+v", sp)
	}
	if len(sp.EnabledStages) == 0 {
		t.Error("EnabledStages is empty, want default-enabled stages")
	}
}

// TestScopeIDDeterministicRegardlessOfOrder confirms NFR-001: two SessionParams values with the
// same logical scope produce the same scope_id regardless of slice element order.
func TestScopeIDDeterministicRegardlessOfOrder(t *testing.T) {
	a := SessionParams{
		ManagementGroups: []string{"mg-1", "mg-2"},
		Subscriptions:    []string{"sub-1", "sub-2"},
		ResourceGroups:   []string{"rg-1", "rg-2"},
		ScannerKeys:      []string{"vm", "disk"},
		EnabledStages:    []string{"Resource Discovery", "Graph Scan"},
	}
	b := SessionParams{
		ManagementGroups: []string{"mg-2", "mg-1"},
		Subscriptions:    []string{"sub-2", "sub-1"},
		ResourceGroups:   []string{"rg-2", "rg-1"},
		ScannerKeys:      []string{"disk", "vm"},
		EnabledStages:    []string{"Graph Scan", "Resource Discovery"},
	}

	if ScopeID(a) != ScopeID(b) {
		t.Errorf("ScopeID(a) = %q, ScopeID(b) = %q, want equal regardless of slice order", ScopeID(a), ScopeID(b))
	}
	if ScopeID(a) == "" {
		t.Error("ScopeID(a) is empty, want a non-empty hash")
	}

	c := SessionParams{Subscriptions: []string{"sub-3"}}
	if ScopeID(a) == ScopeID(c) {
		t.Error("ScopeID(a) == ScopeID(c), want different hashes for different scopes")
	}
}
