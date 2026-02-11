// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// scanPluginHandler creates a handler for plugin-specific scans
func scanPluginHandler(pluginName string) func(context.Context, mcp.CallToolRequest, models.PluginScanArgs) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest, args models.PluginScanArgs) (*mcp.CallToolResult, error) {
		currentDir, err := getCurrentFolder(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get current working directory")
		}

		params := models.NewScanParamsForPlugins(args)
		params.Xlsx = true
		params.Json = true
		params.OutputName = fmt.Sprintf("%s/azqr_%s_results", currentDir, pluginName)

		// Enable the specific plugin for execution
		params.EnabledInternalPlugins = map[string]bool{
			pluginName: true,
		}

		scanner := pipeline.Scanner{}
		r := scanner.ScanPlugins(params)

		fileName := params.OutputName + ".xlsx"
		uri := fmt.Sprintf("file://%s", fileName)
		uriJSON := fmt.Sprintf("file://%s.json", params.OutputName)

		// Register the scan results as a resource
		jsonResults := mcp.NewResource(
			uriJSON,
			fmt.Sprintf("Azure Quick Review %s Scan Results (JSON)", pluginName),
			mcp.WithResourceDescription(fmt.Sprintf(`The results of the Azure Quick Review (azqr) %s scan (JSON).`, pluginName)),
			mcp.WithMIMEType("application/json"),
		)

		jsonBlob, err := os.ReadFile(params.OutputName + ".json")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read scan results JSON file")
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
			fmt.Sprintf("Azure Quick Review %s Scan Results", pluginName),
			mcp.WithResourceDescription(fmt.Sprintf(`The results of the Azure Quick Review (azqr) %s scan.`, pluginName)),
			mcp.WithMIMEType("binary/octet-stream"),
		)

		fileBlob, err := os.ReadFile(fileName)
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
}
