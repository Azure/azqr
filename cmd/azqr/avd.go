// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(avdCmd)
}

var avdCmd = &cobra.Command{
	Use:   "avd",
	Short: "Scan Azure Virtual Desktop",
	Long:  "Scan Azure Virtual Desktop",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["avd"])
	},
}
