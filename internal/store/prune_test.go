// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"testing"
	"time"
)

func createAgedSession(t *testing.T, s *Store, status SessionStatus, startedAt time.Time) *Session {
	t.Helper()
	sess, err := s.CreateSession(SessionKindFull, SessionTriggerCLI, NewSessionParams(nil), "test")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if _, err := s.db.Exec(`UPDATE sessions SET started_at = ? WHERE id = ?`, startedAt.Format(time.RFC3339Nano), sess.ID); err != nil {
		t.Fatalf("backdate session %q: %v", sess.ID, err)
	}
	if err := s.FinalizeSession(sess.ID, status, ""); err != nil {
		t.Fatalf("FinalizeSession() error = %v", err)
	}
	sess.StartedAt = startedAt
	return sess
}

func TestPruneSessionsRejectsZeroPolicy(t *testing.T) {
	s := openTestStore(t)
	if _, err := s.PruneSessions(PrunePolicy{}); err == nil {
		t.Error("PruneSessions(zero policy) error = nil, want error (both guards disabled)")
	}
}

func TestPruneSessionsByMaxAge(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	old := createAgedSession(t, s, SessionStatusCompleted, now.Add(-200*24*time.Hour))
	recent := createAgedSession(t, s, SessionStatusCompleted, now.Add(-1*time.Hour))

	result, err := s.PruneSessions(PrunePolicy{MaxAge: 180 * 24 * time.Hour})
	if err != nil {
		t.Fatalf("PruneSessions() error = %v", err)
	}
	if len(result.Removed) != 1 || result.Removed[0].ID != old.ID {
		t.Fatalf("Removed = %+v, want only %s", result.Removed, old.ID)
	}
	if result.DryRun {
		t.Error("result.DryRun = true, want false")
	}

	if _, err := s.GetSession(old.ID); err == nil {
		t.Error("GetSession(old) error = nil, want not-found after prune")
	}
	if _, err := s.GetSession(recent.ID); err != nil {
		t.Errorf("GetSession(recent) error = %v, want nil (recent session must survive)", err)
	}
}

func TestPruneSessionsByMaxCount(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	oldest := createAgedSession(t, s, SessionStatusCompleted, now.Add(-3*time.Hour))
	middle := createAgedSession(t, s, SessionStatusCompleted, now.Add(-2*time.Hour))
	newest := createAgedSession(t, s, SessionStatusCompleted, now.Add(-1*time.Hour))

	result, err := s.PruneSessions(PrunePolicy{MaxCount: 2})
	if err != nil {
		t.Fatalf("PruneSessions() error = %v", err)
	}
	if len(result.Removed) != 1 || result.Removed[0].ID != oldest.ID {
		t.Fatalf("Removed = %+v, want only %s", result.Removed, oldest.ID)
	}

	if _, err := s.GetSession(oldest.ID); err == nil {
		t.Error("GetSession(oldest) error = nil, want not-found after prune")
	}
	if _, err := s.GetSession(middle.ID); err != nil {
		t.Errorf("GetSession(middle) error = %v, want nil", err)
	}
	if _, err := s.GetSession(newest.ID); err != nil {
		t.Errorf("GetSession(newest) error = %v, want nil", err)
	}
}

func TestPruneSessionsDryRunDeletesNothing(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	old := createAgedSession(t, s, SessionStatusCompleted, now.Add(-200*24*time.Hour))

	result, err := s.PruneSessions(PrunePolicy{MaxAge: 180 * 24 * time.Hour, DryRun: true})
	if err != nil {
		t.Fatalf("PruneSessions() error = %v", err)
	}
	if !result.DryRun {
		t.Error("result.DryRun = false, want true")
	}
	if len(result.Removed) != 1 || result.Removed[0].ID != old.ID {
		t.Fatalf("Removed = %+v, want only %s (dry run must still report candidates)", result.Removed, old.ID)
	}

	if _, err := s.GetSession(old.ID); err != nil {
		t.Errorf("GetSession(old) error = %v, want nil (dry run must not delete)", err)
	}
}

func TestPruneSessionsNoMatchesReturnsEmptyResult(t *testing.T) {
	s := openTestStore(t)
	createAgedSession(t, s, SessionStatusCompleted, time.Now().UTC())

	result, err := s.PruneSessions(PrunePolicy{MaxAge: 180 * 24 * time.Hour})
	if err != nil {
		t.Fatalf("PruneSessions() error = %v", err)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %+v, want empty (session is recent)", result.Removed)
	}
}

func TestPruneSessionsCallsBeforeDeleteForEachRemovedSession(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	old1 := createAgedSession(t, s, SessionStatusCompleted, now.Add(-200*24*time.Hour))
	old2 := createAgedSession(t, s, SessionStatusCompleted, now.Add(-190*24*time.Hour))

	var exported []string
	result, err := s.PruneSessions(PrunePolicy{
		MaxAge: 180 * 24 * time.Hour,
		BeforeDelete: func(sess *Session) error {
			exported = append(exported, sess.ID)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("PruneSessions() error = %v", err)
	}
	if len(result.Removed) != 2 {
		t.Fatalf("Removed = %+v, want 2 sessions", result.Removed)
	}
	if len(exported) != 2 || exported[0] != old1.ID || exported[1] != old2.ID {
		t.Errorf("exported = %v, want [%s, %s] in oldest-first order", exported, old1.ID, old2.ID)
	}

	if _, err := s.GetSession(old1.ID); err == nil {
		t.Error("GetSession(old1) error = nil, want not-found after prune")
	}
	if _, err := s.GetSession(old2.ID); err == nil {
		t.Error("GetSession(old2) error = nil, want not-found after prune")
	}
}

func TestPruneSessionsBeforeDeleteErrorAbortsAndPreservesSession(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	old := createAgedSession(t, s, SessionStatusCompleted, now.Add(-200*24*time.Hour))

	_, err := s.PruneSessions(PrunePolicy{
		MaxAge:       180 * 24 * time.Hour,
		BeforeDelete: func(sess *Session) error { return errFake },
	})
	if err == nil {
		t.Fatal("PruneSessions() error = nil, want error from BeforeDelete failure")
	}

	if _, err := s.GetSession(old.ID); err != nil {
		t.Errorf("GetSession(old) error = %v, want nil (a failed export must not delete the session)", err)
	}
}

func TestPruneSessionsDryRunNeverCallsBeforeDelete(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	createAgedSession(t, s, SessionStatusCompleted, now.Add(-200*24*time.Hour))

	called := false
	_, err := s.PruneSessions(PrunePolicy{
		MaxAge:       180 * 24 * time.Hour,
		DryRun:       true,
		BeforeDelete: func(sess *Session) error { called = true; return nil },
	})
	if err != nil {
		t.Fatalf("PruneSessions() error = %v", err)
	}
	if called {
		t.Error("BeforeDelete was called during a dry run, want it skipped")
	}
}
