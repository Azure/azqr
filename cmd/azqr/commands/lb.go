// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(lbCmd)
}

var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Scan Azure Load Balancer",
	Long:  "Scan Azure Load Balancer",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"lb"})
	},
}
