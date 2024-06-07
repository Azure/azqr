// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/pdnsz"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(pdnszCmd)
}

var pdnszCmd = &cobra.Command{
	Use:   "pdnsz",
	Short: "Scan Private DNS Zone",
	Long:  "Scan Private DNS Zone",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&pdnsz.PrivateDNSZoneScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
