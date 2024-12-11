// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(sigrCmd)
}

var sigrCmd = &cobra.Command{
	Use:   "sigr",
	Short: "Scan Azure SignalR",
	Long:  "Scan Azure SignalR",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"sigr"})
	},
}
