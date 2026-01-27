// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type PluginScanArgs struct {
	Mask     *bool    `json:"mask,omitempty"`
	Services []string `json:"services,omitempty"`
}

// scanPluginHandler creates a handler for plugin-specific scans
func scanPluginHandler(pluginName string) func(context.Context, mcp.CallToolRequest, PluginScanArgs) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest, args PluginScanArgs) (*mcp.CallToolResult, error) {
		currentDir, err := getCurrentFolder(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get current working directory")
		}

		// Use plugin-only mode by setting scannerKeys to the specific plugin
		scannerKeys := args.Services
		filters := models.LoadFilters("", scannerKeys)
		params := models.NewScanParams()

		// Plugin-only mode: disable defender, advisor, cost, policy, arc
		params.Defender = false
		params.Advisor = false
		params.Cost = false
		params.Policy = false
		params.Arc = false
		params.Mask = true

		// Override mask if provided
		if args.Mask != nil {
			params.Mask = *args.Mask
		}

		params.Xlsx = true
		params.Json = true
		params.ScannerKeys = scannerKeys
		params.Filters = filters
		params.OutputName = fmt.Sprintf("%s/azqr_%s_results", currentDir, pluginName)

		// Enable the specific plugin for execution
		params.EnabledInternalPlugins = map[string]bool{
			pluginName: true,
		}

		scanner := internal.Scanner{}
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
		resultText := fmt.Sprintf("%s\n\nScan results saved to:\n- Excel: %s\n- JSON: %s", r, uri, uriJSON)
		return mcp.NewToolResultText(resultText), nil
	}
}
