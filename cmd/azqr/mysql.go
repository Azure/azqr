// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/mysql"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(mysqlCmd)
}

var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Scan Azure Database for MySQL",
	Long:  "Scan Azure Database for MySQL",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&mysql.MySQLScanner{},
			&mysql.MySQLFlexibleScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
