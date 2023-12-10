// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/ca"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(caCmd)
}

var caCmd = &cobra.Command{
	Use:   "ca",
	Short: "Scan Azure Container Apps",
	Long:  "Scan Azure Container Apps",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&ca.ContainerAppsScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
