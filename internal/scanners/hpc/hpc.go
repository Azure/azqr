// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hpc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["hpc"] = []models.IAzureScanner{&HighPerformanceComputingScanner{
		BaseScanner: models.NewBaseScanner("Specialized.Workload/HPC"),
	}}
}

// HighPerformanceComputingScanner - Scanner for High Performance Computing
type HighPerformanceComputingScanner struct {
	models.BaseScanner
}

// Init - Initializes the High Performance Computing Scanner
func (a *HighPerformanceComputingScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
