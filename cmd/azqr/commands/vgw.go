// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vgwCmd)
}

var vgwCmd = &cobra.Command{
	Use:   "vgw",
	Short: "Scan Virtual Network Gateway",
	Long:  "Scan Virtual Network Gateway",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"vgw"})
	},
}
