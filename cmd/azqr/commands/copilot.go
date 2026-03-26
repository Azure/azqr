// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/Azure/azqr/internal/copilot"

	"github.com/spf13/cobra"
)

const defaultCopilotModel = "claude-sonnet-4.6"

var (
	copilotModel  string
	copilotPrompt string
)

func init() {
	copilotCmd.Flags().StringVar(&copilotModel, "model", defaultCopilotModel, "Copilot model to use (default: Claude Sonnet 4.6 Medium)")
	copilotCmd.Flags().StringVarP(&copilotPrompt, "prompt", "p", "", "Run a single prompt and print the response")
	copilotCmd.AddCommand(copilotModelsCmd)
	rootCmd.AddCommand(copilotCmd)
}

var copilotModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available GitHub Copilot models",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copilot.ListModels()
	},
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

Examples:
  # Start interactive session
  azqr copilot

  # Run a single prompt
  azqr copilot -p "Scan my Azure subscription for compliance issues"

  # Use a specific model
  azqr copilot --model claude-sonnet-4.6`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copilot.Run(copilotModel, copilotPrompt)
	},
}
