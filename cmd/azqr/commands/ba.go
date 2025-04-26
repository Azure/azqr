// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(baCmd)
}

var baCmd = &cobra.Command{
	Use:   "ba",
	Short: "Scan Azure Batch Account",
	Long:  "Scan Azure Batch Account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"ba"})
	},
}
