// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

var s *server.MCPServer

// ServerMode defines the transport mode for the MCP server
type ServerMode string

const (
	ModeStdio ServerMode = "stdio"
	ModeHTTP  ServerMode = "http"
)

// Start initializes and starts the MCP server in stdio mode (default)
func Start() {
	StartWithMode(ModeStdio, ":8080")
}

// StartWithMode starts the MCP server with the specified mode and address
// mode: "stdio" for standard input/output, "http" for HTTP/SSE
// addr: address to listen on (only used for HTTP mode, e.g., ":8080")
func StartWithMode(mode ServerMode, addr string) {
	s = server.NewMCPServer(
		"Azure Quick Review 🚀",
		"0.1.0",
		server.WithToolCapabilities(true), // Enable tool notifications
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithRecovery(), // Graceful panic recovery
		server.WithRoots(),
	)

	RegisterPrompts(s)
	RegisterTools(s)

	switch mode {
	case ModeHTTP:
		// Start HTTP/SSE server
		fmt.Printf("Starting MCP server in HTTP/SSE mode on %s\n", addr)
		sseServer := server.NewSSEServer(s)
		if err := sseServer.Start(addr); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	case ModeStdio:
		fallthrough
	default:
		// Start stdio server (default)
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Stdio server error: %v\n", err)
		}
	}
}
