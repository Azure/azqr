// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(appiCmd)
}

var appiCmd = &cobra.Command{
	Use:   "appi",
	Short: "Scan Azure Application Insights",
	Long:  "Scan Azure Application Insights",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"appi"})
	},
}
