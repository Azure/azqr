// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(appcsCmd)
}

var appcsCmd = &cobra.Command{
	Use:   "appcs",
	Short: "Scan Azure App Configuration",
	Long:  "Scan Azure App Configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"appcs"})
	},
}
