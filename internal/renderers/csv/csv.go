// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package csv

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/renderers"
)

func CreateCsvReport(data *renderers.ReportData) {
	// Only create AZQR-related CSV files if the feature is enabled
	if data.ScanEnabled {
		records := data.RecommendationsTable()
		writeData(records, data.OutputFileName, "recommendations")

		records = data.ImpactedTable()
		writeData(records, data.OutputFileName, "impacted")

		records = data.ResourceTypesTable()
		writeData(records, data.OutputFileName, "resourceType")

		records = data.ResourcesTable()
		writeData(records, data.OutputFileName, "inventory")

		records = data.ExcludedResourcesTable()
		writeData(records, data.OutputFileName, "outofscope")
	} else {
		log.Debug().Msg("Skipping AZQR CSV files. Feature is disabled")
	}

	// Only create Defender CSV files if the feature is enabled
	if data.DefenderEnabled {
		records := data.DefenderTable()
		writeData(records, data.OutputFileName, "defender")

		records = data.DefenderRecommendationsTable()
		writeData(records, data.OutputFileName, "defenderRecommendations")
	} else {
		log.Debug().Msg("Skipping Defender CSV files. Feature is disabled")
	}

	// Only create Azure Policy CSV files if the feature is enabled
	if data.PolicyEnabled {
		records := data.AzurePolicyTable()
		writeData(records, data.OutputFileName, "azurePolicy")
	} else {
		log.Debug().Msg("Skipping Azure Policy CSV file. Feature is disabled")
	}

	// Only create Arc SQL CSV files if the feature is enabled
	if data.ArcEnabled {
		records := data.ArcSQLTable()
		writeData(records, data.OutputFileName, "arcSQL")
	} else {
		log.Debug().Msg("Skipping Arc SQL CSV file. Feature is disabled")
	}

	// Only create Advisor CSV files if the feature is enabled
	if data.AdvisorEnabled {
		records := data.AdvisorTable()
		writeData(records, data.OutputFileName, "advisor")
	} else {
		log.Debug().Msg("Skipping Advisor CSV file. Feature is disabled")
	}

	// Only create Cost CSV files if the feature is enabled
	if data.CostEnabled {
		records := data.CostTable()
		writeData(records, data.OutputFileName, "costs")
	} else {
		log.Debug().Msg("Skipping Cost CSV file. Feature is disabled")
	}

	// Render external plugin results
	for _, result := range data.PluginResults {
		if len(result.Table) > 0 {
			// Use plugin name for the CSV filename extension
			writeData(result.Table, data.OutputFileName, fmt.Sprintf("plugin_%s", result.PluginName))
		}
	}
}

func writeData(data [][]string, fileName, extension string) {
	filename := fmt.Sprintf("%s.%s.csv", fileName, extension)
	log.Info().Msgf("Generating Report: %s", filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating csv:")
	}

	w := csv.NewWriter(f)
	err = w.WriteAll(data) // calls Flush internally

	if err != nil {
		log.Fatal().Err(w.Error()).Msg("error writing csv:")
	}
}
