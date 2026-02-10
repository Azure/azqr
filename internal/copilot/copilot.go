// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package copilot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azqr/internal/copilot/config"
	"github.com/Azure/azqr/internal/copilot/tui"
	tea "github.com/charmbracelet/bubbletea"
	copilotSdk "github.com/github/copilot-sdk/go"
)

// ListModels connects to the Copilot backend and prints available models.
func ListModels() error {
	client := copilotSdk.NewClient(nil)
	if err := client.Start(context.Background()); err != nil {
		fmt.Println("Failed to start Copilot client:", err)
		fmt.Println()
		fmt.Println("Prerequisites:")
		fmt.Println("  1. Install GitHub CLI: https://cli.github.com/")
		fmt.Println("  2. Authenticate: gh auth login")
		fmt.Println("  3. Verify Copilot subscription at: https://github.com/settings/copilot")
		return err
	}
	defer func() {
		if err := client.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}()

	models, err := client.ListModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	fmt.Printf("%-45s %s\n", "ID", "NAME")
	fmt.Printf("%-45s %s\n", "--", "----")
	for _, m := range models {
		fmt.Printf("%-45s %s\n", m.ID, m.Name)
	}
	return nil
}

func Run(copilotModel, prompt string) error {
	// Load user configuration
	cfg := config.DefaultConfig()
	cfg.Model = copilotModel

	client := copilotSdk.NewClient(nil)
	if err := client.Start(context.Background()); err != nil {
		fmt.Println("Failed to start Copilot client:", err)
		fmt.Println()
		fmt.Println("Prerequisites:")
		fmt.Println("  1. Install GitHub CLI: https://cli.github.com/")
		fmt.Println("  2. Authenticate: gh auth login")
		fmt.Println("  3. Verify Copilot subscription at: https://github.com/settings/copilot")
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
		fmt.Fprintln(os.Stderr, "Not authenticated with GitHub.")
		fmt.Fprintln(os.Stderr, "Run 'gh auth login' and try again.")
		return nil
	}

	// Set up graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	systemMessage := &copilotSdk.SystemMessageConfig{
		Mode: "replace",
		Content: `You are AZQR Assistant, an AI-powered Azure governance advisor built on Azure Quick Review (azqr).

You are operating inside a dedicated AZQR experience, not a general coding session.

Core behavior:
- Act as a single merged Azure specialist that combines compliance scanning, recommendation analysis, supported-service coverage checks, carbon emissions analysis, Azure OpenAI throttling diagnostics, and availability zone mapping insights.
- Use the provided azqr tools directly when they help answer the request.
- Do not narrate internal tool-selection mechanics, parameter schema discovery, or chain-of-thought planning.
- Do not say things like "I have access to tool X" or "Let me inspect the parameters".
- When a scan or analysis request is clear, briefly state the action in user terms and perform it.
- When scope is missing or ambiguous, ask only the minimum clarifying question needed.

Post-scan behavior (critical):
- After running the "scan" tool, immediately analyze and present the findings — do NOT ask "would you like me to summarize?".
- Organize findings by severity: High → Medium → Low.
- For each group, list findings with resource name, recommendation ID, and a brief description.
- Call "get-recommendations-catalog" for remediation steps on the top High-severity items.
- Conclude with the 3 most important next actions.

Available tool intents:
- Use "scan" for Azure resource compliance, security, reliability, and best-practices assessments.
- Use "get-recommendations-catalog" to ground findings and remediation guidance in azqr's catalog.
- Use "get-supported-services" to confirm azqr coverage.
- Use "scan-carbon-emissions" for sustainability and emissions analysis.
- Use "scan-openai-throttling" for Azure OpenAI and Cognitive Services throttling diagnostics.
- Use "scan-zone-mapping" for logical-to-physical availability zone mapping and HA design analysis.

Response style:
- Be direct, architecture-focused, and proactive — deliver the analysis, do not defer it.
- Lead with the outcome, then detail findings and next steps.
- For scan requests, avoid preambles longer than one short sentence before using tools.
- Treat azqr as read-only advisory tooling; do not propose code edits unless explicitly requested.
- Reference Microsoft Learn when relevant through the configured MCP server.
`,
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
	agents := BuildAgents(mcpServers)

	// For interactive mode, allocate the program double-pointer early so
	// OnUserInputRequest can forward ask_user calls into the TUI.
	var progPtr **tea.Program
	var userInputHandler copilotSdk.UserInputHandler
	if prompt == "" {
		progPtr = new(*tea.Program)
		userInputHandler = tui.BuildUserInputHandler(progPtr)
	}

	sessionConfig := &copilotSdk.SessionConfig{
		Model:               cfg.Model,
		Tools:               tools,
		CustomAgents:        agents,
		Agent:               "azqr-assistant",
		Streaming:           true,
		MCPServers:          mcpServers,
		SystemMessage:       systemMessage,
		InfiniteSessions:    infiniteSessions,
		OnPermissionRequest: permissionHandler,
		OnUserInputRequest:  userInputHandler,
	}

	session, err := client.CreateSession(context.Background(), sessionConfig)
	if err != nil {
		return fmt.Errorf("failed to create Copilot session: %w", err)
	}

	defer func() {
		if err := session.Disconnect(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to disconnect session: %v\n", err)
		}
	}()

	go func() {
		<-sigChan
		_, _ = fmt.Fprintln(os.Stdout, "\nShutting down...")
		_ = session.Disconnect()
		_ = client.Stop()
		os.Exit(0)
	}()

	if prompt != "" {
		return RunSinglePrompt(session, prompt)
	}
	return RunInteractive(session, cfg.Model, progPtr)
}

func buildPermissionHandler(_ *config.Config) copilotSdk.PermissionHandlerFunc {
	return copilotSdk.PermissionHandler.ApproveAll
}
