// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(synwCmd)
}

var synwCmd = &cobra.Command{
	Use:   "synw",
	Short: "Scan Azure Synapse Workspace",
	Long:  "Scan Azure Synapse Workspace",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"synw"})
	},
}
