// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(diskCmd)
}

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Scan Disk",
	Long:  "Scan Disk",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"disk"})
	},
}
