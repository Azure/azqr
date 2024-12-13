// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(rsvCmd)
}

var rsvCmd = &cobra.Command{
	Use:   "rsv",
	Short: "Scan Recovery Service",
	Long:  "Scan Recovery Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["rsv"])
	},
}
