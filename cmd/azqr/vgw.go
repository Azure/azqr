// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/vgw"
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
		serviceScanners := []azqr.IAzureScanner{
			&vgw.VirtualNetworkGatewayScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
