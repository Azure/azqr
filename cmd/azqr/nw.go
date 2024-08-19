// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/nw"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(nwCmd)
}

var nwCmd = &cobra.Command{
	Use:   "nw",
	Short: "Scan Network Watcher",
	Long:  "Scan Network Watcher",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&nw.NetworkWatcherScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
