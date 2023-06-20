// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/appcs"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(appcsCmd)
}

var appcsCmd = &cobra.Command{
	Use:   "appcs",
	Short: "Scan Azure App Configuration",
	Long:  "Scan Azure App Configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&appcs.AppConfigurationScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
