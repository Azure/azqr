// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/oklog/ulid/v2"
)

// SessionKind distinguishes the scope of a scan session.
type SessionKind string

const (
	// SessionKindFull is a session created by the default pipeline (pipeline.BuildDefault).
	SessionKindFull SessionKind = "full"
	// SessionKindPluginOnly is a session created by the plugin-only pipeline (pipeline.BuildPluginOnly).
	SessionKindPluginOnly SessionKind = "plugin_only"
)

// SessionTrigger identifies which caller started a session.
type SessionTrigger string

const (
	SessionTriggerCLI     SessionTrigger = "cli"
	SessionTriggerMCP     SessionTrigger = "mcp"
	SessionTriggerDesktop SessionTrigger = "desktop"
)

// SessionStatus is the lifecycle state of a session.
type SessionStatus string

const (
	// SessionStatusRunning is set immediately when a session is created, before the pipeline executes.
	SessionStatusRunning SessionStatus = "running"
	// SessionStatusCompleted means every stage executed (or was intentionally skipped) without error.
	SessionStatusCompleted SessionStatus = "completed"
	// SessionStatusPartial means the pipeline stopped on a stage error after at least one earlier
	// stage had already completed and persisted data, so the session has usable partial results.
	SessionStatusPartial SessionStatus = "partial"
	// SessionStatusFailed means the pipeline stopped on a stage error before any stage completed.
	SessionStatusFailed SessionStatus = "failed"
	// SessionStatusCancelled means the caller cancelled the session before it finalized naturally.
	SessionStatusCancelled SessionStatus = "cancelled"
)

// SessionParams captures the subset of models.ScanParams needed to display and
// reproduce a session's scope. It intentionally does not embed models.ScanParams
// directly: ScanParams carries unexported fields and interface-valued members
// (models.StageConfigs, models.Filters) that cannot round-trip through JSON.
type SessionParams struct {
	ManagementGroups []string `json:"managementGroups,omitempty"`
	Subscriptions    []string `json:"subscriptions,omitempty"`
	ResourceGroups   []string `json:"resourceGroups,omitempty"`
	ScannerKeys      []string `json:"scannerKeys,omitempty"`
	EnabledStages    []string `json:"enabledStages,omitempty"`
	OutputName       string   `json:"outputName,omitempty"`
	Mask             bool     `json:"mask"`
	Xlsx             bool     `json:"xlsx,omitempty"`
	Csv              bool     `json:"csv,omitempty"`
	Json             bool     `json:"json,omitempty"`
	Sarif            bool     `json:"sarif,omitempty"`
}

// NewSessionParams builds a SessionParams snapshot from live models.ScanParams.
// A nil params is treated as an empty scope rather than an error.
func NewSessionParams(params *models.ScanParams) SessionParams {
	if params == nil {
		return SessionParams{}
	}

	sp := SessionParams{
		ManagementGroups: params.ManagementGroups,
		Subscriptions:    params.Subscriptions,
		ResourceGroups:   params.ResourceGroups,
		ScannerKeys:      params.ScannerKeys,
		OutputName:       params.OutputName,
		Mask:             params.Mask,
		Xlsx:             params.Xlsx,
		Csv:              params.Csv,
		Json:             params.Json,
		Sarif:            params.Sarif,
	}
	if params.Stages != nil {
		sp.EnabledStages = params.Stages.GetEnabledStages()
	}
	return sp
}

// Session is one recorded scan, whether still running, completed, partial, failed, or cancelled.
type Session struct {
	ID           string
	Kind         SessionKind
	Trigger      SessionTrigger
	Status       SessionStatus
	AzqrVersion  string
	Params       SessionParams
	StartedAt    time.Time
	FinishedAt   *time.Time
	DurationMs   *int64
	ErrorMessage string
	// ScopeID is a deterministic hash of the session's declared scan scope (see ScopeID),
	// computed and persisted at creation time. Empty for sessions created before this field
	// existed.
	ScopeID string
	// TenantID is the Azure AD tenant of the credential used for this scan, captured from the
	// first acquired token's `tid` claim once the Initialization stage runs. Empty until
	// captured, or if capture failed.
	TenantID string
}

// ScopeID returns a deterministic, 64-character hex-encoded SHA-256 hash of the session's
// declared scan scope (management groups, subscriptions, resource groups, scanner keys, and
// enabled stages), independent of input slice ordering. It intentionally reads only
// SessionParams' already-exported, already-deterministic fields rather than the raw
// models.Filters object, which carries interface-valued and unexported members that do not
// round-trip through JSON and therefore cannot hash deterministically.
func ScopeID(params SessionParams) string {
	managementGroups := sortedLower(params.ManagementGroups)
	subscriptions := sortedLower(params.Subscriptions)
	resourceGroups := sortedLower(params.ResourceGroups)
	scannerKeys := sortedLower(params.ScannerKeys)
	enabledStages := sortedLower(params.EnabledStages)

	descriptor := strings.Join([]string{
		strings.Join(managementGroups, ","),
		strings.Join(subscriptions, ","),
		strings.Join(resourceGroups, ","),
		strings.Join(scannerKeys, ","),
		strings.Join(enabledStages, ","),
	}, "|")

	sum := sha256.Sum256([]byte(descriptor))
	return hex.EncodeToString(sum[:])
}

// sortedLower returns a sorted, lower-cased copy of values, leaving the input slice untouched.
func sortedLower(values []string) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = strings.ToLower(v)
	}
	sort.Strings(out)
	return out
}

// newSessionID returns a new ULID string: lexicographically sortable by creation time.
func newSessionID() (string, error) {
	id, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return id.String(), nil
}

// CreateSession inserts a new sessions row with status SessionStatusRunning and returns it.
// Call this immediately before pipeline.Pipeline.Execute runs.
func (s *Store) CreateSession(kind SessionKind, trigger SessionTrigger, params SessionParams, azqrVersion string) (*Session, error) {
	id, err := newSessionID()
	if err != nil {
		return nil, err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal session params: %w", err)
	}

	sess := &Session{
		ID:          id,
		Kind:        kind,
		Trigger:     trigger,
		Status:      SessionStatusRunning,
		AzqrVersion: azqrVersion,
		Params:      params,
		StartedAt:   time.Now().UTC(),
		ScopeID:     ScopeID(params),
	}

	_, err = s.db.Exec(
		s.rebind(`INSERT INTO sessions (id, kind, trigger, status, azqr_version, scan_params_json, started_at, scope_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`),
		sess.ID, string(sess.Kind), string(sess.Trigger), string(sess.Status), sess.AzqrVersion,
		string(paramsJSON), sess.StartedAt.Format(time.RFC3339Nano), sess.ScopeID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert session %q: %w", id, err)
	}

	return sess, nil
}

// FinalizeSession sets a session's terminal status, finish time, duration, and optional error
// message. Call this once, after pipeline.Pipeline.Execute returns.
func (s *Store) FinalizeSession(id string, status SessionStatus, errMessage string) error {
	var startedAtRaw string
	err := s.db.QueryRow(s.rebind(`SELECT started_at FROM sessions WHERE id = ?`), id).Scan(&startedAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("finalize session %q: not found", id)
		}
		return fmt.Errorf("finalize session %q: %w", id, err)
	}
	startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
	if err != nil {
		return fmt.Errorf("finalize session %q: parse started_at %q: %w", id, startedAtRaw, err)
	}

	finishedAt := time.Now().UTC()
	durationMs := finishedAt.Sub(startedAt).Milliseconds()

	var errMsg sql.NullString
	if errMessage != "" {
		errMsg = sql.NullString{String: errMessage, Valid: true}
	}

	_, err = s.db.Exec(
		s.rebind(`UPDATE sessions SET status = ?, finished_at = ?, duration_ms = ?, error_message = ? WHERE id = ?`),
		string(status), finishedAt.Format(time.RFC3339Nano), durationMs, errMsg, id,
	)
	if err != nil {
		return fmt.Errorf("finalize session %q: %w", id, err)
	}
	return nil
}

// SaveSessionSummary persists a completed scan's already-computed *findings.Summary into the
// session's summary_json column, so a later LoadSessionSummary call can retrieve it in one
// bounded read instead of reconstructing it from the session's full resource/graph rows. Call
// this once, alongside FinalizeSession, when the finished *findings.Summary is available.
func (s *Store) SaveSessionSummary(id string, summary *findings.Summary) error {
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("marshal summary for session %q: %w", id, err)
	}

	_, err = s.db.Exec(
		s.rebind(`UPDATE sessions SET summary_json = ? WHERE id = ?`),
		string(summaryJSON), id,
	)
	if err != nil {
		return fmt.Errorf("save summary for session %q: %w", id, err)
	}
	return nil
}

// LoadSessionSummary reads a session's persisted findings.Summary directly from summary_json,
// without touching resources, recommendation_results, or any other per-session result table.
// It returns (nil, nil) — not an error — when the session exists but has no persisted summary
// (recorded before this column existed, or the save step failed), so callers can distinguish
// "summary not available" from a query failure. A missing session is still reported as an error.
func (s *Store) LoadSessionSummary(id string) (*findings.Summary, error) {
	var summaryJSON sql.NullString
	err := s.db.QueryRow(s.rebind(`SELECT summary_json FROM sessions WHERE id = ?`), id).Scan(&summaryJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session %q: not found", id)
		}
		return nil, fmt.Errorf("load summary for session %q: %w", id, err)
	}
	if !summaryJSON.Valid || summaryJSON.String == "" {
		return nil, nil
	}

	var summary findings.Summary
	if err := json.Unmarshal([]byte(summaryJSON.String), &summary); err != nil {
		return nil, fmt.Errorf("session %q: unmarshal summary_json: %w", id, err)
	}
	return &summary, nil
}

// DeriveStatus maps a pipeline outcome to a SessionStatus:
//   - pipelineErr == nil                           -> SessionStatusCompleted
//   - pipelineErr != nil && stagesExecuted == 0     -> SessionStatusFailed (nothing persisted)
//   - pipelineErr != nil && stagesExecuted > 0       -> SessionStatusPartial (some stages persisted)
//
// pipeline.Pipeline.Execute returns on the first stage error, so at most one stage ever fails
// per run; stagesExecuted is the number of stages that completed (including the failing one, if
// any) before Execute returned, taken from pipeline.PipelineMetrics.StagesExecuted.
func DeriveStatus(pipelineErr error, stagesExecuted int) SessionStatus {
	if pipelineErr == nil {
		return SessionStatusCompleted
	}
	if stagesExecuted <= 1 {
		// Either the failing stage was the first one executed, or nothing executed at all:
		// no earlier stage had a chance to persist usable data.
		return SessionStatusFailed
	}
	return SessionStatusPartial
}
