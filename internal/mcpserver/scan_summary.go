// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/store"
)

// ScanSummaryEnvelope is the structured content returned by the MCP scan tools: a small,
// session-identified summary rather than the full in-memory *renderers.ReportData. SessionID
// and ScopeID are left empty when the session store was unavailable, since no session was
// persisted to identify.
type ScanSummaryEnvelope struct {
	SessionID string            `json:"sessionId,omitempty"`
	ScopeID   string            `json:"scopeId,omitempty"`
	Summary   *findings.Summary `json:"summary,omitempty"`
}

// scanSummaryEnvelope builds the structured-content envelope for a completed MCP scan: the
// session identity captured via onSessionCreated (sess is nil when the store was unavailable)
// plus the already-computed findings.Summary, trimmed to impacted recommendations only. It
// never panics on a nil session or nil summary; the envelope simply omits what is missing.
func scanSummaryEnvelope(sess *store.Session, summary *findings.Summary) *ScanSummaryEnvelope {
	env := &ScanSummaryEnvelope{}
	if sess != nil {
		env.SessionID = sess.ID
		env.ScopeID = sess.ScopeID
	}
	if summary != nil {
		env.Summary = trimSummaryToImpacted(summary)
	}
	return env
}

// trimSummaryToImpacted returns a shallow copy of summary with Recommendations filtered down to
// entries that impacted at least one resource, so the envelope stays small regardless of the
// size of the full recommendation catalog.
func trimSummaryToImpacted(summary *findings.Summary) *findings.Summary {
	filtered := make([]findings.Recommendation, 0, len(summary.Recommendations))
	for _, r := range summary.Recommendations {
		if r.ImpactedResources > 0 {
			filtered = append(filtered, r)
		}
	}
	trimmed := *summary
	trimmed.Recommendations = filtered
	return &trimmed
}

// scanResultText builds the MCP tool result text for a completed scan, reporting only the files
// actually written (xlsx/json file export is opt-in and off by default; see ScanArgs/PluginScanArgs).
// When neither file was generated, it points the caller at the azqr-query-results tool to retrieve
// the session's data instead.
func scanResultText(sess *store.Session, outputName string, xlsx, json bool) string {
	var lines []string
	if xlsx {
		lines = append(lines, fmt.Sprintf("- Excel: file://%s.xlsx", outputName))
	}
	if json {
		lines = append(lines, fmt.Sprintf("- JSON: file://%s.json", outputName))
	}
	if len(lines) == 0 {
		sessionID := ""
		if sess != nil {
			sessionID = sess.ID
		}
		return fmt.Sprintf("No files were generated (xlsx/json export is opt-in and was not requested). Use the azqr-query-results tool with session %q to retrieve the scan's data.", sessionID)
	}
	return "Scan results saved to:\n" + strings.Join(lines, "\n")
}
