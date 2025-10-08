// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(arcCmd)
}

var arcCmd = &cobra.Command{
	Use:   "arc",
	Short: "Scan Azure Arc-enabled machines",
	Long:  "Scan Azure Arc-enabled machines",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"arc"})
	},
}
