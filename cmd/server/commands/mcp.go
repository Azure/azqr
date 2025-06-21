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

type ScanArgs struct {
	// Name      string   `json:"name"`
	// Age       int      `json:"age"`
	// IsVIP     bool     `json:"is_vip"`
	Services []string `json:"services"`
	// Metadata  struct {
	// 	Location string `json:"location"`
	// 	Timezone string `json:"timezone"`
	// } `json:"metadata"`
}

// mcp starts the MCP server using mark3labs/mcp-go.
// It registers the azqr tools and serves requests over stdio.
func run(cmd *cobra.Command) {
	s = server.NewMCPServer(
		"Azure Quick Review 🚀",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	types := mcp.NewTool("types",
		mcp.WithDescription(`List all supported recommendations. This command returns details of the Recommendations
			supported by Azure Quick Review (azqr). Use this to explore recommendations per id, category, impact and resource type.`),
	)

	// Add tool handler
	s.AddTool(types, typesHandler)

	r := mcp.NewTool("recommendations",
		mcp.WithDescription(`List all supported recommendations. This command returns details of the Recommendations
			supported by Azure Quick Review (azqr). Use this to explore recommendations per id, category, impact and resource type.`),
	)

	s.AddTool(r, recommendationsHandler)

	// Create a new resource to get the current working directory
	currentFolder := mcp.NewTool("current-folder",
		mcp.WithDescription("Returns the current working directory of the server process."),
	)

	// Register the resource and its handler
	s.AddTool(currentFolder, currentFolderHandler)

	scan := mcp.NewTool("scan",
		mcp.WithDescription(`Run an Azure Quick Review (azqr) scan for a given reource type.`),
		mcp.WithArray("services",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Required(),
			mcp.Description("Type abbreviation of the resource type to scan, e.g. 'aks' for Azure Kubernetes Service."),
		),
	)

	s.AddTool(scan, mcp.NewTypedToolHandler(scanHandler))

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func typesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	st := renderers.SupportedTypes{}
	output := st.GetAll()

	return mcp.NewToolResultText(output), nil
}

func recommendationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output := renderers.GetAllRecommendations(true)

	return mcp.NewToolResultText(output), nil
}

func currentFolderHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currentDir, err := getCurrentFolder()
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(currentDir), nil
}

func scanHandler(ctx context.Context, request mcp.CallToolRequest, args ScanArgs) (*mcp.CallToolResult, error) {
	// go func() {
	currentDir, err := getCurrentFolder()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get current working directory: %w", err))
	}

	scannerKeys := args.Services
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
	r := scanner.Scan(params)

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

	return mcp.NewToolResultText(r), nil
}

func getCurrentFolder() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return currentDir, nil
}
