// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/ng"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(ngCmd)
}

var ngCmd = &cobra.Command{
	Use:   "ng",
	Short: "Scan Azure NAT Gateway",
	Long:  "Scan Azure NAT Gateway",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&ng.NatGatewayScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
