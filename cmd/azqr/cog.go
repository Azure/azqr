// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/cog"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(cogCmd)
}

var cogCmd = &cobra.Command{
	Use:   "cog",
	Short: "Scan Azure Cognitive Service Accounts",
	Long:  "Scan Azure Cognitive Service Accounts",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&cog.CognitiveScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
