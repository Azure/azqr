// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/adx"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(adxCmd)
}

var adxCmd = &cobra.Command{
	Use:   "adx",
	Short: "Scan Azure Data Explorer",
	Long:  "Scan Azure Data Explorer",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&adx.DataExplorerScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
