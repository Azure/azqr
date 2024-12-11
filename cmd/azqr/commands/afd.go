// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

// init initializes the afd command and adds it to the scan command.
func init() {
	scanCmd.AddCommand(afdCmd)
}

// afdCmd represents the afd command.
var afdCmd = &cobra.Command{
	Use:   "afd",
	Short: "Scan Azure Front Door",
	Long:  "Scan Azure Front Door",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Call the scan function with the "afd" argument.
		scan(cmd, []string{"afd"})
	},
}
