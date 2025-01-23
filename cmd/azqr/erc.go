// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(ercCmd)
}

var ercCmd = &cobra.Command{
	Use:   "erc",
	Short: "Scan Express Route Circuits",
	Long:  "Scan Express Route Circuits",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"erc"})
	},
}
