// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/vm"
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
		serviceScanners := []azqr.IAzureScanner{
			&vm.VirtualMachineScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
