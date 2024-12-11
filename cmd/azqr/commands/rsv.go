// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(rsvCmd)
}

var rsvCmd = &cobra.Command{
	Use:   "rsv",
	Short: "Scan Recovery Service",
	Long:  "Scan Recovery Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"rsv"})
	},
}
