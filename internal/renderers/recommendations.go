// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	_ "image/png"
	"log"

	"github.com/xuri/excelize/v2"
)

func renderRecommendations(f *excelize.File, data ReportData) {
	_, err := f.NewSheet("Recommendations")
	if err != nil {
		log.Fatal(err)
	}

	renderedRules := map[string]bool{}

	heathers := []string{"Id", "Category", "Subcategory", "Description", "Severity", "Learn"}
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
				rows = append(rows, mapToRow(heathers, rulesToRender)...)
			}
		}
	}

	createFirstRow(f, "Recommendations", heathers)

	currentRow := 4
	for _, row := range rows {
		currentRow += 1
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			log.Fatal(err)
		}
		err = f.SetSheetRow("Recommendations", cell, &row)
		if err != nil {
			log.Fatal(err)
		}

		setHyperLink(f, "Recommendations", 6, currentRow)
	}

	configureSheet(f, "Recommendations", heathers, currentRow)
}
