// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(aaCmd)
}

var aaCmd = &cobra.Command{
	Use:   "aa",
	Short: "Scan Azure Automation Account",
	Long:  "Scan Azure Automation Account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"aa"})
	},
}
