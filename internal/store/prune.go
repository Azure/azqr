// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"fmt"
	"time"
)

// Default retention policy applied by `azqr session prune` when the user does not override it.
// DefaultPruneMaxAge keeps roughly 6 months of history; DefaultPruneMaxCount is a secondary
// guard against unbounded growth even when sessions are created in rapid succession.
const (
	DefaultPruneMaxAge   = 180 * 24 * time.Hour
	DefaultPruneMaxCount = 500
)

// PrunePolicy configures Store.PruneSessions. A zero value for MaxAge or MaxCount disables that
// guard; at least one of them must be positive, otherwise PruneSessions returns an error rather
// than silently pruning nothing (or everything).
type PrunePolicy struct {
	// MaxAge removes sessions whose started_at is older than now-MaxAge. Zero disables this guard.
	MaxAge time.Duration
	// MaxCount keeps at most this many of the newest sessions, removing older ones beyond it.
	// Zero disables this guard.
	MaxCount int
	// DryRun, when true, computes and returns the sessions that would be removed without
	// deleting anything.
	DryRun bool
	// BeforeDelete, when non-nil, is called for each session immediately before it is deleted
	// (never called in a dry run). It is the export-before-delete safety hook: a caller can
	// export the session (for example via Store.LoadReportData + an exporter) and return an
	// error to abort that session's deletion, so a failed export never silently loses data. A
	// BeforeDelete error stops the whole prune and is returned to the caller; sessions already
	// deleted earlier in the run remain deleted.
	BeforeDelete func(sess *Session) error
}

// PruneResult reports which sessions were (or, in a dry run, would be) removed by PruneSessions.
type PruneResult struct {
	// Removed lists the sessions that were deleted (or would be, for a dry run), oldest first.
	Removed []*Session
	// DryRun mirrors the policy's DryRun flag, so callers can tell apart "nothing to prune" from
	// "a dry run found nothing to delete".
	DryRun bool
}

// PruneSessions removes sessions matching policy's age and/or count thresholds, cascading
// through the same foreign-key relationships as DeleteSession. Sessions are evaluated oldest
// first: a session is eligible once started_at is older than now-MaxAge, or once it falls
// outside the MaxCount newest sessions (whichever guards are non-zero). In a dry run, no
// sessions are deleted; Removed reports what would be.
func (s *Store) PruneSessions(policy PrunePolicy) (*PruneResult, error) {
	if policy.MaxAge <= 0 && policy.MaxCount <= 0 {
		return nil, fmt.Errorf("prune policy must set a positive MaxAge and/or MaxCount")
	}

	all, err := s.ListSessions(SessionFilter{}) // newest first
	if err != nil {
		return nil, fmt.Errorf("list sessions for prune: %w", err)
	}

	eligible := make(map[string]*Session)

	if policy.MaxAge > 0 {
		cutoff := time.Now().UTC().Add(-policy.MaxAge)
		for _, sess := range all {
			if sess.StartedAt.Before(cutoff) {
				eligible[sess.ID] = sess
			}
		}
	}

	if policy.MaxCount > 0 && len(all) > policy.MaxCount {
		// all is newest-first, so the overflow beyond MaxCount is the oldest tail.
		for _, sess := range all[policy.MaxCount:] {
			eligible[sess.ID] = sess
		}
	}

	result := &PruneResult{DryRun: policy.DryRun}
	for _, sess := range all {
		if _, ok := eligible[sess.ID]; ok {
			result.Removed = append(result.Removed, sess)
		}
	}
	// Report oldest first, matching the "pruned oldest sessions" framing.
	for i, j := 0, len(result.Removed)-1; i < j; i, j = i+1, j-1 {
		result.Removed[i], result.Removed[j] = result.Removed[j], result.Removed[i]
	}

	if policy.DryRun {
		return result, nil
	}

	for _, sess := range result.Removed {
		if policy.BeforeDelete != nil {
			if err := policy.BeforeDelete(sess); err != nil {
				return nil, fmt.Errorf("export session %q before delete: %w", sess.ID, err)
			}
		}
		if err := s.DeleteSession(sess.ID); err != nil {
			return nil, fmt.Errorf("delete session %q during prune: %w", sess.ID, err)
		}
	}

	return result, nil
}
