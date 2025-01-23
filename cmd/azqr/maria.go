// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(mariaCmd)
}

var mariaCmd = &cobra.Command{
	Use:   "maria",
	Short: "Scan Azure Database for MariaDB",
	Long:  "Scan Azure Database for MariaDB",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"maria"})
	},
}
