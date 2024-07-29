// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/cae"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(caeCmd)
}

var caeCmd = &cobra.Command{
	Use:   "cae",
	Short: "Scan Azure Container Apps Environment",
	Long:  "Scan Azure Container Apps Environment",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&cae.ContainerAppsEnvironmentScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
