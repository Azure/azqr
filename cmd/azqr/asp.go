// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/asp"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "asp",
	Short: "Scan Azure App Service",
	Long:  "Scan Azure App Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&asp.AppServiceScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
