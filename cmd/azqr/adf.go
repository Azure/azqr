// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(adfCmd)
}

var adfCmd = &cobra.Command{
	Use:   "adf",
	Short: "Scan Azure Data Factory",
	Long:  "Scan Azure Data Factory",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["adf"])
	},
}
