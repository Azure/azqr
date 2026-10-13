// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func scanHandler(ctx context.Context, request mcp.CallToolRequest, args models.ScanArgs) (*mcp.CallToolResult, error) {
	currentDir, err := getCurrentFolder(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current working directory")
	}

	params := models.NewScanParamsWithDefaults(args)
	params.OutputName = currentDir + "/azqr_scan_results"

	var sess *store.Session
	r, err := runPersistedScanMCP(store.SessionKindFull, params, func(sc *pipeline.Scanner) (*findings.Summary, error) {
		return sc.Scan(params)
	}, func(s *store.Session) { sess = s })
	if err != nil {
		return nil, err
	}

	resultText := scanResultText(sess, params.OutputName, params.Xlsx, params.Json)
	result := mcp.NewToolResultStructured(scanSummaryEnvelope(sess, r), resultText)
	return result, nil
}

// runPersistedScanMCP opens the session store (store.DefaultPath()) and runs scanFn through
// store.RunPersistedScan so the MCP-triggered scan becomes a persisted session. If the store
// cannot be opened, it logs a warning and runs scanFn against a plain *pipeline.Scanner instead,
// so store unavailability never blocks an otherwise successful scan and its file exports. In
// that fallback case onSessionCreated is never called, since no session was persisted.
func runPersistedScanMCP(
	kind store.SessionKind,
	params *models.ScanParams,
	scanFn func(*pipeline.Scanner) (*findings.Summary, error),
	onSessionCreated func(*store.Session),
) (*findings.Summary, error) {
	st := store.OpenForScan("")
	if st == nil {
		return scanFn(&pipeline.Scanner{})
	}
	defer func() { _ = st.Close() }()

	return st.RunPersistedScan(kind, store.SessionTriggerMCP, params, azqrVersion, scanFn, onSessionCreated)
}
