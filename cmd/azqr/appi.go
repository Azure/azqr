// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/appi"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(appiCmd)
}

var appiCmd = &cobra.Command{
	Use:   "appi",
	Short: "Scan Azure Application Insights",
	Long:  "Scan Azure Application Insights",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&appi.AppInsightsScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
