// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(apimCmd)
}

var apimCmd = &cobra.Command{
	Use:   "apim",
	Short: "Scan Azure API Management",
	Long:  "Scan Azure API Management",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"apim"})
	},
}
