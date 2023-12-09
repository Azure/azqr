// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/traf"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(trafCmd)
}

var trafCmd = &cobra.Command{
	Use:   "traf",
	Short: "Scan Azure Traffic Manager",
	Long:  "Scan Azure Traffic Manager",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&traf.TrafficManagerScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
