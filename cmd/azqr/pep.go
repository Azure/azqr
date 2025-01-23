// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(pepCmd)
}

var pepCmd = &cobra.Command{
	Use:   "pep",
	Short: "Scan Private Endpoint",
	Long:  "Scan Private Endpoint",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"pep"})
	},
}
