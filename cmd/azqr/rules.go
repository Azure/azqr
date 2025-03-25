// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/spf13/cobra"

	"github.com/rs/zerolog/log"
)

func init() {
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output rules list in JSON format")
	rootCmd.AddCommand(rulesCmd)
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Print all recommendations",
	Long:  "Print all recommendations as markdown table",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		oj, _ := cmd.Flags().GetBool("json")
		_, serviceScanners := scanners.GetScanners()
		aprlScanner := internal.NewAprlScanner(serviceScanners, nil, nil)
		aprl := aprlScanner.GetAprlRecommendations()

		if !oj {
			fmt.Println("## Recommendations List")
			fmt.Println("")
			fmt.Println("Total recommendations:", len(aprl))
			fmt.Println("")
			fmt.Println("|  | Id | Resource Type | Category | Impact | Recommendation | Learn")
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
		} else {
			j := []map[string]string{}
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
					j = append(j, map[string]string{})
					j[i] = map[string]string{}
					j[i]["recommendationId"] = recommendations[k].RecommendationID
					j[i]["resourceType"] = recommendations[k].ResourceType
					j[i]["category"] = string(recommendations[k].Category)
					j[i]["impact"] = string(recommendations[k].Impact)
					j[i]["recommendation"] = recommendations[k].Recommendation
					j[i]["learnMoreUrl"] = recommendations[k].LearnMoreUrl
					i++
				}

				for _, t := range scanner.ResourceTypes() {
					for _, r := range aprl[strings.ToLower(t)] {
						if strings.Contains(r.GraphQuery, "cannot-be-validated-with-arg") ||
							strings.Contains(r.GraphQuery, "under-development") ||
							strings.Contains(r.GraphQuery, "under development") {
							continue
						}

						j = append(j, map[string]string{})
						j[i] = map[string]string{}
						j[i]["recommendationId"] = r.RecommendationID
						j[i]["resourceType"] = r.ResourceType
						j[i]["category"] = string(r.Category)
						j[i]["impact"] = string(r.Impact)
						j[i]["recommendation"] = r.Recommendation
						j[i]["learnMoreUrl"] = r.LearnMoreLink[0].Url
						i++
					}
				}
			}
			// print j as json to stdout
			js, err := json.MarshalIndent(j, "", "\t")
			if err != nil {
				log.Fatal().Err(err).Msg("error marshaling data:")
			}
			fmt.Println(string(js))
		}
	},
}
