// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/cr"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(crCmd)
}

var crCmd = &cobra.Command{
	Use:   "cr",
	Short: "Scan Azure Container Registries",
	Long:  "Scan Azure Container Registries",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&cr.ContainerRegistryScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
