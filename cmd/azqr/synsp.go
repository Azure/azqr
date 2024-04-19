// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/synsp"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(synspCmd)
}

var synspCmd = &cobra.Command{
	Use:   "synsp",
	Short: "Scan Azure Synapse Spark Pool",
	Long:  "Scan Azure Synapse Spark Pool",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&synsp.SynapseSparkPoolScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
