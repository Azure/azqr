// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// registerScanResources registers JSON and Excel scan results as MCP resources
func registerScanResources(outputName, resultName string, uriJSON, uriExcel string) {
	if s == nil {
		return
	}

	// Register JSON resource
	jsonResource := mcp.NewResource(
		uriJSON,
		resultName+" (JSON)",
		mcp.WithResourceDescription("The results of the Azure Quick Review (azqr) scan (JSON)."),
		mcp.WithMIMEType("application/json"),
	)

	encodedJSON := encodeFileBase64(outputName + ".json")
	s.AddResource(jsonResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.BlobResourceContents{
				URI:      uriJSON,
				MIMEType: "application/json",
				Blob:     encodedJSON,
			},
		}, nil
	})

	// Register Excel resource
	excelResource := mcp.NewResource(
		uriExcel,
		resultName,
		mcp.WithResourceDescription("The results of the Azure Quick Review (azqr) scan."),
		mcp.WithMIMEType("binary/octet-stream"),
	)

	encodedExcel := encodeFileBase64(outputName + ".xlsx")
	s.AddResource(excelResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.BlobResourceContents{
				URI:      uriExcel,
				MIMEType: "binary/octet-stream",
				Blob:     encodedExcel,
			},
		}, nil
	})
}

// encodeFileBase64 reads a file and returns its base64-encoded content
func encodeFileBase64(path string) string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("failed to read file for base64 encoding")
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}
