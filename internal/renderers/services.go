// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"
	_ "image/png"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderServices(f *excelize.File, data ReportData) {
	if len(data.MainData) > 0 {
		_, err := f.NewSheet("Services")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Services sheet")
		}

		headers := []string{"Subscription", "Resource Group", "Location", "Type", "Service Name", "Broken", "Category", "Subcategory", "Severity", "Description", "Result", "Learn"}

		rbroken := [][]string{}
		rok := [][]string{}
		for _, d := range data.MainData {
			for _, r := range d.Rules {
				row := []string{
					scanners.MaskSubscriptionID(d.SubscriptionID, data.Mask),
					d.ResourceGroup,
					scanners.ParseLocation(d.Location),
					d.Type,
					d.ServiceName,
					fmt.Sprintf("%t", r.IsBroken),
					r.Category,
					r.Subcategory,
					r.Severity,
					r.Description,
					r.Result,
					r.Learn,
				}
				if r.IsBroken {
					rbroken = append([][]string{row}, rbroken...)
				} else {
					rok = append([][]string{row}, rok...)
				}
			}
		}

		createFirstRow(f, "Services", headers)

		rows := append(rbroken, rok...)

		currentRow := 4
		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Services", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
			setHyperLink(f, "Services", 12, currentRow)
		}

		configureSheet(f, "Services", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Services. No data to render")
	}
}
