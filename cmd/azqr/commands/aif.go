// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(cogCmd)
}

var cogCmd = &cobra.Command{
	Use:   "aif",
	Short: "Scan Azure AI Foundry and Cognitive Services",
	Long:  "Scan Azure AI and Cognitive Services",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"aif"})
	},
}
