// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(nsgCmd)
}

var nsgCmd = &cobra.Command{
	Use:   "nsg",
	Short: "Scan NSG",
	Long:  "Scan NSG",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"nsg"})
	},
}
