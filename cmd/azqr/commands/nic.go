// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(nicCmd)
}

var nicCmd = &cobra.Command{
	Use:   "nic",
	Short: "Scan NICs",
	Long:  "Scan NICs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"nic"})
	},
}
