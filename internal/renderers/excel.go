package renderers

import (
	"fmt"
	_ "image/png"
	"log"
	"unicode/utf8"

	"github.com/cmendible/azqr/internal/embeded"
	"github.com/xuri/excelize/v2"
)

func CreateExcelReport(data ReportData) {
	if len(data.MainData) > 0 {
		filename := fmt.Sprintf("%s.xlsx", data.OutputFileName)
		log.Printf("Generating Report: %s", filename)
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		RenderOverview(f, data)
		RenderRecommendations(f, data)
		RenderDefender(f, data)

		if err := f.SaveAs(filename); err != nil {
			log.Fatal(err)
		}
	}
}

func RenderOverview(f *excelize.File, data ReportData) {
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

	currentRow := 4
	cell, err := excelize.CoordinatesToCellName(1, currentRow)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetSheetRow("Overview", cell, &heathers)
	if err != nil {
		log.Fatal(err)
	}
	font := excelize.Font{Bold: true}
	style, err := f.NewStyle(&excelize.Style{Font: &font})
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetRowStyle("Overview", 4, 4, style)
	if err != nil {
		log.Fatal(err)
	}

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

	_ = Autofit(f, "Overview")

	cell, err = excelize.CoordinatesToCellName(len(heathers), currentRow)
	if err != nil {
		log.Fatal(err)
	}
	err = f.AutoFilter("Overview", fmt.Sprintf("A4:%s", cell), nil)
	if err != nil {
		log.Fatal(err)
	}

	logo := embeded.GetTemplates("microsoft.png")
	opt := &excelize.GraphicOptions{
		ScaleX:      1,
		ScaleY:      1,
		Positioning: "absolute",
	}
	if err := f.AddPictureFromBytes("Overview", "A1", "Azure Logo", ".png", logo, opt); err != nil {
		log.Fatal(err)
	}
}

func RenderRecommendations(f *excelize.File, data ReportData) {
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
					"Learn":       rr.Url,
				}
				renderedRules[rr.Id] = true
				rows = append(rows, mapToRow(heathers, rulesToRender)...)
			}
		}
	}

	display := "Learn"
	tooltip := "Learn more..."

	currentRow := 4
	cell, err := excelize.CoordinatesToCellName(1, currentRow)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetSheetRow("Recommendations", cell, &heathers)
	if err != nil {
		log.Fatal(err)
	}
	font := excelize.Font{Bold: true}
	style, err := f.NewStyle(&excelize.Style{Font: &font})
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetRowStyle("Recommendations", 4, 4, style)
	if err != nil {
		log.Fatal(err)
	}

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
		cell, _ = excelize.CoordinatesToCellName(6, currentRow)
		link, _ := f.GetCellValue("Recommendations", cell)
		if link != "" {
			_ = f.SetCellValue("Recommendations", cell, display)
			_ = f.SetCellHyperLink("Recommendations", cell, link, "External", excelize.HyperlinkOpts{Display: &display, Tooltip: &tooltip})
		}
	}

	_ = Autofit(f, "Recommendations")

	cell, err = excelize.CoordinatesToCellName(len(heathers), currentRow)
	if err != nil {
		log.Fatal(err)
	}
	err = f.AutoFilter("Recommendations", fmt.Sprintf("A4:%s", cell), nil)
	if err != nil {
		log.Fatal(err)
	}

	logo := embeded.GetTemplates("microsoft.png")
	opt := &excelize.GraphicOptions{
		ScaleX:      1,
		ScaleY:      1,
		Positioning: "absolute",
	}
	if err := f.AddPictureFromBytes("Recommendations", "A1", "Azure Logo", ".png", logo, opt); err != nil {
		log.Fatal(err)
	}

}

func RenderDefender(f *excelize.File, data ReportData) {
	if len(data.DefenderData) > 0 {
		_, err := f.NewSheet("Defender")
		if err != nil {
			log.Fatal(err)
		}

		heathers := data.DefenderData[0].GetProperties()

		rows := [][]string{}
		for _, r := range data.DefenderData {
			rows = append(mapToRow(heathers, r.ToMap(data.Mask)), rows...)
		}

		currentRow := 4
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			log.Fatal(err)
		}
		err = f.SetSheetRow("Defender", cell, &heathers)
		if err != nil {
			log.Fatal(err)
		}
		font := excelize.Font{Bold: true}
		style, err := f.NewStyle(&excelize.Style{Font: &font})
		if err != nil {
			log.Fatal(err)
		}
		err = f.SetRowStyle("Defender", 4, 4, style)
		if err != nil {
			log.Fatal(err)
		}

		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal(err)
			}
			err = f.SetSheetRow("Defender", cell, &row)
			if err != nil {
				log.Fatal(err)
			}
		}

		_ = Autofit(f, "Defender")

		cell, err = excelize.CoordinatesToCellName(len(heathers), currentRow)
		if err != nil {
			log.Fatal(err)
		}
		err = f.AutoFilter("Defender", fmt.Sprintf("A4:%s", cell), nil)
		if err != nil {
			log.Fatal(err)
		}

		logo := embeded.GetTemplates("microsoft.png")
		opt := &excelize.GraphicOptions{
			ScaleX:      1,
			ScaleY:      1,
			Positioning: "absolute",
		}
		if err := f.AddPictureFromBytes("Defender", "A1", "Azure Logo", ".png", logo, opt); err != nil {
			log.Fatal(err)
		}
	}
}

func Autofit(f *excelize.File, sheetName string) error {
	cols, err := f.GetCols(sheetName)
	if err != nil {
		return err
	}
	for idx, col := range cols {
		largestWidth := 0
		for _, rowCell := range col {
			cellWidth := utf8.RuneCountInString(rowCell) + 1
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}
		name, err := excelize.ColumnNumberToName(idx + 1)
		if err != nil {
			return err
		}
		err = f.SetColWidth(sheetName, name, name, float64(largestWidth))
		if err != nil {
			return err
		}
	}
	return nil
}

func mapToRow(heathers []string, m map[string]string) [][]string {
	v := make([]string, 0, len(m))

	for _, k := range heathers {
		v = append(v, m[k])
	}

	return [][]string{v}
}
