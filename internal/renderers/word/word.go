// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package word

import (
	"fmt"
	"os"

	"github.com/fumiama/go-docx"
	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/renderers"
)

func CreateWordReport(data *renderers.ReportData) {
	filename := fmt.Sprintf("%s.docx", data.OutputFileName)
	log.Info().Msgf("Generating Report: %s", filename)

	records := data.RecommendationsTable()

	w := docx.New().WithDefaultTheme()

	para1 := w.AddParagraph()
	para1.AddText("Recommendations").Size("44")

	borderColors := &docx.APITableBorderColors{
		Top:     "#ff0000",
		Left:    "#ff0000",
		Bottom:  "#ff0000",
		Right:   "#ff0000",
		InsideH: "#ff0000",
		InsideV: "",
	}

	// add table
	cols := 11
	rows := len(records)
	table := w.AddTable(rows, cols, 1000, borderColors)
	for x, r := range table.TableRows {
		for y, c := range r.TableCells {
			c.AddParagraph().AddText(records[x][y+1])
		}
	}

	f, err := os.Create(filename)
	// save to file
	if err != nil {
		log.Fatal().Err(err).Msg("error creating word:")
	}
	_, err = w.WriteTo(f)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating word:")
	}
	err = f.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("error creating word:")
	}
}
