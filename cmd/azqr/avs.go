// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/avs"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(avsCmd)
}

var avsCmd = &cobra.Command{
	Use:   "avs",
	Short: "Scan Azure VMware Solution",
	Long:  "Scan Azure VMware Solution",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&avs.AVSScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
