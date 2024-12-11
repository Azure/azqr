// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vmCmd)
}

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Scan Virtual Machine",
	Long:  "Scan Virtual Machine",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"vm"})
	},
}
