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
	Use:   "cog",
	Short: "Scan Azure Cognitive Service Accounts",
	Long:  "Scan Azure Cognitive Service Accounts",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"cog"})
	},
}
