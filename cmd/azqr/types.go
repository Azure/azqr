// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"fmt"
	"sort"

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
		fmt.Println("Abbreviation  | Resource Type ")
		fmt.Println("---|---")
		keys := make([]string, 0, len(scanners.ScannerList))
		for key := range scanners.ScannerList {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, t := range scanners.ScannerList[key] {
				for _, rt := range t.ResourceTypes() {
					fmt.Printf("%s | %s", key, rt)
					fmt.Println()
				}
			}
		}
	},
}
