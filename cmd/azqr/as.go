// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/as"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(asCmd)
}

var asCmd = &cobra.Command{
	Use:   "as",
	Short: "Scan Azure Analytics Service",
	Long:  "Scan Azure Analytics Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&as.AnalysisServicesScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
