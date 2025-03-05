package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
)

func CreateJsonReport(data *renderers.ReportData) {
	writeData(data.RecommendationsTable(), data.OutputFileName, "recommendations")
	writeData(data.ImpactedTable(), data.OutputFileName, "impacted")
	writeData(data.ResourceTypesTable(), data.OutputFileName, "resourceType")
	writeData(data.ResourcesTable(), data.OutputFileName, "inventory")
	writeData(data.DefenderTable(), data.OutputFileName, "defender")
  	writeData(data.DefenderRecommendationsTable(), data.OutputFileName, "defenderRecommendations")
	writeData(data.AdvisorTable(), data.OutputFileName, "advisor")
	writeData(data.CostTable(), data.OutputFileName, "costs")
  	writeData(data.ExcludedResourcesTable(), data.OutputFileName, "outofscope")
}

func writeData(data [][]string, fileName, extension string) {
	filename := fmt.Sprintf("%s.%s.json", fileName, extension)
	log.Info().Msgf("Generating Report: %s", filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating json:")
	}
	defer f.Close()

	jsonData := convertToJSON(data)

	js, err := json.MarshalIndent(jsonData, "", "\t")
	if err != nil {
		log.Fatal().Err(err).Msg("error marshaling data:")
	}

	_, err = f.Write(js)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json:")
	}
}

func convertToJSON(data [][]string) []map[string]string {
	var result []map[string]string
	headers := data[0]

	for _, row := range data[1:] {
		item := make(map[string]string)
		for i, value := range row {
			item[headers[i]] = value
		}
		result = append(result, item)
	}

	return result
}
