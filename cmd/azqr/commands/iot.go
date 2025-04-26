// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(iotCmd)
}

var iotCmd = &cobra.Command{
	Use:   "iot",
	Short: "Scan Azure IoT Hub",
	Long:  "Scan Azure IoT Hub",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"iot"})
	},
}
