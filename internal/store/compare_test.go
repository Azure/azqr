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

func TestCompareSessionsDetectsResourceAndRecommendationChanges(t *testing.T) {
	s := openTestStore(t)

	vmID := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"
	diskID := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Compute/disks/disk1"
	newID := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm2"

	// Session A: vm1 and disk1, one recommendation hitting both.
	sessA, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(A) error = %v", err)
	}
	recA := NewRecorder(s, sessA.ID, "test")
	resourcesA := &renderers.ReportData{
		Resources: []*models.Resource{
			{ID: vmID, SubscriptionID: "sub1", ResourceGroup: "rg1", Type: "microsoft.compute/virtualmachines", Name: "vm1"},
			{ID: diskID, SubscriptionID: "sub1", ResourceGroup: "rg1", Type: "microsoft.compute/disks", Name: "disk1"},
		},
	}
	recA.StageCompleted(&pipeline.ScanContext{ReportData: resourcesA}, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)
	graphA := &renderers.ReportData{
		Graph: []*models.GraphResult{
			{RecommendationID: "rec-1", ResourceType: "microsoft.compute/virtualmachines", Recommendation: "Do the thing",
				ResourceID: vmID, SubscriptionID: "sub1", ResourceGroup: "rg1", Name: "vm1", Category: models.CategorySecurity, Impact: models.ImpactHigh},
			{RecommendationID: "rec-1", ResourceType: "microsoft.compute/virtualmachines", Recommendation: "Do the thing",
				ResourceID: diskID, SubscriptionID: "sub1", ResourceGroup: "rg1", Name: "disk1", Category: models.CategorySecurity, Impact: models.ImpactHigh},
		},
	}
	recA.StageCompleted(&pipeline.ScanContext{ReportData: graphA}, "Graph Scan", 2, pipeline.StageStatusExecuted, time.Millisecond, nil)
	if err := s.FinalizeSession(sessA.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession(A) error = %v", err)
	}

	// Session B: vm1 still present (now also hit by rec-1), disk1 removed, vm2 added; rec-1 now
	// only hits vm1 (impacted count drops from 2 to 1).
	sessB, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(B) error = %v", err)
	}
	recB := NewRecorder(s, sessB.ID, "test")
	resourcesB := &renderers.ReportData{
		Resources: []*models.Resource{
			{ID: vmID, SubscriptionID: "sub1", ResourceGroup: "rg1", Type: "microsoft.compute/virtualmachines", Name: "vm1"},
			{ID: newID, SubscriptionID: "sub1", ResourceGroup: "rg1", Type: "microsoft.compute/virtualmachines", Name: "vm2"},
		},
	}
	recB.StageCompleted(&pipeline.ScanContext{ReportData: resourcesB}, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)
	graphB := &renderers.ReportData{
		Graph: []*models.GraphResult{
			{RecommendationID: "rec-1", ResourceType: "microsoft.compute/virtualmachines", Recommendation: "Do the thing",
				ResourceID: vmID, SubscriptionID: "sub1", ResourceGroup: "rg1", Name: "vm1", Category: models.CategorySecurity, Impact: models.ImpactHigh},
			{RecommendationID: "rec-2", ResourceType: "microsoft.compute/virtualmachines", Recommendation: "Do another thing",
				ResourceID: newID, SubscriptionID: "sub1", ResourceGroup: "rg1", Name: "vm2", Category: models.CategoryGovernance, Impact: models.ImpactMedium},
		},
	}
	recB.StageCompleted(&pipeline.ScanContext{ReportData: graphB}, "Graph Scan", 2, pipeline.StageStatusExecuted, time.Millisecond, nil)
	if err := s.FinalizeSession(sessB.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession(B) error = %v", err)
	}

	result, err := s.CompareSessions(sessA.ID, sessB.ID)
	if err != nil {
		t.Fatalf("CompareSessions() error = %v", err)
	}

	var added, removed []ResourceChange
	for _, rc := range result.ResourceChanges {
		switch rc.ChangeType {
		case "added":
			added = append(added, rc)
		case "removed":
			removed = append(removed, rc)
		default:
			t.Errorf("unexpected ResourceChange.ChangeType = %q", rc.ChangeType)
		}
	}
	if len(added) != 1 || added[0].ResourceID != newID {
		t.Errorf("added resource changes = %+v, want only %s", added, newID)
	}
	if len(removed) != 1 || removed[0].ResourceID != diskID {
		t.Errorf("removed resource changes = %+v, want only %s", removed, diskID)
	}

	var rec1Change, rec2Change *RecommendationChange
	for i, rc := range result.RecommendationChanges {
		switch rc.RecommendationID {
		case "rec-1":
			rec1Change = &result.RecommendationChanges[i]
		case "rec-2":
			rec2Change = &result.RecommendationChanges[i]
		}
	}
	if rec1Change == nil {
		t.Fatal("rec-1 change not found, want a changed entry (impacted count 2 -> 1)")
	}
	if rec1Change.ChangeType != "changed" || rec1Change.OldImpactedResources != 2 || rec1Change.NewImpactedResources != 1 || rec1Change.ImpactedResourcesDiff != -1 {
		t.Errorf("rec-1 change = %+v, want changed 2 -> 1 (diff -1)", rec1Change)
	}
	if rec2Change == nil {
		t.Fatal("rec-2 change not found, want an added entry")
	}
	if rec2Change.ChangeType != "added" || rec2Change.OldImpactedResources != 0 || rec2Change.NewImpactedResources != 1 {
		t.Errorf("rec-2 change = %+v, want added 0 -> 1", rec2Change)
	}
}

func TestCompareSessionsMissingSessionReturnsError(t *testing.T) {
	s := openTestStore(t)
	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if _, err := s.CompareSessions("does-not-exist", sess.ID); err == nil {
		t.Error("CompareSessions(missing, valid) error = nil, want not-found error")
	}
	if _, err := s.CompareSessions(sess.ID, "does-not-exist"); err == nil {
		t.Error("CompareSessions(valid, missing) error = nil, want not-found error")
	}
}

// TestCompareSessionsRejectsScopeMismatch confirms CompareSessions hard-blocks a comparison
// between two sessions whose scope_id differs (different declared scan scope), without
// performing any diff.
func TestCompareSessionsRejectsScopeMismatch(t *testing.T) {
	s := openTestStore(t)

	sessA, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(A) error = %v", err)
	}
	sessB, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, SessionParams{Subscriptions: []string{"sub-2"}}, "test")
	if err != nil {
		t.Fatalf("CreateSession(B) error = %v", err)
	}

	if _, err := s.CompareSessions(sessA.ID, sessB.ID); err == nil {
		t.Error("CompareSessions(scope mismatch) error = nil, want scope mismatch error")
	}
}

// TestCompareSessionsRejectsTenantMismatch confirms CompareSessions hard-blocks a comparison
// between two sessions whose tenant_id differs. The test harness does not run a live
// Initialization stage, so tenant_id is set directly for both sessions.
func TestCompareSessionsRejectsTenantMismatch(t *testing.T) {
	s := openTestStore(t)

	sessA, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(A) error = %v", err)
	}
	sessB, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(B) error = %v", err)
	}

	if _, err := s.db.Exec(`UPDATE sessions SET tenant_id = ? WHERE id = ?`, "tenant-a", sessA.ID); err != nil {
		t.Fatalf("set tenant_id(A) error = %v", err)
	}
	if _, err := s.db.Exec(`UPDATE sessions SET tenant_id = ? WHERE id = ?`, "tenant-b", sessB.ID); err != nil {
		t.Fatalf("set tenant_id(B) error = %v", err)
	}

	if _, err := s.CompareSessions(sessA.ID, sessB.ID); err == nil {
		t.Error("CompareSessions(tenant mismatch) error = nil, want tenant mismatch error")
	}
}

// TestCompareSessionsIgnoresResourceIDCasing reproduces a real-world scenario where the same
// resource is returned with different segment casing (e.g. resource group casing) across two
// separate scans. Azure Resource Manager IDs are case-insensitive, so a casing-only difference
// must not be reported as the resource being both added and removed.
func TestCompareSessionsIgnoresResourceIDCasing(t *testing.T) {
	s := openTestStore(t)

	lowerID := "/subscriptions/sub1/resourcegroups/rg-sre/providers/microsoft.compute/virtualmachines/vm-api"
	upperID := "/subscriptions/sub1/resourceGroups/RG-SRE/providers/Microsoft.Compute/virtualMachines/VM-API"

	sessA, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(A) error = %v", err)
	}
	recA := NewRecorder(s, sessA.ID, "test")
	recA.StageCompleted(&pipeline.ScanContext{ReportData: &renderers.ReportData{
		Resources: []*models.Resource{
			{ID: upperID, SubscriptionID: "sub1", ResourceGroup: "RG-SRE", Type: "Microsoft.Compute/virtualMachines", Name: "VM-API"},
		},
	}}, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)
	if err := s.FinalizeSession(sessA.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession(A) error = %v", err)
	}

	sessB, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession(B) error = %v", err)
	}
	recB := NewRecorder(s, sessB.ID, "test")
	recB.StageCompleted(&pipeline.ScanContext{ReportData: &renderers.ReportData{
		Resources: []*models.Resource{
			{ID: lowerID, SubscriptionID: "sub1", ResourceGroup: "rg-sre", Type: "microsoft.compute/virtualmachines", Name: "vm-api"},
		},
	}}, "Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)
	if err := s.FinalizeSession(sessB.ID, SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession(B) error = %v", err)
	}

	result, err := s.CompareSessions(sessA.ID, sessB.ID)
	if err != nil {
		t.Fatalf("CompareSessions() error = %v", err)
	}
	if len(result.ResourceChanges) != 0 {
		t.Errorf("ResourceChanges = %+v, want none (casing-only difference)", result.ResourceChanges)
	}
}
