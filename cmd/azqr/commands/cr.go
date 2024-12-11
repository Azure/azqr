// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(crCmd)
}

var crCmd = &cobra.Command{
	Use:   "cr",
	Short: "Scan Azure Container Registries",
	Long:  "Scan Azure Container Registries",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"cr"})
	},
}
