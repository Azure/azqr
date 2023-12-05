// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/asp"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Scan Azure App Service",
	Long:  "Scan Azure App Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&asp.AppServiceScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
