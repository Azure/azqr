// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vmssCmd)
}

var vmssCmd = &cobra.Command{
	Use:   "vmss",
	Short: "Scan Virtual Machine Scale Set",
	Long:  "Scan Virtual Machine Scale Set",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"vmss"})
	},
}
