// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(pipCmd)
}

var pipCmd = &cobra.Command{
	Use:   "pip",
	Short: "Scan Public IP",
	Long:  "Scan Public IP",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"pip"})
	},
}
