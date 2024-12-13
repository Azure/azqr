// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(afdCmd)
}

var afdCmd = &cobra.Command{
	Use:   "afd",
	Short: "Scan Azure Front Door",
	Long:  "Scan Azure Front Door",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["afd"])
	},
}
