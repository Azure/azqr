// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/nsg"
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
		serviceScanners := []azqr.IAzureScanner{
			&nsg.NSGScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
