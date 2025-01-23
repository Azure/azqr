// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(sqlCmd)
}

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Scan Azure SQL Database",
	Long:  "Scan Azure SQL Database",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"sql"})
	},
}
