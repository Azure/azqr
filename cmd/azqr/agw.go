// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(agwCmd)
}

var agwCmd = &cobra.Command{
	Use:   "agw",
	Short: "Scan Azure Application Gateway",
	Long:  "Scan Azure Application Gateway",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"agw"})
	},
}
