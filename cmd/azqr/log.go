// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/log"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(logCmd)
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Scan Log Analytics workspace",
	Long:  "Scan Log Analytics workspace",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&log.LogAnalyticsScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
