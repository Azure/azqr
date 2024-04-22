// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/vpng"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vpngCmd)
}

var vpngCmd = &cobra.Command{
	Use:   "vpng",
	Short: "Scan Azure VPN Gateway",
	Long:  "Scan Azure VPN Gateway",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&vpng.VPNGatewayScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
