// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
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
		scan(cmd, []string{"dbw"})
	},
}
