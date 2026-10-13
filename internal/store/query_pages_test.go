// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/renderers"
)

// seedResourcesAndGraphResults writes n resources and n graph results for a new session,
// returning the session ID. IDs are zero-padded so lexical (resource_id) ordering matches
// numeric ordering, which LoadResourcesPage/LoadGraphResultsPage rely on.
func seedResourcesAndGraphResults(t *testing.T, s *Store, n int) string {
	t.Helper()

	params := NewSessionParams(&models.ScanParams{Stages: models.NewStageConfigsWithDefaults()})
	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, params, "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	rec := NewRecorder(s, sess.ID, "test")

	resources := make([]*models.Resource, 0, n)
	graph := make([]*models.GraphResult, 0, n)
	for i := 0; i < n; i++ {
		resourceID := fmt.Sprintf("/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm%02d", i)
		resources = append(resources, &models.Resource{
			ID: resourceID, SubscriptionID: "sub1", ResourceGroup: "rg1",
			Type: "microsoft.compute/virtualmachines", Name: fmt.Sprintf("vm%02d", i),
		})
		graph = append(graph, &models.GraphResult{
			RecommendationID: fmt.Sprintf("rec-%02d", i), ResourceType: "microsoft.compute/virtualmachines",
			Recommendation: "Do the thing", ResourceID: resourceID, SubscriptionID: "sub1",
			SubscriptionName: "Sub One", ResourceGroup: "rg1", Name: fmt.Sprintf("vm%02d", i),
			Category: models.CategorySecurity, Impact: models.ImpactHigh, Source: "APRL",
		})
	}

	rec.StageCompleted(&pipeline.ScanContext{ReportData: &renderers.ReportData{Resources: resources}},
		"Resource Discovery", 1, pipeline.StageStatusExecuted, time.Millisecond, nil)
	rec.StageCompleted(&pipeline.ScanContext{ReportData: &renderers.ReportData{Graph: graph}},
		"Graph Scan", 2, pipeline.StageStatusExecuted, time.Millisecond, nil)

	return sess.ID
}

func TestLoadResourcesPagePaginatesDeterministically(t *testing.T) {
	s := openTestStore(t)
	sessionID := seedResourcesAndGraphResults(t, s, 5)

	page1, hasMore, err := s.LoadResourcesPage(sessionID, 2, 0)
	if err != nil {
		t.Fatalf("LoadResourcesPage() error = %v", err)
	}
	if !hasMore {
		t.Errorf("LoadResourcesPage() hasMore = false, want true (5 rows, page size 2)")
	}
	if len(page1) != 2 || page1[0].Name != "vm00" || page1[1].Name != "vm01" {
		t.Errorf("LoadResourcesPage() page1 = %+v, want [vm00, vm01]", page1)
	}

	page3, hasMore, err := s.LoadResourcesPage(sessionID, 2, 4)
	if err != nil {
		t.Fatalf("LoadResourcesPage() error = %v", err)
	}
	if hasMore {
		t.Errorf("LoadResourcesPage() hasMore = true, want false (last page)")
	}
	if len(page3) != 1 || page3[0].Name != "vm04" {
		t.Errorf("LoadResourcesPage() page3 = %+v, want [vm04]", page3)
	}
}

func TestLoadResourcesPageClampsNonPositiveAndOversizedLimit(t *testing.T) {
	s := openTestStore(t)
	sessionID := seedResourcesAndGraphResults(t, s, 3)

	page, _, err := s.LoadResourcesPage(sessionID, 0, 0)
	if err != nil {
		t.Fatalf("LoadResourcesPage() error = %v", err)
	}
	if len(page) != 3 {
		t.Errorf("LoadResourcesPage(limit=0) returned %d rows, want 3 (DefaultResultPageSize applied)", len(page))
	}

	page, _, err = s.LoadResourcesPage(sessionID, MaxResultPageSize+500, 0)
	if err != nil {
		t.Fatalf("LoadResourcesPage() error = %v", err)
	}
	if len(page) != 3 {
		t.Errorf("LoadResourcesPage(limit=MaxResultPageSize+500) returned %d rows, want 3", len(page))
	}
}

// TestLoadResourcesPageAndGraphResultsPageClampNegativeOffset confirms a negative offset behaves
// identically to offset 0 (first page) rather than erroring or producing a driver-dependent
// result, per the review follow-up (RV-002) hardening the paginated loaders for every backend.
func TestLoadResourcesPageAndGraphResultsPageClampNegativeOffset(t *testing.T) {
	s := openTestStore(t)
	sessionID := seedResourcesAndGraphResults(t, s, 5)

	resourcesAtNegative, hasMoreNeg, err := s.LoadResourcesPage(sessionID, 2, -10)
	if err != nil {
		t.Fatalf("LoadResourcesPage(offset=-10) error = %v", err)
	}
	resourcesAtZero, hasMoreZero, err := s.LoadResourcesPage(sessionID, 2, 0)
	if err != nil {
		t.Fatalf("LoadResourcesPage(offset=0) error = %v", err)
	}
	if len(resourcesAtNegative) != len(resourcesAtZero) || hasMoreNeg != hasMoreZero {
		t.Fatalf("LoadResourcesPage(offset=-10) = (%+v, %t), want same as offset=0 (%+v, %t)",
			resourcesAtNegative, hasMoreNeg, resourcesAtZero, hasMoreZero)
	}
	for i := range resourcesAtNegative {
		if resourcesAtNegative[i].Name != resourcesAtZero[i].Name {
			t.Errorf("LoadResourcesPage(offset=-10)[%d].Name = %q, want %q (same as offset=0)",
				i, resourcesAtNegative[i].Name, resourcesAtZero[i].Name)
		}
	}

	graphAtNegative, _, err := s.LoadGraphResultsPage(sessionID, 2, -10)
	if err != nil {
		t.Fatalf("LoadGraphResultsPage(offset=-10) error = %v", err)
	}
	graphAtZero, _, err := s.LoadGraphResultsPage(sessionID, 2, 0)
	if err != nil {
		t.Fatalf("LoadGraphResultsPage(offset=0) error = %v", err)
	}
	if len(graphAtNegative) != 2 || graphAtNegative[0].RecommendationID != graphAtZero[0].RecommendationID {
		t.Errorf("LoadGraphResultsPage(offset=-10) = %+v, want same as offset=0 %+v", graphAtNegative, graphAtZero)
	}
}

func TestClampResultPageOffsetClampsNegativeToZero(t *testing.T) {
	if got := ClampResultPageOffset(-5); got != 0 {
		t.Errorf("ClampResultPageOffset(-5) = %d, want 0", got)
	}
	if got := ClampResultPageOffset(0); got != 0 {
		t.Errorf("ClampResultPageOffset(0) = %d, want 0", got)
	}
	if got := ClampResultPageOffset(42); got != 42 {
		t.Errorf("ClampResultPageOffset(42) = %d, want 42 (unchanged)", got)
	}
}

func TestLoadGraphResultsPagePaginatesDeterministically(t *testing.T) {
	s := openTestStore(t)
	sessionID := seedResourcesAndGraphResults(t, s, 5)

	page1, hasMore, err := s.LoadGraphResultsPage(sessionID, 2, 0)
	if err != nil {
		t.Fatalf("LoadGraphResultsPage() error = %v", err)
	}
	if !hasMore {
		t.Errorf("LoadGraphResultsPage() hasMore = false, want true (5 rows, page size 2)")
	}
	if len(page1) != 2 || page1[0].RecommendationID != "rec-00" || page1[1].RecommendationID != "rec-01" {
		t.Errorf("LoadGraphResultsPage() page1 = %+v, want [rec-00, rec-01]", page1)
	}

	page3, hasMore, err := s.LoadGraphResultsPage(sessionID, 2, 4)
	if err != nil {
		t.Fatalf("LoadGraphResultsPage() error = %v", err)
	}
	if hasMore {
		t.Errorf("LoadGraphResultsPage() hasMore = true, want false (last page)")
	}
	if len(page3) != 1 || page3[0].RecommendationID != "rec-04" {
		t.Errorf("LoadGraphResultsPage() page3 = %+v, want [rec-04]", page3)
	}
}

func TestLoadResourcesPageMissingSessionReturnsEmpty(t *testing.T) {
	s := openTestStore(t)

	page, hasMore, err := s.LoadResourcesPage("does-not-exist", 10, 0)
	if err != nil {
		t.Fatalf("LoadResourcesPage() error = %v, want nil (no rows, not a failure)", err)
	}
	if hasMore {
		t.Errorf("LoadResourcesPage() hasMore = true, want false")
	}
	if len(page) != 0 {
		t.Errorf("LoadResourcesPage() = %+v, want empty", page)
	}
}
