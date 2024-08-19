// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/hpc"
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
		serviceScanners := []azqr.IAzureScanner{
			&hpc.HighPerformanceComputingScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
