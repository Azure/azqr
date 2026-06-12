// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

// renderArcSQL creates and populates the Arc SQL sheet in the Excel report.
func renderArcSQL(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName: models.StageNameArc,
		sheetName: "Arc SQL",
		tableFunc: data.ArcSQLTable,
	}, styles)
}
