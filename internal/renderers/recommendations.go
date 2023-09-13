// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	_ "image/png"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderRecommendations(f *excelize.File, data ReportData) {
	if len(data.MainData) > 0 {
		_, err := f.NewSheet("Recommendations")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Recommendations sheet")
		}

		renderedRules := map[string]bool{}

		headers := []string{"Id", "Category", "Subcategory", "Description", "Severity", "Learn"}
		rows := [][]string{}
		for _, result := range data.MainData {
			for _, rr := range result.Rules {
				_, exists := renderedRules[rr.Id]
				if !exists && rr.IsBroken {
					rulesToRender := map[string]string{
						"Id":          rr.Id,
						"Category":    rr.Category,
						"Subcategory": rr.Subcategory,
						"Description": rr.Description,
						"Severity":    rr.Severity,
						"Learn":       rr.Learn,
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
