// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(netappCmd)
}

var netappCmd = &cobra.Command{
	Use:   "netapp",
	Short: "Scan NetApp",
	Long:  "Scan NetApp",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"netapp"})
	},
}
