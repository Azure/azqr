// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(caCmd)
}

var caCmd = &cobra.Command{
	Use:   "ca",
	Short: "Scan Azure Container Apps",
	Long:  "Scan Azure Container Apps",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"ca"})
	},
}
