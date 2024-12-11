// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(psqlCmd)
}

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Scan Azure Database for psql",
	Long:  "Scan Azure Database for psql",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"psql"})
	},
}
