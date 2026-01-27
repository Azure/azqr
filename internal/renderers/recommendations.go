package renderers

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"

	"github.com/rs/zerolog/log"
)

func GetAllRecommendations(md bool) string {
	_, serviceScanners := models.GetScanners()
	graphScanner := graph.NewScanner(serviceScanners, nil, nil)
	graphRec := graphScanner.GetRecommendations()
	diagSettingsRec := scanners.GetRecommendations()

	var output string

	graphRecommendations := map[string]models.GraphRecommendation{}
	for _, scanner := range serviceScanners {
		for _, t := range scanner.ResourceTypes() {
			for _, r := range graphRec[strings.ToLower(t)] {
				if strings.Contains(r.GraphQuery, "cannot-be-validated-with-arg") ||
					strings.Contains(r.GraphQuery, "under-development") ||
					strings.Contains(r.GraphQuery, "under development") ||
					strings.EqualFold(r.MetadataState, "disabled") {
					continue
				}
				graphRecommendations[r.RecommendationID] = r
			}
			for _, r := range diagSettingsRec[strings.ToLower(t)] {
				graphRecommendations[r.RecommendationID] = r
			}
		}
	}

	graphKeys := make([]string, 0, len(graphRecommendations))
	for k := range graphRecommendations {
		graphKeys = append(graphKeys, k)
	}
	sort.Strings(graphKeys)

	if md {
		output += "## Recommendations List\n\n"
		output += fmt.Sprintf("Total Supported Azure Resource Types: %d\n\n", len(graphRec))
		output += "|  | Id | Resource Type | Category | Impact | Recommendation | Learn\n"
		output += "---|---|---|---|---|---|---\n"

		i := 0

		for _, k := range graphKeys {
			r := graphRecommendations[k]
			i++
			output += fmt.Sprintf("%s | %s | %s | %s | %s | %s | [Learn](%s)\n", fmt.Sprint(i), r.RecommendationID, r.ResourceType, r.Category, r.Impact, r.Recommendation, r.LearnMoreLink[0].Url)
		}
	} else {
		j := []map[string]string{}
		i := 0

		for _, k := range graphKeys {
			r := graphRecommendations[k]
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

		// print j as json to stdout
		js, err := json.MarshalIndent(j, "", "\t")
		if err != nil {
			log.Fatal().Err(err).Msg("error marshaling data:")
		}
		output = string(js)
	}

	return output
}
