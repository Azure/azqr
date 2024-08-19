// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"fmt"
	"slices"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(typesCmd)
}

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "Print all supported azure resource types",
	Long:  "Print all supported azure resource types",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := scanners.GetScanners()

		strs := []string{}

		for _, scanner := range serviceScanners {
			strs = append(strs, scanner.ResourceTypes()...)
		}
		slices.Sort(strs)
		
		for _, t := range strs {
			fmt.Printf("* %s", t)
			fmt.Println()
		}

	},
}
