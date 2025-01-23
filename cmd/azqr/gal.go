// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(galCmd)
}

var galCmd = &cobra.Command{
	Use:   "gal",
	Short: "Scan Azure Galleries",
	Long:  "Scan Azure Galleries",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"gal"})
	},
}
