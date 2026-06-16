// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

var s *server.MCPServer

// ServerMode defines the transport mode for the MCP server
type ServerMode string

const (
	ModeStdio ServerMode = "stdio"
	ModeHTTP  ServerMode = "http"
)

// StartWithMode starts the MCP server with the specified mode and address
// mode: "stdio" for standard input/output, "http" for HTTP/SSE
// addr: address to listen on (only used for HTTP mode, e.g., ":8080")
// version: version of the MCP server
func StartWithMode(mode ServerMode, addr, version string) {
	if version == "" {
		version = "dev"
	}

	s = server.NewMCPServer(
		"Azure Quick Review 🚀",
		version,
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
		// Start Streamable HTTP server (MCP 2025-03 spec)
		fmt.Printf("Starting MCP server in Streamable HTTP mode on %s\n", addr)
		httpSrv, mcpHTTPServer := buildHTTPServer(s)
		if err := mcpHTTPServer.Start(addr); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
		_ = httpSrv
	case ModeStdio:
		fallthrough
	default:
		// Start stdio server (default)
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Stdio server error: %v\n", err)
		}
	}
}

// buildHTTPServer constructs the http.Server and StreamableHTTPServer for HTTP/SSE mode.
//
// Design rationale (Azure Container Apps / Envoy compatibility):
//   - WriteTimeout=0: a non-zero value terminates long-lived SSE streams mid-flight.
//   - X-Accel-Buffering: no / Cache-Control: no-cache: Envoy buffers SSE responses
//     by default; these headers force pass-through so large StructuredContent payloads
//     (e.g. full scan results) are streamed to the client immediately rather than being
//     held in the proxy buffer until the connection closes.
//   - WithHeartbeatInterval: periodic SSE ping comments prevent intermediate proxies
//     from dropping idle connections during tool execution.
func buildHTTPServer(mcpSrv *server.MCPServer) (*http.Server, *server.StreamableHTTPServer) {
	customSrv := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	mcpHTTPServer := server.NewStreamableHTTPServer(mcpSrv,
		server.WithHeartbeatInterval(15*time.Second),
		server.WithStreamableHTTPServer(customSrv),
	)

	mux := http.NewServeMux()
	mux.Handle("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Cache-Control", "no-cache")
		mcpHTTPServer.ServeHTTP(w, r)
	}))
	customSrv.Handler = mux

	return customSrv, mcpHTTPServer
}
