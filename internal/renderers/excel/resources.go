// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

func renderResources(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName:    models.StageNameGraph,
		sheetName:    "Inventory",
		tableFunc:    data.ResourcesTable,
		hyperlinkCol: 12,
	}, styles)
}

func renderExcludedResources(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName:    models.StageNameGraph,
		sheetName:    "OutOfScope",
		tableFunc:    data.ExcludedResourcesTable,
		hyperlinkCol: 12,
	}, styles)
}
