// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/ba"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(baCmd)
}

var baCmd = &cobra.Command{
	Use:   "ba",
	Short: "Scan Azure Batch Account",
	Long:  "Scan Azure Batch Account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&ba.BatchAccountScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
