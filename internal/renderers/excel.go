package renderers

import (
	"fmt"
	"log"

	"github.com/cmendible/azqr/internal/scanners"
	"github.com/xuri/excelize/v2"
)

func CreateExcelReport(all []scanners.IAzureServiceResult, outputFile string) {
	if len(all) > 0 {
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		err := f.SetSheetName("Sheet1", "Overview")
		if err != nil {
			log.Fatal(err)
			return
		}

		heathers := all[0].GetProperties()

		rows := [][]string{}
		for _, r := range all {
			rows = append(mapToRow(heathers, r.ToMap()), rows...)
		}

		for idx, row := range rows {
			cell, err := excelize.CoordinatesToCellName(1, idx+1)
			if err != nil {
				log.Fatal(err)
				return
			}
			if idx > 0 {
				err := f.SetSheetRow("Overview", cell, &row)
				if err != nil {
					log.Fatal(err)
					return
				}
			} else {
				err := f.SetSheetRow("Overview", cell, &heathers)
				if err != nil {
					log.Fatal(err)
					return
				}
				font := excelize.Font{Bold: true}
				style, err := f.NewStyle(&excelize.Style{Font: &font})
				if err != nil {
					fmt.Println(err)
					return
				}
				err = f.SetRowStyle("Overview", 1, 1, style)
				if err != nil {
					log.Fatal(err)
					return
				}
			}
		}

		if err := f.SaveAs(fmt.Sprintf("%s.xlsx", outputFile)); err != nil {
			log.Fatal(err)
		}
	}
}
