// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(evgdCmd)
}

var evgdCmd = &cobra.Command{
	Use:   "evgd",
	Short: "Scan Azure Event Grid Domains",
	Long:  "Scan Azure Event Grid Domains",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"evgd"})
	},
}
