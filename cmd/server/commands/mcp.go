// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/invopop/jsonschema"
	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
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
		mcp(cmd)
	},
}

type EmptyArguments struct {
	// No arguments needed
}

type ScanArguments struct {
	ServiceKey string `json:"key" jsonschema:"required,description=The abrreviation of the resource type to scan"`
}

func mcp(cmd *cobra.Command) {
	done := make(chan struct{})

	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	jsonschema.Version = "https://json-schema.org/draft-07/schema"

	// Register the new tool in the MCP server
	// Register the "types" tool in the MCP server with a multiline description.
	err := server.RegisterTool(
		"types",
		`List all supported resource types. This commnands returns details of the Azure Resource Types
		suported by Azure Quick rteview (azqr). Use this to explore supported services and their abbrevaitions.
		The output is a markdown table.`,
		func(arguments EmptyArguments) (*mcp_golang.ToolResponse, error) {
			// Create a SupportedTypes instance and get all supported types.
			st := renderers.SupportedTypes{}
			output := st.GetAll()
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(output)), nil
		},
	)
	if err != nil {
		panic(err)
	}

	err = server.RegisterTool(
		"recommendations",
		`List all supported recommendations. This command returns details of the Recommendations
		supported by Azure Quick Review (azqr). Use this to explore recommendations per id, category, impact and resource type.
		Provide the resource type abbreviation (e.g., "aks", "sql", "kv") as the "key" argument.`,
		func(arguments ScanArguments) (*mcp_golang.ToolResponse, error) {
			// Get the list of scanners for the provided service key.
			scanners := models.ScannerList[arguments.ServiceKey]
			rec := map[string]models.AzqrRecommendation{}

			// Iterate over each scanner and get recommendations.
			for _, sc := range scanners {
				r := sc.GetRecommendations() // You can process or collect recommendations as needed.
				for id, recommendation := range r {
					rec[id] = recommendation
				}
			}

			jsonBytes, err := json.Marshal(rec)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize recommendations: %w", err)
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonBytes))), nil
		},
	)
	if err != nil {
		panic(err)
	}

	// Register the "scan" tool in the MCP server.
	// This tool runs an azqr scan for a given Azure resource type.
	// The user must provide the resource type abbreviation as the "key" argument.
	err = server.RegisterTool(
		"scan",
		`Run an Azure Quick Review (azqr) scan for a given Azure resource type.
		Provide the resource type abbreviation (e.g., "aks", "sql", "kv") as the "key" argument.
		This will trigger a scan for the specified resource type. The scan runs asynchronously; 
		you will receive a confirmation message immediately, and results will be available once the scan completes.`,
		func(arguments ScanArguments) (*mcp_golang.ToolResponse, error) {
			// Inform the user that the scan has started.
			output := "Scan started for resource type '" + arguments.ServiceKey + "'. Please wait for the results."

			// Run the scan asynchronously.
			go func() {
				// Prepare scanner keys and filters.
				scannerKeys := []string{arguments.ServiceKey}
				filters := models.LoadFilters("", scannerKeys)

				// Set up scan parameters.
				params := internal.NewScanParams()
				params.Cost = false
				params.Defender = false
				params.Advisor = false
				params.ScannerKeys = scannerKeys
				params.Filters = filters

				// Create and run the scanner.
				scanner := internal.Scanner{}
				scanner.Scan(params)
			}()

			// Return immediate response to the user.
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(output)), nil
		},
	)
	if err != nil {
		panic(err)
	}

	err = server.Serve()

	fmt.Println("Server started, waiting for requests...")

	if err != nil {
		panic(err)
	}

	<-done
}
