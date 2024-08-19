// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/rt"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(rtCmd)
}

var rtCmd = &cobra.Command{
	Use:   "rt",
	Short: "Scan Route Table",
	Long:  "Scan Route Table",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&rt.RouteTableScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
