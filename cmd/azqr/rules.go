// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rulesCmd)
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Print all recommendations",
	Long:  "Print all recommendations as markdown table",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		_, serviceScanners := scanners.GetScanners()
		aprlScanner := internal.AprlScanner{}
		aprl := aprlScanner.GetAprlRecommendations()

		fmt.Println("#  | Id | Resource Type | Category | Impact | Recommendation | Learn")
		fmt.Println("---|---|---|---|---|---|---")

		i := 0
		for _, scanner := range serviceScanners {
			rm := scanner.GetRecommendations()

			recommendations := map[string]scanners.AzqrRecommendation{}
			for _, r := range rm {
				recommendations[r.RecommendationID] = r
			}

			keys := make([]string, 0, len(recommendations))
			for k := range recommendations {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				r := recommendations[k]
				i++
				fmt.Printf("%s | %s | %s | %s | %s | %s | [Learn](%s)", fmt.Sprint(i), r.RecommendationID, r.ResourceType, r.Category, r.Impact, r.Recommendation, r.LearnMoreUrl)
				fmt.Println()
			}

			for _, t := range scanner.ResourceTypes() {
				for _, r := range aprl[strings.ToLower(t)] {
					if strings.Contains(r.GraphQuery, "cannot-be-validated-with-arg") ||
						strings.Contains(r.GraphQuery, "under-development") ||
						strings.Contains(r.GraphQuery, "under development") {
						continue
					}

					i++
					fmt.Printf("%s | %s | %s | %s | %s | %s | [Learn](%s)", fmt.Sprint(i), r.RecommendationID, r.ResourceType, r.Category, r.Impact, r.Recommendation, r.LearnMoreLink[0].Url)
					fmt.Println()
				}
			}
		}
	},
}
