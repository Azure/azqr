package renderers

import (
	"github.com/cmendible/azqr/internal/scanners"
)

type ReportData struct {
	OutputFileName     string
	EnableDetailedScan bool
	MainData           []scanners.IAzureServiceResult
	DefenderData       []scanners.DefenderResult
}