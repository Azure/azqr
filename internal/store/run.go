// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/rs/zerolog/log"
)

// OpenForScan opens the session store at path (DefaultPath() when empty) for use by a CLI/MCP
// scan call site. On failure it logs a warning and returns nil so callers can fall back to an
// unpersisted scan instead of aborting an otherwise-successful run.
func OpenForScan(path string) *Store {
	s, err := Open(path)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("session store unavailable; continuing scan without persistence")
		return nil
	}
	return s
}

// RunPersistedScan creates a session, registers a Recorder as a pipeline.StageObserver on the
// *pipeline.Scanner the caller constructs inside scanFn, runs scanFn, and finalizes the
// session's terminal status from the resulting error and stage-execution count.
//
// scanFn must call Scan or ScanPlugins on the provided scanner (so the registered Recorder
// observes every stage) and return its result unchanged. This keeps CLI/MCP call sites a thin,
// one-line change: wrap an existing `scanner.Scan(&params)` call in a scanFn closure.
//
// onSessionCreated, when non-nil, is called once with the newly created session immediately
// after CreateSession succeeds and before scanFn runs (it is not called when CreateSession
// fails, since there is no session to report). This lets a long-running caller such as the
// Wails desktop app learn the session ID up front, before the scan itself completes, without
// duplicating any session create/finalize logic.
//
// Trailing extraObservers are registered on the scanner alongside the persistence Recorder (for
// example, a caller that wants per-stage progress notifications without duplicating session
// create/finalize logic). They receive the same StageCompleted callbacks as the Recorder.
func (s *Store) RunPersistedScan(
	kind SessionKind,
	trigger SessionTrigger,
	params *models.ScanParams,
	azqrVersion string,
	scanFn func(scanner *pipeline.Scanner) (*findings.Summary, error),
	onSessionCreated func(*Session),
	extraObservers ...pipeline.StageObserver,
) (*findings.Summary, error) {
	sess, err := s.CreateSession(kind, trigger, NewSessionParams(params), azqrVersion)
	if err != nil {
		log.Warn().Err(err).Msg("store: create session failed, continuing scan without persistence")
		return scanFn(&pipeline.Scanner{Observers: extraObservers})
	}

	if onSessionCreated != nil {
		onSessionCreated(sess)
	}

	recorder := NewRecorder(s, sess.ID, azqrVersion)
	observers := append([]pipeline.StageObserver{recorder}, extraObservers...)
	scanner := &pipeline.Scanner{Observers: observers}

	summary, scanErr := scanFn(scanner)

	status := DeriveStatus(scanErr, scanner.LastMetrics().StagesExecuted)
	errMsg := ""
	if scanErr != nil {
		errMsg = scanErr.Error()
	}
	if finalizeErr := s.FinalizeSession(sess.ID, status, errMsg); finalizeErr != nil {
		log.Warn().Err(finalizeErr).Str("session", sess.ID).Msg("store: finalize session failed")
	}

	if summary != nil {
		if summaryErr := s.SaveSessionSummary(sess.ID, summary); summaryErr != nil {
			log.Warn().Err(summaryErr).Str("session", sess.ID).Msg("store: save session summary failed")
		}
	}

	return summary, scanErr
}
