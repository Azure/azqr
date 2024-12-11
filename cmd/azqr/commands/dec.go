// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(decCmd)
}

var decCmd = &cobra.Command{
	Use:   "dec",
	Short: "Scan Azure Data Explorer",
	Long:  "Scan Azure Data Explorer",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"dec"})
	},
}
