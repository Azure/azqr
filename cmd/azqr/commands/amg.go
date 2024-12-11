// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(amgCmd)
}

var amgCmd = &cobra.Command{
	Use:   "amg",
	Short: "Scan Azure Managed Grafana",
	Long:  "Scan Azure Managed Grafana",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"amg"})
	},
}
