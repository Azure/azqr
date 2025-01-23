// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(itCmd)
}

var itCmd = &cobra.Command{
	Use:   "it",
	Short: "Scan Image Template",
	Long:  "Scan Image Template",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"it"})
	},
}
