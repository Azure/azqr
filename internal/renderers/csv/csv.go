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
	records := data.ServicesTable()
	writeData(records, data.OutputFileName, "services")

	records = data.RecommendationsTable()
	writeData(records, data.OutputFileName, "recommendations")

	records = data.ImpactedTable()
	writeData(records, data.OutputFileName, "impacted")

	records = data.DefenderTable()
	writeData(records, data.OutputFileName, "defender")

	records = data.AdvisorTable()
	writeData(records, data.OutputFileName, "advisor")

	records = data.CostTable()
	writeData(records, data.OutputFileName, "costs")
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
