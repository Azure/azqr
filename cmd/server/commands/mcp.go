// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(mcpCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server",
	Long:  "Start the MCP server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd)
	},
}

var s *server.MCPServer

// mcp starts the MCP server using mark3labs/mcp-go.
// It registers the azqr tools and serves requests over stdio.
func run(cmd *cobra.Command) {
	s = server.NewMCPServer(
		"Azure Quick Review 🚀",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	types := mcp.NewResource(
		"docs://types",
		"Azure Quick Review Supported Types",
		mcp.WithResourceDescription(`List all supported recommendations. This command returns details of the Recommendations
			supported by Azure Quick Review (azqr). Use this to explore recommendations per id, category, impact and resource type.`),
		mcp.WithMIMEType("text/markdown"),
	)

	// Add tool handler
	s.AddResource(types, typesHandler)

	r := mcp.NewResource(
		"docs://recommendations",
		"Azure Quick Review Supported recommendations",
		mcp.WithResourceDescription(`List all supported recommendations. This command returns details of the Recommendations
			supported by Azure Quick Review (azqr). Use this to explore recommendations per id, category, impact and resource type.`),
		mcp.WithMIMEType("text/markdown"),
	)

	s.AddResource(r, recommendationsHandler)

	// Create a new resource to get the current working directory
	currentFolder := mcp.NewResource(
		"cli://current-folder",
		"Current Working Directory",
		mcp.WithResourceDescription("Returns the current working directory of the server process."),
		mcp.WithMIMEType("text/plain"),
	)

	// Register the resource and its handler
	s.AddResource(currentFolder, currentFolderHandler)

	scan := mcp.NewTool("scan",
		mcp.WithDescription(`Run an Azure Quick Review (azqr) scan for a given reource type.`),
		mcp.WithString("serviceKey",
			mcp.Required(),
			mcp.Description("Type abbreviation of the resource type to scan, e.g. 'aks' for Azure Kubernetes Service."),
		),
	)

	s.AddTool(scan, scanHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func typesHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	st := renderers.SupportedTypes{}
	output := st.GetAll()

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "docs://types",
			MIMEType: "text/markdown",
			Text:     output,
		},
	}, nil
}

func recommendationsHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	output := renderers.GetAllRecommendations(true)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "docs://recommendations",
			MIMEType: "text/markdown",
			Text:     output,
		},
	}, nil
}

func currentFolderHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	currentDir, err := getCurrentFolder()
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "cli://current-folder",
			MIMEType: "text/plain",
			Text:     currentDir,
		},
	}, nil
}

func scanHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceKey, err := request.RequireString("serviceKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// output := "Scan started, please wait for the results."

	// go func() {
	currentDir, err := getCurrentFolder()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get current working directory: %w", err))
	}

	scannerKeys := []string{serviceKey}
	filters := models.LoadFilters("", scannerKeys)
	params := internal.NewScanParams()
	params.Cost = false
	params.Xlsx = true
	params.Defender = false
	params.Advisor = false
	params.ScannerKeys = scannerKeys
	params.Filters = filters
	params.OutputName = currentDir + "/azqr_scan_results"
	scanner := internal.Scanner{}
	scanner.Scan(params)

	fileName := params.OutputName + ".xlsx"
	uri := fmt.Sprintf("file://%s", fileName)

	results := mcp.NewResource(
		uri,
		"Azure Quick Review Scan Results",
		mcp.WithResourceDescription(`The results of the Azure Quick Review (azqr) scan for the specified resource type.`),
		mcp.WithMIMEType("binary/octet-stream"),
	)

	fileBlob, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read scan results file: %w", err))
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

	// s.SendNotificationToClient(context.Background(),
	// 	"notification/update",
	// 	map[string]any{"message": fmt.Sprintf("New data available here: %s", uri)},
	// )
	// }()

	return mcp.NewToolResultText(fileName), nil
}

func getCurrentFolder() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return currentDir, nil
}
