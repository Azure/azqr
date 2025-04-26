// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(evhCmd)
}

var evhCmd = &cobra.Command{
	Use:   "evh",
	Short: "Scan Azure Event Hubs",
	Long:  "Scan Azure Event Hubs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"evh"})
	},
}
