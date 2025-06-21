// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

func TestServerModes(t *testing.T) {
	tests := []struct {
		name string
		mode ServerMode
	}{
		{
			name: "stdio mode constant",
			mode: ModeStdio,
		},
		{
			name: "http mode constant",
			mode: ModeHTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mode == "" {
				t.Errorf("ServerMode should not be empty")
			}
		})
	}
}

func TestMCPServerInitialization(t *testing.T) {
	// Test that we can initialize an MCP server instance
	testServer := server.NewMCPServer(
		"Test Server",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
		server.WithRoots(),
	)

	if testServer == nil {
		t.Fatal("Failed to create MCP server instance")
	}

	// Test registering prompts and tools doesn't panic
	RegisterPrompts(testServer)
	RegisterTools(testServer)
}

func TestSSEServerCreation(t *testing.T) {
	// Create a test MCP server
	testServer := server.NewMCPServer(
		"Test Server",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Test that we can create an SSE server instance
	sseServer := server.NewSSEServer(testServer)
	if sseServer == nil {
		t.Fatal("Failed to create SSE server instance")
	}

	// Test with custom options
	sseServerWithOpts := server.NewSSEServer(
		testServer,
		server.WithKeepAlive(true),
		server.WithKeepAliveInterval(30*time.Second),
	)
	if sseServerWithOpts == nil {
		t.Fatal("Failed to create SSE server with options")
	}
}
