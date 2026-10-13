// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"testing"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TestScanSummaryEnvelopeTrimsToImpactedRecommendations confirms scanSummaryEnvelope builds the
// exact {sessionId, scopeId, summary} shape committed to by the P02 plan tasks, and that the
// summary's Recommendations are filtered to impacted-only entries (PC-001).
func TestScanSummaryEnvelopeTrimsToImpactedRecommendations(t *testing.T) {
	sess := &store.Session{ID: "sess-123", ScopeID: "scope-abc"}
	summary := &findings.Summary{
		Recommendations: []findings.Recommendation{
			{ID: "rec-impacted", ImpactedResources: 2},
			{ID: "rec-not-impacted", ImpactedResources: 0},
		},
		Resources: 10,
	}

	env := scanSummaryEnvelope(sess, summary)

	if env.SessionID != "sess-123" || env.ScopeID != "scope-abc" {
		t.Errorf("envelope identity = {%q, %q}, want {sess-123, scope-abc}", env.SessionID, env.ScopeID)
	}
	if env.Summary == nil {
		t.Fatal("envelope.Summary = nil, want non-nil")
	}
	if len(env.Summary.Recommendations) != 1 || env.Summary.Recommendations[0].ID != "rec-impacted" {
		t.Errorf("envelope.Summary.Recommendations = %+v, want only the impacted recommendation", env.Summary.Recommendations)
	}
	if env.Summary.Resources != 10 {
		t.Errorf("envelope.Summary.Resources = %d, want 10 (unrelated fields copied through)", env.Summary.Resources)
	}
}

// TestScanSummaryEnvelopeHandlesNilSessionAndReport confirms scanSummaryEnvelope never panics and
// leaves SessionID/ScopeID/Summary empty when the session store was unavailable or the scan
// produced no report, matching its documented nil-safety contract.
func TestScanSummaryEnvelopeHandlesNilSessionAndReport(t *testing.T) {
	env := scanSummaryEnvelope(nil, nil)

	if env.SessionID != "" || env.ScopeID != "" {
		t.Errorf("envelope identity = {%q, %q}, want empty when session is nil", env.SessionID, env.ScopeID)
	}
	if env.Summary != nil {
		t.Errorf("envelope.Summary = %+v, want nil when report is nil", env.Summary)
	}
}

// TestScanResultTextReflectsRequestedFiles confirms scanResultText (FR-004/FR-005) reports only
// the files actually requested via xlsx/json, preserves the exact "both requested" wording the
// plan commits to for backward compatibility, and points callers at azqr-query-results when
// neither file was generated.
func TestScanResultTextReflectsRequestedFiles(t *testing.T) {
	sess := &store.Session{ID: "sess-123"}

	tests := []struct {
		name string
		xlsx bool
		json bool
		want string
	}{
		{
			name: "both requested preserves exact prior wording",
			xlsx: true,
			json: true,
			want: "Scan results saved to:\n- Excel: file:///out.xlsx\n- JSON: file:///out.json",
		},
		{
			name: "xlsx only",
			xlsx: true,
			json: false,
			want: "Scan results saved to:\n- Excel: file:///out.xlsx",
		},
		{
			name: "json only",
			xlsx: false,
			json: true,
			want: "Scan results saved to:\n- JSON: file:///out.json",
		},
		{
			name: "neither requested points to azqr-query-results",
			xlsx: false,
			json: false,
			want: `No files were generated (xlsx/json export is opt-in and was not requested). Use the azqr-query-results tool with session "sess-123" to retrieve the scan's data.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanResultText(sess, "/out", tt.xlsx, tt.json)
			if got != tt.want {
				t.Errorf("scanResultText() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRegisterToolsRegistersNoResources confirms RegisterTools (which wires up scanHandler and
// scanPluginHandler) never registers an mcp.Resource, so an MCP client can never fetch scan
// output via a "resources/list" + "resources/read" round trip. This is a standing invariant, not
// tied to any one scan invocation: scanHandler/scanPluginHandler return their result only via
// mcp.NewToolResultStructured's StructuredContent field (see tool_scan.go, tool_plugin.go), with
// no call to mcp.MCPServer.AddResource anywhere in this package.
func TestRegisterToolsRegistersNoResources(t *testing.T) {
	testServer := server.NewMCPServer(
		"Test Server",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	RegisterTools(testServer)

	response := testServer.HandleMessage(context.Background(), []byte(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "resources/list"
	}`))

	resp, ok := response.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("HandleMessage(resources/list) = %T, want mcp.JSONRPCResponse", response)
	}
	result, ok := resp.Result.(mcp.ListResourcesResult)
	if !ok {
		t.Fatalf("resp.Result = %T, want mcp.ListResourcesResult", resp.Result)
	}
	if len(result.Resources) != 0 {
		t.Errorf("resources/list after RegisterTools() = %d resource(s), want 0", len(result.Resources))
	}
}
