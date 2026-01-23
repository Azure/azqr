// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azqr/internal/mcpserver"
	copilot "github.com/github/copilot-sdk/go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(copilotCmd)
}

var copilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "Interactive AI assistant powered by GitHub Copilot",
	Long: `Start a conversational AI session powered by GitHub Copilot to interact with azqr.

This command connects to GitHub Copilot and enables natural language interaction
with Azure Quick Review tools and capabilities.

Requirements:
  1. GitHub CLI installed (https://cli.github.com/)
  2. Authenticated: gh auth login
  3. GitHub Copilot subscription active

Available Tools:
  • scan - Run Azure resource compliance scans
  • get-recommendations-catalog - View APRL recommendations
  • get-supported-services - List supported Azure services

Examples:
  # Start AI assistant
  azqr copilot

  # Natural language queries:
  copilot> Scan my Azure subscription for compliance issues
  copilot> What are the recommendations for virtual machines?
  copilot> Which Azure services does azqr support?
  
Note: For VS Code integration, use the MCP server:
  azqr mcpserver`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runCopilot()
	},
}

func runCopilot() {
	fmt.Println("🤖 azqr Copilot - AI Assistant")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Connecting to GitHub Copilot...")

	// Initialize GitHub Copilot client
	client := copilot.NewClient(nil)
	if err := client.Start(); err != nil {
		fmt.Printf("❌ Failed to start Copilot client: %v\n", err)
		fmt.Println()
		fmt.Println("⚠️  Prerequisites:")
		fmt.Println("   1. Install GitHub CLI: https://cli.github.com/")
		fmt.Println("   2. Authenticate: gh auth login")
		fmt.Println("   3. Verify Copilot subscription at: https://github.com/settings/copilot")
		fmt.Println()
		return
	}
	defer client.Stop()

	// Create azqr tools for Copilot
	tools := createTools()

	// Create a conversation session
	session, err := client.CreateSession(&copilot.SessionConfig{
		Model:     "claude-sonnet-4.5",
		Tools:     tools,
		Streaming: true, // Enable streaming responses
	})
	if err != nil {
		fmt.Printf("❌ Failed to create session: %v\n", err)
		return
	}
	defer func() {
		if err := session.Destroy(); err != nil {
			fmt.Printf("Warning: Failed to destroy session: %v\n", err)
		}
	}()

	fmt.Println("✅ Connected to GitHub Copilot!")
	fmt.Println()
	fmt.Println("💡 You can now ask me anything about Azure Quick Review:")
	fmt.Println("   • \"Scan my Azure resources for compliance issues\"")
	fmt.Println("   • \"What are the security recommendations for VMs?\"")
	fmt.Println("   • \"Which services does azqr support?\"")
	fmt.Println()
	fmt.Println("Type 'exit' to quit")
	fmt.Println()

	// Buffered channel to signal when response is complete (prevents race conditions)
	responseDone := make(chan struct{}, 1)

	// Subscribe to session events for streaming responses
	responseBuffer := strings.Builder{}
	session.On(func(event copilot.SessionEvent) {
		switch event.Type {
		case "assistant.message_delta":
			// Streaming response - print incrementally
			if event.Data.DeltaContent != nil {
				fmt.Print(*event.Data.DeltaContent)
				responseBuffer.WriteString(*event.Data.DeltaContent)
			}
		case "assistant.message":
			// Final message chunk
			if responseBuffer.Len() == 0 && event.Data.Content != nil {
				fmt.Print(*event.Data.Content)
			}
			responseBuffer.Reset()
		case "session.idle":
			// Session is idle - turn is fully complete, signal to continue
			fmt.Println()
			fmt.Println()
			select {
			case responseDone <- struct{}{}:
			default:
			}
		case "session.error":
			if event.Data.Message != nil {
				fmt.Printf("\n❌ Error: %s\n\n", *event.Data.Message)
			}
			select {
			case responseDone <- struct{}{}:
			default:
			}
		case "tool.execution_start":
			if event.Data.ToolName != nil {
				fmt.Printf("\n⚙️  Executing %s...\n", *event.Data.ToolName)
			}
		case "tool.execution_complete":
			fmt.Println("✅ Tool execution complete")
		}
	})

	// Start interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			fmt.Println()
			fmt.Println("👋 Goodbye!")
			break
		}

		// Send message to Copilot
		fmt.Println()
		fmt.Print("🤖 Copilot: ")
		_, err := session.Send(copilot.MessageOptions{
			Prompt: input,
		})
		if err != nil {
			fmt.Printf("\n❌ Error: %v\n\n", err)
			continue
		}

		// Wait for response to complete before prompting for next input
		<-responseDone
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

// createTools creates Copilot tools from azqr capabilities
func createTools() []copilot.Tool {
	tools := []copilot.Tool{
		// Scan tool
		copilot.DefineTool("scan", "Run Azure Quick Review scan to analyze resources for compliance, security, and best practices",
			func(params ScanParams, inv copilot.ToolInvocation) (string, error) {
				ctx := context.Background()
				return executeScanTool(ctx, params)
			}),

		// Get recommendations catalog
		copilot.DefineTool("get-recommendations-catalog", "Get the complete catalog of azqr recommendations",
			func(params EmptyParams, inv copilot.ToolInvocation) (string, error) {
				return getRecommendationsCatalog()
			}),

		// Get supported services
		copilot.DefineTool("get-supported-services", "List all Azure services supported by azqr",
			func(params EmptyParams, inv copilot.ToolInvocation) (string, error) {
				return getSupportedServices()
			}),

		// Scan carbon emissions
		copilot.DefineTool("scan-carbon-emissions", "Analyze carbon emissions by Azure resource type with period-over-period tracking",
			func(params PluginParams, inv copilot.ToolInvocation) (string, error) {
				ctx := context.Background()
				return executePluginScanTool(ctx, "carbon-emissions", params)
			}),

		// Scan OpenAI throttling
		copilot.DefineTool("scan-openai-throttling", "Check OpenAI/Cognitive Services accounts for 429 throttling errors",
			func(params PluginParams, inv copilot.ToolInvocation) (string, error) {
				ctx := context.Background()
				return executePluginScanTool(ctx, "openai-throttling", params)
			}),

		// Scan zone mapping
		copilot.DefineTool("scan-zone-mapping", "Retrieve logical-to-physical availability zone mappings for all Azure regions",
			func(params PluginParams, inv copilot.ToolInvocation) (string, error) {
				ctx := context.Background()
				return executePluginScanTool(ctx, "zone-mapping", params)
			}),
	}

	return tools
}

// Tool parameter types
type ScanParams struct {
	Services []string `json:"services" jsonschema:"Optional array of service type abbreviations to scan (e.g., vm, sql, aks)"`
	Defender bool     `json:"defender" jsonschema:"Include Microsoft Defender scan"`
	Advisor  bool     `json:"advisor" jsonschema:"Include Azure Advisor recommendations"`
	Cost     bool     `json:"cost" jsonschema:"Include cost analysis"`
	Policy   bool     `json:"policy" jsonschema:"Include Azure Policy compliance scan"`
	Arc      bool     `json:"arc" jsonschema:"Include Azure Arc-enabled resources scan"`
	Mask     bool     `json:"mask" jsonschema:"Mask sensitive data in results (default: true)"`
}

type PluginParams struct {
	Services []string `json:"services" jsonschema:"Optional array of service type abbreviations to scan"`
	Mask     bool     `json:"mask" jsonschema:"Mask sensitive data in output (default: true)"`
}

type EmptyParams struct{}

// Tool execution functions - Delegate to MCP server implementations
func executeScanTool(ctx context.Context, params ScanParams) (string, error) {
	fmt.Println("🔍 Initiating Azure resource scan...")

	mcpArgs := mcpserver.ScanArgs{
		Services: params.Services,
		Defender: &params.Defender,
		Advisor:  &params.Advisor,
		Cost:     &params.Cost,
		Policy:   &params.Policy,
		Arc:      &params.Arc,
		Mask:     &params.Mask,
	}

	result, err := mcpserver.ExecuteScanTool(ctx, mcp.CallToolRequest{}, mcpArgs)
	if err != nil {
		return "", fmt.Errorf("scan failed: %w", err)
	}

	return extractMCPResultText(result), nil
}

func getRecommendationsCatalog() (string, error) {
	result, err := mcpserver.ExecuteCatalogTool(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get catalog: %w", err)
	}
	return extractMCPResultText(result), nil
}

func getSupportedServices() (string, error) {
	result, err := mcpserver.ExecuteServicesTool(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get services: %w", err)
	}
	return extractMCPResultText(result), nil
}

// extractMCPResultText extracts text content from MCP tool result
func extractMCPResultText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return "✓ Operation completed successfully."
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return "✓ Operation completed successfully."
}

func executePluginScanTool(ctx context.Context, pluginName string, params PluginParams) (string, error) {
	fmt.Printf("🔍 Initiating %s scan...\n", pluginName)

	mask := params.Mask
	mcpArgs := mcpserver.PluginScanArgs{
		Services: params.Services,
		Mask:     &mask,
	}

	result, err := mcpserver.ExecutePluginScanTool(ctx, pluginName, mcpArgs)
	if err != nil {
		return "", fmt.Errorf("%s scan failed: %w", pluginName, err)
	}

	return extractMCPResultText(result), nil
}
