// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"fmt"
	"sort"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rulesCmd)
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Print all azqr rules",
	Long:  "Print all azqr rules as markdown table",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := internal.GetScanners()

		fmt.Println("#  | Category | Impact | Recommendation | More Info")
		fmt.Println("---|---|---|---|---")

		i := 0
		for _, scanner := range serviceScanners {
			rulesMap := scanner.GetRules()

			rules := map[string]scanners.AzureRule{}
			for _, r := range rulesMap {
				rules[r.Id] = r
			}

			keys := make([]string, 0, len(rules))
			for k := range rules {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				rule := rules[k]
				i++
				fmt.Printf("%s | %s | %s | %s | [Learn](%s)", fmt.Sprint(i), rule.Category, rule.Impact, rule.Recommendation, rule.Url)
				fmt.Println()
			}
		}
	},
}
