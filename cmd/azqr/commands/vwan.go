// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vwanCmd)
}

var vwanCmd = &cobra.Command{
	Use:   "vwan",
	Short: "Scan Azure Virtual WAN",
	Long:  "Scan Azure Virtual WAN",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"vwan"})
	},
}
