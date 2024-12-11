// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(nwCmd)
}

var nwCmd = &cobra.Command{
	Use:   "nw",
	Short: "Scan Network Watcher",
	Long:  "Scan Network Watcher",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"nw"})
	},
}
