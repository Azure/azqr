// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/Azure/azqr/internal/mcpserver"
	"github.com/spf13/cobra"
)

var (
	mcpMode string
	mcpAddr string
)

func init() {
	mcpCmd.Flags().StringVar(&mcpMode, "mode", "stdio", "Server mode: stdio (default) or http")
	mcpCmd.Flags().StringVar(&mcpAddr, "addr", ":8080", "Address to listen on (only used in HTTP mode)")
	rootCmd.AddCommand(mcpCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server",
	Long: `Start the MCP server in stdio or HTTP/SSE mode.

Examples:
  # Start in stdio mode (default)
  azqr mcp
  
  # Start in HTTP/SSE mode on default port :8080
  azqr mcp --mode http
  
  # Start in HTTP/SSE mode on custom port
  azqr mcp --mode http --addr :3000`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		mode := mcpserver.ServerMode(mcpMode)
		mcpserver.StartWithMode(mode, mcpAddr)
	},
}
