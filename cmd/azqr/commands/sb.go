// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(sbCmd)
}

var sbCmd = &cobra.Command{
	Use:   "sb",
	Short: "Scan Azure Service Bus",
	Long:  "Scan Azure Service Bus",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"sb"})
	},
}
