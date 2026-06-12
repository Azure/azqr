// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

// renderAzurePolicy creates and populates the Azure Policy sheet in the Excel report.
func renderAzurePolicy(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName: models.StageNamePolicy,
		sheetName: "Azure Policy",
		tableFunc: data.AzurePolicyTable,
	}, styles)
}
