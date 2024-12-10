// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/azqr"
	iot "github.com/Azure/azqr/internal/scanners/disk"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(diskCmd)
}

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Scan Disk",
	Long:  "Scan Disk",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []azqr.IAzureScanner{
			&iot.DiskScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
