// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/sigr"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(sigrCmd)
}

var sigrCmd = &cobra.Command{
	Use:   "sigr",
	Short: "Scan Azure SignalR",
	Long:  "Scan Azure SignalR",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&sigr.SignalRScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
