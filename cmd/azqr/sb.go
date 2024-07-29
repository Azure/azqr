// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/sb"
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
		serviceScanners := []azqr.IAzureScanner{
			&sb.ServiceBusScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
