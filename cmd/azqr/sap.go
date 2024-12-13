// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(sapCmd)
}

var sapCmd = &cobra.Command{
	Use:   "sap",
	Short: "Scan SAP",
	Long:  "Scan SAP",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["sap"])
	},
}
