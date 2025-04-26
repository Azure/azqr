// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(rtCmd)
}

var rtCmd = &cobra.Command{
	Use:   "rt",
	Short: "Scan Route Table",
	Long:  "Scan Route Table",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"rt"})
	},
}
