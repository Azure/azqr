// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/cosmos"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(cosmosCmd)
}

var cosmosCmd = &cobra.Command{
	Use:   "cosmos",
	Short: "Scan Azure Cosmos DB",
	Long:  "Scan Azure Cosmos DB",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&cosmos.CosmosDBScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
