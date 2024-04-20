// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/syndp"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(syndpCmd)
}

var syndpCmd = &cobra.Command{
	Use:   "syndp",
	Short: "Scan Azure Synapse Dedicated SQL Pool",
	Long:  "Scan Azure Synapse Dedicated SQL Pool",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&syndp.SynapseSqlPoolScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
