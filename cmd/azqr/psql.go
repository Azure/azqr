// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/psql"
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
		serviceScanners := []azqr.IAzureScanner{
			&psql.PostgreScanner{},
			&psql.PostgreFlexibleScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
