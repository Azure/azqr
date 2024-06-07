// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/erc"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(ercCmd)
}

var ercCmd = &cobra.Command{
	Use:   "erc",
	Short: "Scan Express Route Circuits",
	Long:  "Scan Express Route Circuits",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&erc.ExpressRouteScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
