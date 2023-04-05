// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/sb"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(sbCmd)
}

var sbCmd = &cobra.Command{
	Use:   "sb",
	Short: "Scan Azure Service Bus",
	Long:  "Scan Azure Service Bus",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&sb.ServiceBusScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
