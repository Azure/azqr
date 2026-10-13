// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/renderers"
)

func TestListSessionsFiltersAndOrdersNewestFirst(t *testing.T) {
	s := openTestStore(t)

	sessFull, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if err := s.FinalizeSession(sessFull.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}

	time.Sleep(2 * time.Millisecond) // ensure a distinct started_at ordering
	sessPlugin, err := s.CreateSession(SessionKindPluginOnly, SessionTriggerMCP, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if err := s.FinalizeSession(sessPlugin.ID, SessionStatusFailed, "boom"); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}

	all, err := s.ListSessions(SessionFilter{})
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}
	if all[0].ID != sessPlugin.ID || all[1].ID != sessFull.ID {
		t.Errorf("ListSessions() order = [%s, %s], want newest first [%s, %s]", all[0].ID, all[1].ID, sessPlugin.ID, sessFull.ID)
	}

	byStatus, err := s.ListSessions(SessionFilter{Status: SessionStatusFailed})
	if err != nil {
		t.Fatalf("ListSessions(Status=failed) error = %v", err)
	}
	if len(byStatus) != 1 || byStatus[0].ID != sessPlugin.ID {
		t.Errorf("ListSessions(Status=failed) = %+v, want only %s", byStatus, sessPlugin.ID)
	}

	byKind, err := s.ListSessions(SessionFilter{Kind: SessionKindFull})
	if err != nil {
		t.Fatalf("ListSessions(Kind=full) error = %v", err)
	}
	if len(byKind) != 1 || byKind[0].ID != sessFull.ID {
		t.Errorf("ListSessions(Kind=full) = %+v, want only %s", byKind, sessFull.ID)
	}

	limited, err := s.ListSessions(SessionFilter{Limit: 1})
	if err != nil {
		t.Fatalf("ListSessions(Limit=1) error = %v", err)
	}
	if len(limited) != 1 || limited[0].ID != sessPlugin.ID {
		t.Errorf("ListSessions(Limit=1) = %+v, want only newest %s", limited, sessPlugin.ID)
	}

	offset, err := s.ListSessions(SessionFilter{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListSessions(Limit=1,Offset=1) error = %v", err)
	}
	if len(offset) != 1 || offset[0].ID != sessFull.ID {
		t.Errorf("ListSessions(Limit=1,Offset=1) = %+v, want only %s", offset, sessFull.ID)
	}
}

func TestGetSessionReturnsErrorWhenMissing(t *testing.T) {
	s := openTestStore(t)
	if _, err := s.GetSession("does-not-exist"); err == nil {
		t.Error("GetSession() error = nil, want not-found error")
	}
}

func TestGetSessionRoundTripsFields(t *testing.T) {
	s := openTestStore(t)
	params := NewSessionParams(&models.ScanParams{
		Subscriptions: []string{"sub1"},
		ScannerKeys:   []string{"vm"},
		Mask:          true,
	})
	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, params, "1.2.3")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if err := s.FinalizeSession(sess.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}

	got, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.Kind != SessionKindFull || got.Trigger != SessionTriggerCLI || got.Status != SessionStatusCompleted {
		t.Errorf("GetSession() = %+v, want kind/trigger/status full/cli/completed", got)
	}
	if got.AzqrVersion != "1.2.3" {
		t.Errorf("GetSession().AzqrVersion = %q, want 1.2.3", got.AzqrVersion)
	}
	if len(got.Params.Subscriptions) != 1 || got.Params.Subscriptions[0] != "sub1" {
		t.Errorf("GetSession().Params.Subscriptions = %v, want [sub1]", got.Params.Subscriptions)
	}
	if got.FinishedAt == nil || got.DurationMs == nil {
		t.Error("GetSession() FinishedAt/DurationMs = nil, want populated after finalize")
	}
}

func TestListStageRunsOrdersByPosition(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rec.StageCompleted(nil, "Initialization", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)
	rec.StageCompleted(nil, "Graph Scan", 2, pipeline.StageStatusFailed, 2*time.Millisecond, errFake)

	runs, err := s.ListStageRuns(sess.ID)
	if err != nil {
		t.Fatalf("ListStageRuns() error = %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("len(runs) = %d, want 2", len(runs))
	}
	if runs[0].StageName != "Initialization" || runs[1].StageName != "Graph Scan" {
		t.Errorf("ListStageRuns() order = [%s, %s], want [Initialization, Graph Scan]", runs[0].StageName, runs[1].StageName)
	}
	if runs[1].ErrorMessage != errFake.Error() {
		t.Errorf("runs[1].ErrorMessage = %q, want %q", runs[1].ErrorMessage, errFake.Error())
	}
}

func TestDeleteSessionCascades(t *testing.T) {
	s := openTestStore(t)
	sess := newTestSession(t, s)
	rec := NewRecorder(s, sess.ID, "test")

	rd := &renderers.ReportData{
		Resources: []*models.Resource{{ID: "res1", SubscriptionID: "sub1", Type: "microsoft.compute/virtualmachines", Name: "vm1"}},
	}
	rec.StageCompleted(&pipeline.ScanContext{ReportData: rd}, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)

	if err := s.DeleteSession(sess.ID); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}

	if _, err := s.GetSession(sess.ID); err == nil {
		t.Error("GetSession() after delete = nil error, want not-found")
	}

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM resources WHERE session_id = ?`, sess.ID).Scan(&count); err != nil {
		t.Fatalf("count resources: %v", err)
	}
	if count != 0 {
		t.Errorf("resources remaining after DeleteSession() = %d, want 0", count)
	}

	var stageCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM session_stage_runs WHERE session_id = ?`, sess.ID).Scan(&stageCount); err != nil {
		t.Fatalf("count session_stage_runs: %v", err)
	}
	if stageCount != 0 {
		t.Errorf("session_stage_runs remaining after DeleteSession() = %d, want 0", stageCount)
	}
}

func TestDeleteSessionMissingReturnsError(t *testing.T) {
	s := openTestStore(t)
	if err := s.DeleteSession("does-not-exist"); err == nil {
		t.Error("DeleteSession() error = nil, want not-found error")
	}
}

func TestLoadReportDataRoundTripsAllDomains(t *testing.T) {
	s := openTestStore(t)
	params := NewSessionParams(&models.ScanParams{Stages: models.NewStageConfigsWithDefaults()})
	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, params, "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	rec := NewRecorder(s, sess.ID, "test")

	resourceID := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"
	rd := &renderers.ReportData{
		Resources: []*models.Resource{
			{ID: resourceID, SubscriptionID: "sub1", ResourceGroup: "rg1", Type: "microsoft.compute/virtualmachines", Name: "vm1", Tags: map[string]string{"env": "prod"}},
		},
	}
	rec.StageCompleted(&pipeline.ScanContext{ReportData: rd}, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)

	graphRD := &renderers.ReportData{
		Graph: []*models.GraphResult{
			{
				RecommendationID: "rec-1", ResourceType: "microsoft.compute/virtualmachines",
				Recommendation: "Do the thing", LongDescription: "Long", PotentialBenefits: "Benefits",
				ResourceID: resourceID, SubscriptionID: "sub1", SubscriptionName: "Sub One",
				ResourceGroup: "rg1", Name: "vm1", Category: models.CategorySecurity, Impact: models.ImpactHigh,
				Learn: "https://example.test", Source: "APRL",
			},
		},
	}
	rec.StageCompleted(&pipeline.ScanContext{ReportData: graphRD}, "Graph Scan", 2, pipeline.StageStatusExecuted, time.Millisecond, nil)

	advisorRD := &renderers.ReportData{
		Advisor: []*models.AdvisorResult{{RecommendationID: "adv-1", ResourceID: resourceID, SubscriptionID: "sub1", SubscriptionName: "Sub One", Type: "microsoft.compute/virtualmachines", Name: "vm1", Category: "Cost", Impact: "High", Description: "desc"}},
	}
	rec.StageCompleted(&pipeline.ScanContext{ReportData: advisorRD}, "Advisor Scan", 3, pipeline.StageStatusExecuted, time.Millisecond, nil)

	defenderRD := &renderers.ReportData{
		Defender: []*models.DefenderResult{{SubscriptionID: "sub1", SubscriptionName: "Sub One", Name: "VirtualMachines", Tier: "Standard"}},
	}
	rec.StageCompleted(&pipeline.ScanContext{ReportData: defenderRD}, "Defender Status Scan", 4, pipeline.StageStatusExecuted, time.Millisecond, nil)

	pluginRD := &renderers.ReportData{
		PluginResults: []*renderers.PluginResult{
			{PluginName: "region-selection", SheetName: "Regions", Table: [][]string{{"Region", "Score"}, {"eastus", "10"}}},
		},
	}
	rec.StageCompleted(&pipeline.ScanContext{ReportData: pluginRD}, "Plugin Execution", 5, pipeline.StageStatusExecuted, time.Millisecond, nil)

	loaded, err := s.LoadReportData(sess.ID, "out", true)
	if err != nil {
		t.Fatalf("LoadReportData() error = %v", err)
	}

	if len(loaded.Resources) != 1 || loaded.Resources[0].ID != resourceID {
		t.Errorf("loaded.Resources = %+v, want one resource %s", loaded.Resources, resourceID)
	}
	if loaded.Resources[0].Tags["env"] != "prod" {
		t.Errorf("loaded.Resources[0].Tags = %v, want env=prod", loaded.Resources[0].Tags)
	}
	if len(loaded.Graph) != 1 || loaded.Graph[0].RecommendationID != "rec-1" {
		t.Errorf("loaded.Graph = %+v, want one result rec-1", loaded.Graph)
	}
	if len(loaded.Advisor) != 1 || loaded.Advisor[0].RecommendationID != "adv-1" {
		t.Errorf("loaded.Advisor = %+v, want one result adv-1", loaded.Advisor)
	}
	if len(loaded.Defender) != 1 || loaded.Defender[0].Name != "VirtualMachines" {
		t.Errorf("loaded.Defender = %+v, want one result VirtualMachines", loaded.Defender)
	}
	if len(loaded.PluginResults) != 1 || loaded.PluginResults[0].PluginName != "region-selection" {
		t.Errorf("loaded.PluginResults = %+v, want one result region-selection", loaded.PluginResults)
	}
	if len(loaded.PluginResults[0].Table) != 2 || len(loaded.PluginResults[0].Table[0]) != 2 {
		t.Errorf("loaded.PluginResults[0].Table = %+v, want 2 rows x 2 cols", loaded.PluginResults[0].Table)
	}
	if loaded.Summary == nil || len(loaded.Summary.Recommendations) != 1 {
		t.Errorf("loaded.Summary = %+v, want one recommendation built from catalog", loaded.Summary)
	}
	if loaded.Stages == nil {
		t.Fatal("loaded.Stages = nil, want a non-nil StageConfigs (exporters dereference it unconditionally)")
	}
	if !loaded.Stages.IsStageEnabled(models.StageNameGraph) {
		t.Error("loaded.Stages.IsStageEnabled(graph) = false, want true (graph is always enabled)")
	}
	if len(loaded.ResourceTypeCount) != 1 || loaded.ResourceTypeCount[0].Count != 1 {
		t.Errorf("loaded.ResourceTypeCount = %+v, want one entry with count 1", loaded.ResourceTypeCount)
	}
}

func TestLoadReportDataMissingSessionReturnsError(t *testing.T) {
	s := openTestStore(t)
	if _, err := s.LoadReportData("does-not-exist", "out", true); err == nil {
		t.Error("LoadReportData() error = nil, want not-found error")
	}
}

var errFake = fakeErr("boom")

type fakeErr string

func (e fakeErr) Error() string { return string(e) }
