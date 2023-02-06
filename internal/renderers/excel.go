package renderers

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func CreateExcelReport(data ReportData) {
	if len(data.MainData) > 0 {
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

		heathers := data.MainData[0].GetHeathers()

		rows := [][]string{}
		for _, r := range data.MainData {
			rows = append(mapToRow(heathers, r.ToMap(data.Mask)), rows...)
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

		if len(data.DefenderData) > 0 {
			_, err := f.NewSheet("Defender")
			if err != nil {
				log.Fatal(err)
				return
			}

			heathers := data.DefenderData[0].GetProperties()

			rows := [][]string{}
			for _, r := range data.DefenderData {
				rows = append(mapToRow(heathers, r.ToMap(data.Mask)), rows...)
			}

			for idx, row := range rows {
				cell, err := excelize.CoordinatesToCellName(1, idx+1)
				if err != nil {
					log.Fatal(err)
					return
				}
				if idx > 0 {
					err := f.SetSheetRow("Defender", cell, &row)
					if err != nil {
						log.Fatal(err)
						return
					}
				} else {
					err := f.SetSheetRow("Defender", cell, &heathers)
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
					err = f.SetRowStyle("Defender", 1, 1, style)
					if err != nil {
						log.Fatal(err)
						return
					}
				}
			}
		}

		if err := f.SaveAs(fmt.Sprintf("%s.xlsx", data.OutputFileName)); err != nil {
			log.Fatal(err)
		}
	}
}
