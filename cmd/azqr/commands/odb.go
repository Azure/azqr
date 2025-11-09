// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(odbCmd)
}

var odbCmd = &cobra.Command{
	Use:   "odb",
	Short: "Scan Oracle Database@Azure",
	Long:  "Scan Oracle Database@Azure",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"odb"})
	},
}
