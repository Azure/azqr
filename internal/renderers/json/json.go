package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/iancoleman/strcase"

	"github.com/rs/zerolog/log"
)

// buildConsolidatedReport builds the consolidated JSON structure from report data
func buildConsolidatedReport(data *renderers.ReportData) map[string]interface{} {
	consolidatedReport := map[string]interface{}{}

	// Only include AZQR-related data if the feature is enabled
	if data.ScanEnabled {
		consolidatedReport["recommendations"] = convertToJSON(data.RecommendationsTable())
		consolidatedReport["impacted"] = convertToJSON(data.ImpactedTable())
		consolidatedReport["resourceType"] = convertToJSON(data.ResourceTypesTable())
		consolidatedReport["inventory"] = convertToJSON(data.ResourcesTable())
		consolidatedReport["outOfScope"] = convertToJSON(data.ExcludedResourcesTable())
	} else {
		log.Debug().Msg("Skipping AZQR data in JSON. Feature is disabled")
	}

	// Only include Advisor data if the feature is enabled
	if data.AdvisorEnabled {
		consolidatedReport["advisor"] = convertToJSON(data.AdvisorTable())
	} else {
		log.Debug().Msg("Skipping Advisor data in JSON. Feature is disabled")
	}

	// Only include Azure Policy data if the feature is enabled
	if data.PolicyEnabled {
		consolidatedReport["azurePolicy"] = convertToJSON(data.AzurePolicyTable())
	} else {
		log.Debug().Msg("Skipping Azure Policy data in JSON. Feature is disabled")
	}

	// Only include Arc SQL data if the feature is enabled
	if data.ArcEnabled {
		consolidatedReport["arcSQL"] = convertToJSON(data.ArcSQLTable())
	} else {
		log.Debug().Msg("Skipping Arc SQL data in JSON. Feature is disabled")
	}

	// Only include Defender data if the feature is enabled
	if data.DefenderEnabled {
		consolidatedReport["defender"] = convertToJSON(data.DefenderTable())
		consolidatedReport["defenderRecommendations"] = convertToJSON(data.DefenderRecommendationsTable())
	} else {
		log.Debug().Msg("Skipping Defender data in JSON. Feature is disabled")
	}

	// Only include Cost data if the feature is enabled
	if data.CostEnabled {
		consolidatedReport["costs"] = convertToJSON(data.CostTable())
	} else {
		log.Debug().Msg("Skipping Cost data in JSON. Feature is disabled")
	}

	// Add external plugin results
	if len(data.PluginResults) > 0 {
		plugins := make(map[string]interface{})
		for _, result := range data.PluginResults {
			pluginData := map[string]interface{}{
				"description": result.Description,
				"sheetName":   result.SheetName,
				"data":        convertToJSON(result.Table),
			}
			plugins[result.PluginName] = pluginData
		}
		consolidatedReport["externalPlugins"] = plugins
	}

	return consolidatedReport
}

// CreateJsonReport generates a single consolidated JSON report file
func CreateJsonReport(data *renderers.ReportData) {
	filename := fmt.Sprintf("%s.json", data.OutputFileName)
	log.Info().Msgf("Generating Report: %s", filename)

	consolidatedReport := buildConsolidatedReport(data)
	writeData(consolidatedReport, filename)
}

// CreateJsonOutput generates the same consolidated JSON structure as CreateJsonReport
// but returns it as a string for console output instead of writing to a file
func CreateJsonOutput(data *renderers.ReportData) string {
	consolidatedReport := buildConsolidatedReport(data)

	js, err := json.MarshalIndent(consolidatedReport, "", "\t")
	if err != nil {
		log.Fatal().Err(err).Msg("error marshaling data:")
	}

	return string(js)
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
