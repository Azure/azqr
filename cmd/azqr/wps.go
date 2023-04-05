// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/wps"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(wpsCmd)
}

var wpsCmd = &cobra.Command{
	Use:   "wps",
	Short: "Scan Azure Web PubSub",
	Long:  "Scan Azure Web PubSub",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&wps.WebPubSubScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
