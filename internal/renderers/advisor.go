package renderers

import (
	_ "image/png"
	"log"

	"github.com/xuri/excelize/v2"
)

func renderAdvisor(f *excelize.File, data ReportData) {
	if len(data.AdvisorData) > 0 {
		_, err := f.NewSheet("Advisor")
		if err != nil {
			log.Fatal(err)
		}

		heathers := data.AdvisorData[0].GetProperties()

		rows := [][]string{}
		for _, r := range data.AdvisorData {
			rows = append(mapToRow(heathers, r.ToMap(data.Mask)), rows...)
		}

		createFirstRow(f, "Advisor", heathers)

		currentRow := 4
		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal(err)
			}
			err = f.SetSheetRow("Advisor", cell, &row)
			if err != nil {
				log.Fatal(err)
			}
		}

		configureSheet(f, "Advisor", heathers, currentRow)
	}
}
