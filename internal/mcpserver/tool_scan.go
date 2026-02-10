// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func scanHandler(ctx context.Context, request mcp.CallToolRequest, args models.ScanArgs) (*mcp.CallToolResult, error) {
	currentDir, err := getCurrentFolder(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current working directory")
	}

	params := models.NewScanParamsWithDefaults(args)
	params.Xlsx = true
	params.Json = true
	params.OutputName = currentDir + "/azqr_scan_results"

	scanner := pipeline.Scanner{}
	r := scanner.Scan(params)

	fileName := r.OutputFileName + ".xlsx"
	uri := fmt.Sprintf("file://%s", fileName)
	uriJSON := fmt.Sprintf("file://%s.json", r.OutputFileName)

	registerScanResources(r.OutputFileName, "Azure Quick Review Scan Results", uriJSON, uri)

	resultText := fmt.Sprintf("Scan results saved to:\n- Excel: %s\n- JSON: %s", uri, uriJSON)
	return mcp.NewToolResultStructured(r, resultText), nil
}

// ExecuteScanTool is a public wrapper for scanHandler that can be called from other packages
func ExecuteScanTool(ctx context.Context, request mcp.CallToolRequest, args models.ScanArgs) (*mcp.CallToolResult, error) {
	return scanHandler(ctx, request, args)
}
