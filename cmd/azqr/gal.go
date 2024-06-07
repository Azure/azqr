// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/gal"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(galCmd)
}

var galCmd = &cobra.Command{
	Use:   "gal",
	Short: "Scan Azure Galleries",
	Long:  "Scan Azure Galleries",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&gal.GalleryScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
