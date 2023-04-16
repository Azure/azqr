// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"github.com/cmendible/azqr/internal/scanners"
)

type ReportData struct {
	OutputFileName     string
	EnableDetailedScan bool
	Mask               bool
	MainData           []scanners.AzureServiceResult
	DefenderData       []scanners.DefenderResult
	AdvisorData        []scanners.AdvisorResult
	CostData           *scanners.CostResult
}
