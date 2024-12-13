// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
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
		scan(cmd, scanners.ScannerList["ng"])
	},
}
