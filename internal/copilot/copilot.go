// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package copilot

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azqr/internal/copilot/config"
	"github.com/Azure/azqr/internal/copilot/tui"
	copilotSdk "github.com/github/copilot-sdk/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Run(copilotModel, copilotResume string) error {
	// Configure zerolog to discard logs (we use TUI for output)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(io.Discard)

	// Load user configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	cfg.Model = copilotModel

	// Set up graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	client := copilotSdk.NewClient(nil)
	if err := client.Start(context.Background()); err != nil {
		fmt.Println("❌ Failed to start Copilot client:", err)
		fmt.Println()
		fmt.Println("⚠️  Prerequisites:")
		fmt.Println("   1. Install GitHub CLI: https://cli.github.com/")
		fmt.Println("   2. Authenticate: gh auth login")
		fmt.Println("   3. Verify Copilot subscription at: https://github.com/settings/copilot")
		fmt.Println()
		return err
	}

	defer func() {
		if err := client.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}()

	authStatus, err := client.GetAuthStatus(context.Background())
	if err != nil || !authStatus.IsAuthenticated {
		fmt.Fprintln(os.Stderr, "❌ Not authenticated with GitHub.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Run 'gh auth login' and try again.")
		return nil
	}

	// Handle graceful shutdown on signal
	go func() {
		<-sigChan
		_, _ = fmt.Fprintln(os.Stdout, "\nShutting down...")
		_ = client.Stop()
		os.Exit(0)
	}()

	// Build common session configuration
	systemMessage := &copilotSdk.SystemMessageConfig{
		Content: `You are an Azure Quick Review (azqr) assistant that helps users analyze Azure resources for compliance, security, and best practices.

Your capabilities:
- Run compliance scans on Azure subscriptions and resource groups
- Provide recommendations catalog for Azure services
- List supported Azure services
- Analyze carbon emissions, OpenAI throttling, and availability zone mappings
- Access Microsoft Learn documentation for Azure best practices and guidance

When users ask about Azure resources, use the available tools to provide accurate, actionable insights. Always explain findings clearly and suggest remediation steps when issues are found. Use the Microsoft Learn MCP server to fetch official documentation when needed.`,
	}

	mcpServers := map[string]copilotSdk.MCPServerConfig{
		"microsoft-learn": {
			"type": "http",
			"url":  "https://learn.microsoft.com/api/mcp",
		},
	}

	infiniteSessions := &copilotSdk.InfiniteSessionConfig{
		Enabled:                       copilotSdk.Bool(true),
		BackgroundCompactionThreshold: copilotSdk.Float64(0.80),
		BufferExhaustionThreshold:     copilotSdk.Float64(0.95),
	}

	permissionHandler := buildPermissionHandler(cfg)
	tools := BuildTools()

	var session *copilotSdk.Session
	var infoMessages []string

	// Resume a previous session if --resume was specified
	if copilotResume != "" {
		resumeConfig := &copilotSdk.ResumeSessionConfig{
			Model:               cfg.Model,
			Tools:               tools,
			SystemMessage:       systemMessage,
			Streaming:           true,
			MCPServers:          mcpServers,
			InfiniteSessions:    infiniteSessions,
			OnPermissionRequest: permissionHandler,
		}

		session, err = client.ResumeSessionWithOptions(context.Background(), copilotResume, resumeConfig)
		if err != nil {
			infoMessages = append(infoMessages, fmt.Sprintf("Could not resume session %s: %v", copilotResume, err))
			session = nil
		} else {
			infoMessages = append(infoMessages, fmt.Sprintf("Resumed session %s", copilotResume))
		}
	}

	// Create a new session by default, or if resume failed
	if session == nil {
		sessionConfig := &copilotSdk.SessionConfig{
			Model:               cfg.Model,
			Tools:               tools,
			Streaming:           true,
			MCPServers:          mcpServers,
			SystemMessage:       systemMessage,
			InfiniteSessions:    infiniteSessions,
			OnPermissionRequest: permissionHandler,
		}

		session, err = client.CreateSession(context.Background(), sessionConfig)
		if err != nil {
			return fmt.Errorf("failed to create Copilot session: %w", err)
		}
		infoMessages = append(infoMessages, fmt.Sprintf("Session %s", session.SessionID))
	}

	defer func() {
		if err := session.Destroy(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to destroy session: %v\n", err)
		}
	}()

	// Run the TUI with alt screen for proper viewport control
	return tui.Run(cfg, client, session, infoMessages)
}

func buildPermissionHandler(_ *config.Config) copilotSdk.PermissionHandlerFunc {
	return func(_ copilotSdk.PermissionRequest, _ copilotSdk.PermissionInvocation) (copilotSdk.PermissionRequestResult, error) {
		return copilotSdk.PermissionRequestResult{Kind: "approved"}, nil
	}
}
