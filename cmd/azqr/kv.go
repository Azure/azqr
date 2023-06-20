// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/kv"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(kvCmd)
}

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "Scan Azure Key Vault",
	Long:  "Scan Azure Key Vault",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&kv.KeyVaultScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
