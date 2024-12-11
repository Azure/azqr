// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"

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

type Content struct {
	Title       string  `json:"title" jsonschema:"required,description=The title to submit"`
	Description *string `json:"description" jsonschema:"description=The description to submit"`
}

type SubmitterArguments struct {
	Submitter string  `json:"submitter" jsonschema:"required,description=The name of the thing calling this tool (openai, google, claude, etc)"`
	Content   Content `json:"content" jsonschema:"required,description=The content to submit"`
}

func mcp(cmd *cobra.Command) {
	done := make(chan struct{})

	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	jsonschema.Version = "https://json-schema.org/draft-07/schema"

	// Register the new tool in the MCP server
	err := server.RegisterTool(
		"azqr-types",
		"Retrieve or list all Azure resource types supported by the azqr tool",
		func(arguments SubmitterArguments) (*mcp_golang.ToolResponse, error) {
			st := renderers.SupportedTypes{}
			output := st.GetAll()
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(output)), nil
		},
	)
	if err != nil {
		panic(err)
	}

	err = server.RegisterTool(
		"azqr-rules",
		"Retrieve or list all recommendations checked by azqr",
		func(arguments SubmitterArguments) (*mcp_golang.ToolResponse, error) {
			output := renderers.GetAllRecommendations(false)
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
