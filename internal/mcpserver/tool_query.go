// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
)

// queryResultsPage is the JSON shape returned by the "azqr-query-results" tool. It is always a
// bounded page (or, for resultType "summary", a single already-bounded summary) — never a full
// *renderers.ReportData and never a file blob, per P03-T02's contract.
type queryResultsPage struct {
	SessionID  string            `json:"sessionId"`
	ResultType string            `json:"resultType"`
	Items      any               `json:"items,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Offset     int               `json:"offset,omitempty"`
	HasMore    bool              `json:"hasMore"`
	NextOffset *int              `json:"nextOffset,omitempty"`
	Summary    *findings.Summary `json:"summary,omitempty"`
}

const (
	queryResultTypeResources       = "resources"
	queryResultTypeRecommendations = "recommendations"
	queryResultTypeSummary         = "summary"
)

// resolveQuerySessionID returns args.SessionID when set, confirming the session exists, or else
// the most recently completed session's ID. It never returns an in-progress or failed session as
// the default, per P03-T02's requirement.
func resolveQuerySessionID(st *store.Store, args models.QueryResultsArgs) (string, error) {
	if args.SessionID != "" {
		if _, err := st.GetSession(args.SessionID); err != nil {
			return "", fmt.Errorf("session %q not found: %w", args.SessionID, err)
		}
		return args.SessionID, nil
	}

	sessions, err := st.ListSessions(store.SessionFilter{Status: store.SessionStatusCompleted, Limit: 1})
	if err != nil {
		return "", fmt.Errorf("failed to find latest completed session: %w", err)
	}
	if len(sessions) == 0 {
		return "", fmt.Errorf("no completed session found; run a scan first")
	}
	return sessions[0].ID, nil
}

// queryResultsHandler serves the "azqr-query-results" MCP tool: a bounded, paginated read of a
// past session's resources, recommendation/graph results, or persisted summary, backed directly
// by internal/store (P03-T01's paginated loaders and P03-T03's summary loader). It never returns
// a full *renderers.ReportData or a file blob.
func queryResultsHandler(_ context.Context, _ mcp.CallToolRequest, args models.QueryResultsArgs) (*mcp.CallToolResult, error) {
	st := store.OpenForScan("")
	if st == nil {
		return mcp.NewToolResultError("session store unavailable; cannot query past scan results"), nil
	}
	defer func() { _ = st.Close() }()

	sessionID, err := resolveQuerySessionID(st, args)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to resolve session", err), nil
	}

	resultType := args.ResultType
	if resultType == "" {
		resultType = queryResultTypeResources
	}

	switch resultType {
	case queryResultTypeResources:
		limit := store.ClampResultPageSize(args.Limit)
		offset := store.ClampResultPageOffset(args.Offset)
		items, hasMore, err := st.LoadResourcesPage(sessionID, args.Limit, args.Offset)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to load resources page", err), nil
		}
		return mcp.NewToolResultStructured(
			newPagedResult(sessionID, resultType, items, limit, offset, hasMore),
			fmt.Sprintf("Loaded %d resource(s) for session %s (hasMore=%t)", len(items), sessionID, hasMore),
		), nil

	case queryResultTypeRecommendations:
		limit := store.ClampResultPageSize(args.Limit)
		offset := store.ClampResultPageOffset(args.Offset)
		items, hasMore, err := st.LoadGraphResultsPage(sessionID, args.Limit, args.Offset)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to load recommendation results page", err), nil
		}
		return mcp.NewToolResultStructured(
			newPagedResult(sessionID, resultType, items, limit, offset, hasMore),
			fmt.Sprintf("Loaded %d recommendation result(s) for session %s (hasMore=%t)", len(items), sessionID, hasMore),
		), nil

	case queryResultTypeSummary:
		summary, err := st.LoadSessionSummary(sessionID)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to load session summary", err), nil
		}
		if summary == nil {
			return mcp.NewToolResultError(fmt.Sprintf("no summary available for session %q (recorded before summary persistence, or the scan never finalized)", sessionID)), nil
		}
		return mcp.NewToolResultStructured(
			&queryResultsPage{SessionID: sessionID, ResultType: resultType, Summary: summary},
			fmt.Sprintf("Loaded summary for session %s", sessionID),
		), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf(
			"unknown resultType %q; want one of %q, %q, %q",
			args.ResultType, queryResultTypeResources, queryResultTypeRecommendations, queryResultTypeSummary,
		)), nil
	}
}

// newPagedResult builds the common pagination envelope shared by the "resources" and
// "recommendations" resultType branches.
func newPagedResult(sessionID, resultType string, items any, limit, offset int, hasMore bool) *queryResultsPage {
	page := &queryResultsPage{
		SessionID:  sessionID,
		ResultType: resultType,
		Items:      items,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}
	if hasMore {
		next := offset + limit
		page.NextOffset = &next
	}
	return page
}
