// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(fdfpCmd)
}

var fdfpCmd = &cobra.Command{
	Use:   "fdfp",
	Short: "Scan Front Door Web Application Policy",
	Long:  "Scan Front Door Web Application Policy",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["fdfp"])
	},
}
