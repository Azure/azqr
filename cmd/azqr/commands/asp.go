// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "asp",
	Short: "Scan Azure App Service",
	Long:  "Scan Azure App Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"asp"})
	},
}
