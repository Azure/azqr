// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/vnet"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vnetCmd)
}

var vnetCmd = &cobra.Command{
	Use:   "vnet",
	Short: "Scan Azure Virtual Network",
	Long:  "Scan Azure Virtual Network",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&vnet.VirtualNetworkScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
