// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/pep"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(pepCmd)
}

var pepCmd = &cobra.Command{
	Use:   "pep",
	Short: "Scan Private Endpoint",
	Long:  "Scan Private Endpoint",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&pep.PrivateEndpointScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
