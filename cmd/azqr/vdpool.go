// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(vdPoolCmd)
}

var vdPoolCmd = &cobra.Command{
	Use:   "vdpool",
	Short: "Scan Azure Virtual Desktop",
	Long:  "Scan Azure Virtual Desktop",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"vdpool"})
	},
}
