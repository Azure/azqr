// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(availCmd)
}

var availCmd = &cobra.Command{
	Use:   "avail",
	Short: "Scan Availability Sets",
	Long:  "Scan Availability Sets",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"avail"})
	},
}
