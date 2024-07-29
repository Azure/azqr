// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/conn"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(conCmd)
}

var conCmd = &cobra.Command{
	Use:   "com",
	Short: "Scan Connection",
	Long:  "Scan Connection",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&conn.ConnectionScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
