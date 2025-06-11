// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(hubCmd)
}

var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Scan AI Foundry Hub and Azure Machine Learning Workspaces",
	Long:  "Scan AI Foundry Hub and Azure Machine Learning Workspaces",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"hub"})
	},
}
