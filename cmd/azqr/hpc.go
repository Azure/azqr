// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(hpcCmd)
}

var hpcCmd = &cobra.Command{
	Use:   "hpc",
	Short: "Scan HPC",
	Long:  "Scan HPC",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, scanners.ScannerList["hpc"])
	},
}
