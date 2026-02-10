// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/Azure/azqr/internal/copilot"

	"github.com/spf13/cobra"
)

const defaultCopilotModel = "claude-sonnet-4.5"

var (
	copilotModel  string
	copilotResume string
)

func init() {
	copilotCmd.Flags().StringVar(&copilotModel, "model", defaultCopilotModel, "Copilot model to use during the session")
	copilotCmd.Flags().StringVar(&copilotResume, "resume", "", "Resume a previous session by ID")
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
  • get-recommendations-catalog - View azqr recommendations
  • get-supported-services - List supported Azure services

Examples:
  # Start AI assistant
  azqr copilot

  # Natural language queries:
  copilot> Scan my Azure subscription for compliance issues
  copilot> What are the recommendations for virtual machines?
  copilot> Which Azure services does azqr support?`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copilot.Run(copilotModel, copilotResume)
	},
}
