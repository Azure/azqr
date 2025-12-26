// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(fabricCmd)
}

var fabricCmd = &cobra.Command{
	Use:   "fabric",
	Short: "Scan Microsoft Fabric Capacities",
	Long:  "Scan Microsoft Fabric Capacities",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"fabric"})
	},
}
