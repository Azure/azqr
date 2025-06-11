// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(srchCmd)
}

var srchCmd = &cobra.Command{
	Use:   "srch",
	Short: "Scan Azure AI Search",
	Long:  "Scan Azure AI Search",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"srch"})
	},
}
