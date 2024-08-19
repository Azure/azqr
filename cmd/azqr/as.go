// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/as"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(asCmd)
}

var asCmd = &cobra.Command{
	Use:   "as",
	Short: "Scan Azure Analysis Service",
	Long:  "Scan Azure Analysis Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&as.AnalysisServicesScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
