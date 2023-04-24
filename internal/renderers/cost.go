// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"
	_ "image/png"
	"log"

	"github.com/xuri/excelize/v2"
)

func renderCosts(f *excelize.File, data ReportData) {
	if len(data.CostData.Items) > 0 {
		_, err := f.NewSheet("Costs")
		if err != nil {
			log.Fatal(err)
		}

		heathers := data.CostData.GetProperties()

		rows := [][]string{}
		for _, r := range data.CostData.Items {
			rows = append(mapToRow(heathers, r.ToMap(data.Mask)), rows...)
		}

		createFirstRow(f, "Costs", heathers)

		cell, err := excelize.CoordinatesToCellName(2, 1)
		if err != nil {
			log.Fatal(err)
		}

		err = f.SetCellDefault(
			"Costs",
			cell,
			fmt.Sprintf("Costs from %s to %s", data.CostData.From.Format("2006-01-02"), data.CostData.To.Format("2006-01-02")))
		if err != nil {
			log.Fatal(err)
		}

		currentRow := 4
		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal(err)
			}
			err = f.SetSheetRow("Costs", cell, &row)
			if err != nil {
				log.Fatal(err)
			}
		}

		configureSheet(f, "Costs", heathers, currentRow)
	} else {
		log.Println("Skipping Costs. No data to render")
	}
}
