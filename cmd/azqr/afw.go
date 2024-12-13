// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(afwCmd)
}

var afwCmd = &cobra.Command{
	Use:   "afw",
	Short: "Scan Azure Firewall",
	Long:  "Scan Azure Firewall",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["afw"])
	},
}
