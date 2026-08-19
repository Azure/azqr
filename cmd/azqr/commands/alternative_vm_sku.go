// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/skus"
	"github.com/spf13/cobra"
)

func init() {
	alternativeVmSkuCmd.Flags().StringP("sku", "s", "", "Azure VM SKU name to find alternatives for (required)")
	alternativeVmSkuCmd.Flags().IntP("top", "n", 10, "Number of alternative SKUs to return")
	_ = alternativeVmSkuCmd.MarkFlagRequired("sku")
	rootCmd.AddCommand(alternativeVmSkuCmd)
}

var alternativeVmSkuCmd = &cobra.Command{
	Use:   "alternative-vm-sku",
	Short: "Find alternative VM SKUs similar to the given SKU",
	Long: `Find alternative Azure VM SKUs ranked by compatibility score.

The compatibility score (0.0–1.0) is computed from:
  - vCPU count match           (weight 0.30)
  - Memory match               (weight 0.25)
  - Same model series          (weight 0.15)
  - Version proximity          (weight 0.05)
  - GPU count match            (weight 0.15, GPU SKUs only)
  - GPU workload class match   (weight 0.20, NV/NC/ND — GPU SKUs only)
  - Cross-family bonus         (+0.10 for different family, scaled down 0.02/generation behind)
  - Same-family penalty        (-0.10; same family faces the same regional capacity constraints)
  - Data disk match            (weight 0.05)
  - Accelerated networking     (weight 0.05)

Returns the top N results (default 10) as JSON.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		skuName, _ := cmd.Flags().GetString("sku")
		top, _ := cmd.Flags().GetInt("top")

		target, ok := skus.Lookup(skuName)
		if !ok {
			return fmt.Errorf("SKU %q not found in the known SKU list", skuName)
		}

		recommendations := skus.FindAlternatives(target, top)

		type output struct {
			Source       skus.SKU              `json:"source"`
			Alternatives []skus.Recommendation `json:"alternatives"`
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output{Source: target, Alternatives: recommendations})
	},
}
