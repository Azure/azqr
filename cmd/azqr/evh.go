// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/evh"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(evhCmd)
}

var evhCmd = &cobra.Command{
	Use:   "evh",
	Short: "Scan Azure Event Hubs",
	Long:  "Scan Azure Event Hubs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&evh.EventHubScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
