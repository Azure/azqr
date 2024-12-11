// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(trafCmd)
}

var trafCmd = &cobra.Command{
	Use:   "traf",
	Short: "Scan Azure Traffic Manager",
	Long:  "Scan Azure Traffic Manager",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"traf"})
	},
}
