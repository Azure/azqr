// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/aa"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(aaCmd)
}

var aaCmd = &cobra.Command{
	Use:   "aa",
	Short: "Scan Azure Automation Account",
	Long:  "Scan Azure Automation Account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&aa.AutomationAccountScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
