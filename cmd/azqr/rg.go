// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(rgCmd)
}

var rgCmd = &cobra.Command{
	Use:   "rg",
	Short: "Scan Resource Groups",
	Long:  "Scan Resource Groups",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"rg"})
	},
}
