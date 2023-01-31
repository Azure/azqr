package renderers

import (
	"github.com/cmendible/azqr/internal/scanners"
)

type ReportData struct {
	Customer           string
	OutputFileName     string
	EnableDetailedScan bool
	MainData           []scanners.IAzureServiceResult
	DefenderData       []scanners.DefenderResult
}