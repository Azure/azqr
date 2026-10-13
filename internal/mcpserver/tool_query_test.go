// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"path/filepath"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/store"
)

// TestNewPagedResultSetsNextOffsetOnlyWhenHasMore confirms the pagination envelope only
// advertises NextOffset when the page loader reports more rows remain, matching the
// "hasMore"/"nextOffset" contract from the P03-T02 plan task.
func TestNewPagedResultSetsNextOffsetOnlyWhenHasMore(t *testing.T) {
	items := []int{1, 2, 3}

	withMore := newPagedResult("sess-1", queryResultTypeResources, items, 3, 0, true)
	if withMore.NextOffset == nil {
		t.Fatal("NextOffset = nil, want non-nil when hasMore is true")
	}
	if *withMore.NextOffset != 3 {
		t.Errorf("NextOffset = %d, want 3 (offset 0 + limit 3)", *withMore.NextOffset)
	}
	if !withMore.HasMore {
		t.Error("HasMore = false, want true")
	}

	withoutMore := newPagedResult("sess-1", queryResultTypeResources, items, 3, 6, false)
	if withoutMore.NextOffset != nil {
		t.Errorf("NextOffset = %v, want nil when hasMore is false", *withoutMore.NextOffset)
	}
	if withoutMore.SessionID != "sess-1" || withoutMore.ResultType != queryResultTypeResources {
		t.Errorf("unexpected envelope identity: %+v", withoutMore)
	}
	if withoutMore.Limit != 3 || withoutMore.Offset != 6 {
		t.Errorf("Limit/Offset = %d/%d, want 3/6", withoutMore.Limit, withoutMore.Offset)
	}
}

// openTestStore opens a fresh SQLite-backed *store.Store at a temp-dir path, isolated from the
// real ~/.azqr store that store.OpenForScan("") would otherwise touch.
func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "azqr.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestResolveQuerySessionIDDefaultsToLatestCompleted confirms an empty SessionID selects the most
// recently completed session, never an in-progress or failed one, per P03-T02's requirement.
func TestResolveQuerySessionIDDefaultsToLatestCompleted(t *testing.T) {
	s := openTestStore(t)

	failed, err := s.CreateSession(store.SessionKindFull, store.SessionTriggerMCP, store.SessionParams{}, "test")
	if err != nil {
		t.Fatalf("CreateSession(failed) error = %v", err)
	}
	if err := s.FinalizeSession(failed.ID, store.SessionStatusFailed, "boom"); err != nil {
		t.Fatalf("FinalizeSession(failed) error = %v", err)
	}

	completed, err := s.CreateSession(store.SessionKindFull, store.SessionTriggerMCP, store.SessionParams{}, "test")
	if err != nil {
		t.Fatalf("CreateSession(completed) error = %v", err)
	}
	if err := s.FinalizeSession(completed.ID, store.SessionStatusCompleted, ""); err != nil {
		t.Fatalf("FinalizeSession(completed) error = %v", err)
	}

	got, err := resolveQuerySessionID(s, models.QueryResultsArgs{})
	if err != nil {
		t.Fatalf("resolveQuerySessionID() error = %v", err)
	}
	if got != completed.ID {
		t.Errorf("resolveQuerySessionID() = %q, want latest completed session %q", got, completed.ID)
	}
}

// TestResolveQuerySessionIDUsesExplicitID confirms an explicit SessionID is honored (and
// validated) instead of being overridden by the latest-completed default.
func TestResolveQuerySessionIDUsesExplicitID(t *testing.T) {
	s := openTestStore(t)

	sess, err := s.CreateSession(store.SessionKindFull, store.SessionTriggerMCP, store.SessionParams{}, "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	got, err := resolveQuerySessionID(s, models.QueryResultsArgs{SessionID: sess.ID})
	if err != nil {
		t.Fatalf("resolveQuerySessionID() error = %v", err)
	}
	if got != sess.ID {
		t.Errorf("resolveQuerySessionID() = %q, want explicit %q", got, sess.ID)
	}

	if _, err := resolveQuerySessionID(s, models.QueryResultsArgs{SessionID: "does-not-exist"}); err == nil {
		t.Error("resolveQuerySessionID() with an unknown session ID: error = nil, want non-nil")
	}
}

// TestResolveQuerySessionIDNoCompletedSessionsReturnsError confirms a store with no completed
// sessions (even if non-completed ones exist) returns a clear error rather than an empty ID.
func TestResolveQuerySessionIDNoCompletedSessionsReturnsError(t *testing.T) {
	s := openTestStore(t)

	if _, err := resolveQuerySessionID(s, models.QueryResultsArgs{}); err == nil {
		t.Error("resolveQuerySessionID() on an empty store: error = nil, want non-nil")
	}
}
