// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(logicCmd)
}

var logicCmd = &cobra.Command{
	Use:   "logic",
	Short: "Scan Azure Logic Apps",
	Long:  "Scan Azure Logic Apps",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"logic"})
	},
}
