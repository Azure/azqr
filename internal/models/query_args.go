// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

// QueryResultsArgs are the typed arguments for the MCP "azqr-query-results" tool, following the
// ScanArgs/PluginScanArgs convention of a thin JSON-tagged struct bound by
// mcp.NewTypedToolHandler. SessionID is optional: an empty value means "the latest completed
// session". Limit/Offset are validated/clamped against the internal/store page-size constants by
// the handler, not by this struct.
type QueryResultsArgs struct {
	SessionID  string `json:"sessionId,omitempty"`
	ResultType string `json:"resultType,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}
