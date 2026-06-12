// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

func renderResourceTypes(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName: models.StageNameGraph,
		sheetName: "ResourceTypes",
		tableFunc: data.ResourceTypesTable,
	}, styles)
}
