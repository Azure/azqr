// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	_ "image/png"
	"log"

	"github.com/xuri/excelize/v2"
)

func renderOverview(f *excelize.File, data ReportData) {
	err := f.SetSheetName("Sheet1", "Overview")
	if err != nil {
		log.Fatal(err)
	}

	heathers := data.MainData[0].GetHeathers()

	rows := [][]string{}
	for _, r := range data.MainData {
		rows = append(mapToRow(heathers, r.ToMap(data.Mask)), rows...)
	}

	createFirstRow(f, "Overview", heathers)

	currentRow := 4
	for _, row := range rows {
		currentRow += 1
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			log.Fatal(err)
		}
		err = f.SetSheetRow("Overview", cell, &row)
		if err != nil {
			log.Fatal(err)
		}
	}

	configureSheet(f, "Overview", heathers, currentRow)
}
