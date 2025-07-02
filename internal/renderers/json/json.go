package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/iancoleman/strcase"

	"github.com/rs/zerolog/log"
)

// CreateJsonReport generates a single consolidated JSON report
func CreateJsonReport(data *renderers.ReportData) {
	filename := fmt.Sprintf("%s.json", data.OutputFileName)
	log.Info().Msgf("Generating Report: %s", filename)

	// Build consolidated JSON structure
	consolidatedReport := map[string]interface{}{
		"recommendations":         convertToJSON(data.RecommendationsTable()),
		"impacted":                convertToJSON(data.ImpactedTable()),
		"resourceType":            convertToJSON(data.ResourceTypesTable()),
		"inventory":               convertToJSON(data.ResourcesTable()),
		"defender":                convertToJSON(data.DefenderTable()),
		"defenderRecommendations": convertToJSON(data.DefenderRecommendationsTable()),
		"advisor":                 convertToJSON(data.AdvisorTable()),
		"costs":                   convertToJSON(data.CostTable()),
		"outOfScope":              convertToJSON(data.ExcludedResourcesTable()),
	}

	// Write consolidated JSON to single file
	writeData(consolidatedReport, filename)
}

// writeData writes the consolidated JSON data to a single file
func writeData(data map[string]interface{}, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating json:")
	}

	defer func() {
		// Handle error during file close
		if cerr := f.Close(); cerr != nil {
			log.Fatal().Err(cerr).Msg("error closing file:")
		}
	}()

	js, err := json.MarshalIndent(data, "", "\t")
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
			item[strcase.ToLowerCamel(headers[i])] = value
		}
		result = append(result, item)
	}

	return result
}
