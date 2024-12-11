// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
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
		supported by Azure Quick Review (azqr). Use this to explore recommendations per id, category, impact and resource type.`,
		func(arguments EmptyArguments) (*mcp_golang.ToolResponse, error) {
			output := renderers.GetAllRecommendations(true)
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(output)), nil
		},
	)
	if err != nil {
		panic(err)
	}

	err = server.RegisterTool(
		"scan",
		"Run an azqr scan for a given reource type.",
		func(arguments ScanArguments) (*mcp_golang.ToolResponse, error) {
			output := "Scan started, please wait for the results."

			go func() {
				scannerKeys := []string{arguments.ServiceKey}
				filters := models.LoadFilters("", scannerKeys)
				params := internal.NewScanParams()
				params.Cost = false
				params.Defender = false
				params.Advisor = false
				params.ScannerKeys = scannerKeys
				params.Filters = filters
				scanner := internal.Scanner{}
				scanner.Scan(params)
			}()

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
