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

type ScanArgs struct {
	Services []string `json:"services,omitempty"`
	Defender *bool    `json:"defender,omitempty"`
	Advisor  *bool    `json:"advisor,omitempty"`
	Cost     *bool    `json:"cost,omitempty"`
	Policy   *bool    `json:"policy,omitempty"`
	Arc      *bool    `json:"arc,omitempty"`
	Mask     *bool    `json:"mask,omitempty"`
}

func scanHandler(ctx context.Context, request mcp.CallToolRequest, args ScanArgs) (*mcp.CallToolResult, error) {
	currentDir, err := getCurrentFolder(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current working directory")
	}

	scannerKeys := args.Services
	filters := models.LoadFilters("", scannerKeys)
	params := internal.NewScanParams()

	// Override defaults with provided values
	if args.Defender != nil {
		params.Defender = *args.Defender
	}
	if args.Advisor != nil {
		params.Advisor = *args.Advisor
	}
	if args.Cost != nil {
		params.Cost = *args.Cost
	}
	if args.Policy != nil {
		params.Policy = *args.Policy
	}
	if args.Arc != nil {
		params.Arc = *args.Arc
	}
	if args.Mask != nil {
		params.Mask = *args.Mask
	}

	params.Xlsx = true
	params.Json = true
	params.ScannerKeys = scannerKeys
	params.Filters = filters
	params.OutputName = currentDir + "/azqr_scan_results"
	scanner := internal.Scanner{}
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
