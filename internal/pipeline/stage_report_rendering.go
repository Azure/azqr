// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"fmt"

	"github.com/Azure/azqr/internal/renderers/csv"
	"github.com/Azure/azqr/internal/renderers/excel"
	"github.com/Azure/azqr/internal/renderers/json"
	"github.com/rs/zerolog/log"
)

// ReportRenderingStage generates the final output reports.
type ReportRenderingStage struct {
	*BaseStage
}

func NewReportRenderingStage() *ReportRenderingStage {
	return &ReportRenderingStage{
		BaseStage: NewBaseStage("Report Rendering", true),
	}
}

func (s *ReportRenderingStage) Execute(ctx *ScanContext) error {
	log.Info().Msg("Starting report rendering")

	// Log data summary before rendering
	log.Debug().
		Int("recommendation_types", len(ctx.ReportData.Recommendations)).
		Int("aprl_impacted_resources", len(ctx.ReportData.Graph)).
		Int("resources", len(ctx.ReportData.Resources)).
		Int("advisor_results", len(ctx.ReportData.Advisor)).
		Int("defender_results", len(ctx.ReportData.Defender)).
		Msg("Report data summary")

	// Generate Excel report
	if ctx.Params.Xlsx {
		log.Info().Msg("Generating Excel report")
		excel.CreateExcelReport(ctx.ReportData)
	}

	// Generate JSON report
	if ctx.Params.Json {
		log.Info().Msg("Generating JSON report")
		json.CreateJsonReport(ctx.ReportData)
	}

	// Generate CSV report
	if ctx.Params.Csv {
		log.Info().Msg("Generating CSV report")
		csv.CreateCsvReport(ctx.ReportData)
	}

	// Generate JSON output for stdout
	if ctx.Params.Stdout {
		outputJson := json.CreateJsonOutput(ctx.ReportData)
		fmt.Println(outputJson)
	}

	log.Info().
		Bool("xlsx", ctx.Params.Xlsx).
		Bool("json", ctx.Params.Json).
		Bool("csv", ctx.Params.Csv).
		Bool("stdout", ctx.Params.Stdout).
		Msg("Report rendering completed")

	return nil
}
