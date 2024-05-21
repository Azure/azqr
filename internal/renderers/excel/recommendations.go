// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderRecommendations(f *excelize.File, data *renderers.ReportData) {
	if len(data.MainData) > 0 {
		err := f.SetSheetName("Sheet1", "Recommendations")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Recommendations sheet")
		}

		renderedRules := map[string]bool{}

		headers := []string{"Category", "Impact", "Recommendation", "Learn", "RId"}
		rows := [][]string{}
		for _, result := range data.MainData {
			for _, rr := range result.Rules {
				_, exists := renderedRules[rr.Id]
				if !exists && rr.NotCompliant {
					rulesToRender := map[string]string{
						"Category":       string(rr.Category),
						"Impact":         string(rr.Impact),
						"Recommendation": rr.Recommendation,
						"Learn":          rr.Learn,
						"RId":            rr.Id,
					}
					renderedRules[rr.Id] = true
					rows = append(rows, mapToRow(headers, rulesToRender)...)
				}
			}
		}

		createFirstRow(f, "Recommendations", headers)

		currentRow := 4
		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Recommendations", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}

			setHyperLink(f, "Recommendations", 6, currentRow)
		}

		configureSheet(f, "Recommendations", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Recommendations. No data to render")
	}
}
