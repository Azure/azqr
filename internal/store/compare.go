// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"fmt"
	"strings"
)

// ResourceChange describes a resource present in one session but not the other, identified by
// its natural `resource_id`.
type ResourceChange struct {
	ResourceID    string
	Type          string
	Name          string
	ResourceGroup string
	ChangeType    string // "added" (present in B, not A) or "removed" (present in A, not B)
}

// RecommendationChange describes a recommendation whose impacted-resource count changed, or
// that was newly hit or no longer hit, between two sessions.
type RecommendationChange struct {
	RecommendationID      string
	Recommendation        string
	Category              string
	Impact                string
	ResourceType          string
	OldImpactedResources  int
	NewImpactedResources  int
	ImpactedResourcesDiff int
	ChangeType            string // "added", "removed", or "changed"
}

// SessionComparisonResult is the result of comparing two sessions by ID, built entirely from
// natural-key set-difference queries against the store rather than parsing exported files.
type SessionComparisonResult struct {
	SourceSessionID       string
	TargetSessionID       string
	ResourceChanges       []ResourceChange
	RecommendationChanges []RecommendationChange
}

type resourceRow struct {
	ResourceID    string
	Type          string
	Name          string
	ResourceGroup string
}

// loadResourceRows returns a map of resource_id -> resourceRow for a session, keyed by the
// lower-cased resource ID. Azure Resource Manager IDs are case-insensitive, and segment casing
// (subscription, resource group) is not guaranteed stable across separate scans of the same
// environment, so comparisons must not treat a casing-only difference as added/removed. The
// original (as-scanned) casing is preserved in resourceRow.ResourceID for display.
func (s *Store) loadResourceRows(sessionID string) (map[string]resourceRow, error) {
	rows, err := s.db.Query(
		s.rebind(`SELECT resource_id, type, name, resource_group FROM resources WHERE session_id = ?`),
		sessionID)
	if err != nil {
		return nil, fmt.Errorf("query resources for session %s: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]resourceRow)
	for rows.Next() {
		var id string
		var r resourceRow
		if err := rows.Scan(&id, &r.Type, &r.Name, &r.ResourceGroup); err != nil {
			return nil, fmt.Errorf("scan resource row: %w", err)
		}
		r.ResourceID = id
		result[strings.ToLower(id)] = r
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resource rows: %w", err)
	}
	return result, nil
}

type recommendationAgg struct {
	Recommendation string
	Category       string
	Impact         string
	ResourceType   string
	ImpactedCount  int
}

// loadRecommendationAggregates returns, per recommendation_id, the catalog metadata and the
// count of distinct resources hit by that recommendation in the given session.
func (s *Store) loadRecommendationAggregates(sessionID string) (map[string]recommendationAgg, error) {
	rows, err := s.db.Query(s.rebind(`
		SELECT rr.recommendation_id, c.recommendation, c.category, c.impact, c.resource_type,
			COUNT(DISTINCT rr.resource_id) AS impacted
		FROM recommendation_results rr
		JOIN recommendation_catalog c ON c.recommendation_id = rr.recommendation_id
		WHERE rr.session_id = ?
		GROUP BY rr.recommendation_id, c.recommendation, c.category, c.impact, c.resource_type`),
		sessionID)
	if err != nil {
		return nil, fmt.Errorf("query recommendation results for session %s: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]recommendationAgg)
	for rows.Next() {
		var id string
		var agg recommendationAgg
		if err := rows.Scan(&id, &agg.Recommendation, &agg.Category, &agg.Impact, &agg.ResourceType, &agg.ImpactedCount); err != nil {
			return nil, fmt.Errorf("scan recommendation aggregate row: %w", err)
		}
		result[id] = agg
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recommendation aggregate rows: %w", err)
	}
	return result, nil
}

// CompareSessions compares two persisted sessions by ID and returns resource and recommendation
// differences, computed as plain set-difference queries against `resource_id`/`recommendation_id`
// natural keys rather than requiring either session to be exported to a file first.
func (s *Store) CompareSessions(sourceSessionID, targetSessionID string) (*SessionComparisonResult, error) {
	sourceSession, err := s.GetSession(sourceSessionID)
	if err != nil {
		return nil, fmt.Errorf("source session %s: %w", sourceSessionID, err)
	}
	targetSession, err := s.GetSession(targetSessionID)
	if err != nil {
		return nil, fmt.Errorf("target session %s: %w", targetSessionID, err)
	}

	if sourceSession.ScopeID != "" && targetSession.ScopeID != "" && sourceSession.ScopeID != targetSession.ScopeID {
		return nil, fmt.Errorf(
			"cannot compare sessions with different scan scopes: source session %s (scope %s) vs target session %s (scope %s)",
			sourceSessionID, sourceSession.ScopeID, targetSessionID, targetSession.ScopeID,
		)
	}
	if sourceSession.TenantID != "" && targetSession.TenantID != "" && sourceSession.TenantID != targetSession.TenantID {
		return nil, fmt.Errorf(
			"cannot compare sessions with different tenants: source session %s (tenant %s) vs target session %s (tenant %s)",
			sourceSessionID, sourceSession.TenantID, targetSessionID, targetSession.TenantID,
		)
	}

	resourcesA, err := s.loadResourceRows(sourceSessionID)
	if err != nil {
		return nil, err
	}
	resourcesB, err := s.loadResourceRows(targetSessionID)
	if err != nil {
		return nil, err
	}

	result := &SessionComparisonResult{SourceSessionID: sourceSessionID, TargetSessionID: targetSessionID}

	for id, r := range resourcesB {
		if _, ok := resourcesA[id]; !ok {
			result.ResourceChanges = append(result.ResourceChanges, ResourceChange{
				ResourceID: r.ResourceID, Type: r.Type, Name: r.Name, ResourceGroup: r.ResourceGroup, ChangeType: "added",
			})
		}
	}
	for id, r := range resourcesA {
		if _, ok := resourcesB[id]; !ok {
			result.ResourceChanges = append(result.ResourceChanges, ResourceChange{
				ResourceID: r.ResourceID, Type: r.Type, Name: r.Name, ResourceGroup: r.ResourceGroup, ChangeType: "removed",
			})
		}
	}

	recsA, err := s.loadRecommendationAggregates(sourceSessionID)
	if err != nil {
		return nil, err
	}
	recsB, err := s.loadRecommendationAggregates(targetSessionID)
	if err != nil {
		return nil, err
	}

	processed := make(map[string]bool)
	for id, b := range recsB {
		processed[id] = true
		if a, ok := recsA[id]; ok {
			if a.ImpactedCount != b.ImpactedCount {
				result.RecommendationChanges = append(result.RecommendationChanges, RecommendationChange{
					RecommendationID: id, Recommendation: b.Recommendation, Category: b.Category,
					Impact: b.Impact, ResourceType: b.ResourceType,
					OldImpactedResources: a.ImpactedCount, NewImpactedResources: b.ImpactedCount,
					ImpactedResourcesDiff: b.ImpactedCount - a.ImpactedCount, ChangeType: "changed",
				})
			}
		} else {
			result.RecommendationChanges = append(result.RecommendationChanges, RecommendationChange{
				RecommendationID: id, Recommendation: b.Recommendation, Category: b.Category,
				Impact: b.Impact, ResourceType: b.ResourceType,
				OldImpactedResources: 0, NewImpactedResources: b.ImpactedCount,
				ImpactedResourcesDiff: b.ImpactedCount, ChangeType: "added",
			})
		}
	}
	for id, a := range recsA {
		if processed[id] {
			continue
		}
		result.RecommendationChanges = append(result.RecommendationChanges, RecommendationChange{
			RecommendationID: id, Recommendation: a.Recommendation, Category: a.Category,
			Impact: a.Impact, ResourceType: a.ResourceType,
			OldImpactedResources: a.ImpactedCount, NewImpactedResources: 0,
			ImpactedResourcesDiff: -a.ImpactedCount, ChangeType: "removed",
		})
	}

	return result, nil
}
