// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(stCmd)
}

var stCmd = &cobra.Command{
	Use:   "st",
	Short: "Scan Azure Storage",
	Long:  "Scan Azure Storage",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["st"])
	},
}
