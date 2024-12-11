package renderers

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"

	"github.com/rs/zerolog/log"
)

func GetAllRecommendations(md bool) string {
	_, serviceScanners := models.GetScanners()
	aprlScanner := graph.NewAprlScanner(serviceScanners, nil, nil)
	aprl := aprlScanner.GetAprlRecommendations()

	var output string

	if md {
		output += "## Recommendations List\n\n"
		output += fmt.Sprintf("Total Supported Azure Resource Types: %d\n\n", len(aprl))
		output += "|  | Id | Resource Type | Category | Impact | Recommendation | Learn\n"
		output += "---|---|---|---|---|---|---\n"

		i := 0
		for _, scanner := range serviceScanners {
			rm := scanner.GetRecommendations()

			recommendations := map[string]models.AzqrRecommendation{}
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
				output += fmt.Sprintf("%s | %s | %s | %s | %s | %s | [Learn](%s)\n", fmt.Sprint(i), r.RecommendationID, r.ResourceType, r.Category, r.Impact, r.Recommendation, r.LearnMoreUrl)
			}

			for _, t := range scanner.ResourceTypes() {
				for _, r := range aprl[strings.ToLower(t)] {
					if strings.Contains(r.GraphQuery, "cannot-be-validated-with-arg") ||
						strings.Contains(r.GraphQuery, "under-development") ||
						strings.Contains(r.GraphQuery, "under development") {
						continue
					}

					i++
					output += fmt.Sprintf("%s | %s | %s | %s | %s | %s | [Learn](%s)\n", fmt.Sprint(i), r.RecommendationID, r.ResourceType, r.Category, r.Impact, r.Recommendation, r.LearnMoreLink[0].Url)
				}
			}
		}
	} else {
		j := []map[string]string{}
		i := 0
		for _, scanner := range serviceScanners {
			rm := scanner.GetRecommendations()

			recommendations := map[string]models.AzqrRecommendation{}
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
		output = string(js)
	}

	return output
}
