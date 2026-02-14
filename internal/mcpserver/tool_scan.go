// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

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

	fileName := params.OutputName + ".xlsx"
	uri := fmt.Sprintf("file://%s", fileName)
	uriJSON := fmt.Sprintf("file://%s.json", params.OutputName)

	// Register the scan results as a resource
	jsonResults := mcp.NewResource(
		uriJSON,
		"Azure Quick Review Scan Results Metadata",
		mcp.WithResourceDescription(`The metadata of the Azure Quick Review (azqr) scan for the specified resource type.`),
		mcp.WithMIMEType("application/json"),
	)

	jsonBlob, err := os.ReadFile(params.OutputName + ".json")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read scan results metadata file")
	}

	encodedJSONBlob := make([]byte, base64.StdEncoding.EncodedLen(len(jsonBlob)))
	base64.StdEncoding.Encode(encodedJSONBlob, jsonBlob)

	s.AddResource(jsonResults, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.BlobResourceContents{
				URI:      uriJSON,
				MIMEType: "application/json",
				Blob:     string(encodedJSONBlob),
			},
		}, nil
	})

	results := mcp.NewResource(
		uri,
		"Azure Quick Review Scan Results",
		mcp.WithResourceDescription(`The results of the Azure Quick Review (azqr) scan for the specified resource type.`),
		mcp.WithMIMEType("binary/octet-stream"),
	)

	fileBlob, err := os.ReadFile(filepath.Clean(fileName)) //nolint:gosec // fileName is generated internally by scan
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read scan results file")
	}

	encodedBlob := make([]byte, base64.StdEncoding.EncodedLen(len(fileBlob)))
	base64.StdEncoding.Encode(encodedBlob, fileBlob)

	s.AddResource(results, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.BlobResourceContents{
				URI:      uri,
				MIMEType: "binary/octet-stream",
				Blob:     string(encodedBlob),
			},
		}, nil
	})

	// Return both the scan results and the resource URIs
	resultText := fmt.Sprintf("Scan results saved to:\n- Excel: %s\n- JSON: %s", uri, uriJSON)
	result := mcp.NewToolResultStructured(r, resultText)
	return result, nil
}
