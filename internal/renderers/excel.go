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

		renderOverview(f, data)
		renderRecommendations(f, data)
		renderDefender(f, data)
		renderServices(f, data)
		renderAdvisor(f, data)

		if err := f.SaveAs(filename); err != nil {
			log.Fatal(err)
		}
	}
}

func autofit(f *excelize.File, sheetName string) error {
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

func createFirstRow(f *excelize.File, sheet string, heathers []string) {
	currentRow := 4
	cell, err := excelize.CoordinatesToCellName(1, currentRow)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetSheetRow(sheet, cell, &heathers)
	if err != nil {
		log.Fatal(err)
	}
	font := excelize.Font{Bold: true}
	style, err := f.NewStyle(&excelize.Style{Font: &font})
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetRowStyle(sheet, 4, 4, style)
	if err != nil {
		log.Fatal(err)
	}
}

func setHyperLink(f *excelize.File, sheet string, col, currentRow int) {
	display := "Learn"
	tooltip := "Learn more..."
	cell, _ := excelize.CoordinatesToCellName(col, currentRow)
	link, _ := f.GetCellValue(sheet, cell)
	if link != "" {
		_ = f.SetCellValue(sheet, cell, display)
		_ = f.SetCellHyperLink(sheet, cell, link, "External", excelize.HyperlinkOpts{Display: &display, Tooltip: &tooltip})
	}
}

func configureSheet(f *excelize.File, sheet string, heathers []string, currentRow int) {
	_ = autofit(f, sheet)

	cell, err := excelize.CoordinatesToCellName(len(heathers), currentRow)
	if err != nil {
		log.Fatal(err)
	}
	err = f.AutoFilter(sheet, fmt.Sprintf("A4:%s", cell), nil)
	if err != nil {
		log.Fatal(err)
	}

	logo := embeded.GetTemplates("microsoft.png")
	opt := &excelize.GraphicOptions{
		ScaleX:      1,
		ScaleY:      1,
		Positioning: "absolute",
	}
	if err := f.AddPictureFromBytes(sheet, "A1", "Azure Logo", ".png", logo, opt); err != nil {
		log.Fatal(err)
	}
}
