// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/dbw"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(dbwCmd)
}

var dbwCmd = &cobra.Command{
	Use:   "dbw",
	Short: "Scan Azure Databricks",
	Long:  "Scan Azure Databricks",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&dbw.DatabricksScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
