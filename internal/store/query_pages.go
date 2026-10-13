// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"fmt"

	"github.com/Azure/azqr/internal/models"
)

// Page size bounds for the paginated session-result loaders. Callers cannot request an
// unbounded page: DefaultResultPageSize applies when a caller passes a non-positive limit, and
// MaxResultPageSize caps any larger request.
const (
	DefaultResultPageSize = 100
	MaxResultPageSize     = 1000
)

// clampPageSize normalizes a caller-supplied page size to (0, MaxResultPageSize], defaulting
// non-positive values to DefaultResultPageSize.
func clampPageSize(limit int) int {
	if limit <= 0 {
		return DefaultResultPageSize
	}
	if limit > MaxResultPageSize {
		return MaxResultPageSize
	}
	return limit
}

// ClampResultPageSize exposes clampPageSize's normalization to callers outside this package (for
// example, the MCP query tool) so reported pagination metadata matches the limit actually applied
// by LoadResourcesPage/LoadGraphResultsPage without duplicating the bounds logic.
func ClampResultPageSize(limit int) int {
	return clampPageSize(limit)
}

// clampOffset normalizes a caller-supplied page offset to a non-negative value. A negative offset
// is not a meaningful request for "rows before the start"; some backends (SQLite) silently treat
// it as 0, while others (Postgres) reject it outright, so this clamp keeps behavior identical and
// well-defined across backends rather than depending on driver-specific tolerance.
func clampOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

// ClampResultPageOffset exposes clampOffset's normalization to callers outside this package (for
// example, the MCP query tool) so reported pagination metadata matches the offset actually
// applied by LoadResourcesPage/LoadGraphResultsPage without duplicating the bounds logic.
func ClampResultPageOffset(offset int) int {
	return clampOffset(offset)
}

// LoadResourcesPage returns one page of a session's non-excluded resources, ordered
// deterministically by resource_id, plus whether more rows remain beyond this page. limit is
// clamped via clampPageSize; offset is clamped via clampOffset to a non-negative value.
func (s *Store) LoadResourcesPage(sessionID string, limit, offset int) ([]*models.Resource, bool, error) {
	limit = clampPageSize(limit)
	offset = clampOffset(offset)

	rows, err := s.db.Query(
		s.rebind(resourceSelectColumns+` ORDER BY resource_id LIMIT ? OFFSET ?`),
		sessionID, limit+1, offset,
	)
	if err != nil {
		return nil, false, fmt.Errorf("load resources page for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	items := []*models.Resource{}
	for rows.Next() {
		res, err := scanResourceRow(rows)
		if err != nil {
			return nil, false, err
		}
		items = append(items, res)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("load resources page for session %q: %w", sessionID, err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

// LoadGraphResultsPage returns one page of a session's recommendation/graph results, ordered
// deterministically by recommendation_id then resource_id, plus whether more rows remain beyond
// this page. limit is clamped via clampPageSize; offset is clamped via clampOffset to a
// non-negative value.
func (s *Store) LoadGraphResultsPage(sessionID string, limit, offset int) ([]*models.GraphResult, bool, error) {
	limit = clampPageSize(limit)
	offset = clampOffset(offset)

	rows, err := s.db.Query(
		s.rebind(graphResultSelectColumns+` ORDER BY rr.recommendation_id, rr.resource_id LIMIT ? OFFSET ?`),
		sessionID, limit+1, offset,
	)
	if err != nil {
		return nil, false, fmt.Errorf("load recommendation results page for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	items := []*models.GraphResult{}
	for rows.Next() {
		g, err := scanGraphResultRow(rows)
		if err != nil {
			return nil, false, err
		}
		items = append(items, g)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("load recommendation results page for session %q: %w", sessionID, err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}
