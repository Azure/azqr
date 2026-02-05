// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// DiagnosticsScanStage executes diagnostic settings scanning across all resources.
type DiagnosticsScanStage struct {
	*BaseStage
}

func NewDiagnosticsScanStage() *DiagnosticsScanStage {
	return &DiagnosticsScanStage{
		BaseStage: NewBaseStage("Diagnostics Settings Scan", false),
	}
}

func (s *DiagnosticsScanStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNameDiagnostics)
}

func (s *DiagnosticsScanStage) Execute(ctx *ScanContext) error {
	log.Debug().
		Int("resources_count", len(ctx.ReportData.Resources)).
		Msg("Diagnostics Stage ENTRY - starting execution")

	// Initialize diagnostic settings scanner
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	err := diagnosticsScanner.Init(ctx.Ctx, ctx.Cred, ctx.Params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize diagnostic settings scanner")
		return err
	}

	// Get diagnostic settings recommendations and add to report
	recommendations := scanners.GetRecommendations()
	recommendationCount := 0
	for resourceType, recs := range recommendations {
		for _, rec := range recs {
			// Add GraphRecommendation to report
			if ctx.ReportData.Recommendations[resourceType] == nil {
				ctx.ReportData.Recommendations[resourceType] = make(map[string]*models.GraphRecommendation)
			}

			ctx.ReportData.Recommendations[resourceType][rec.RecommendationID] = &rec
			recommendationCount++
		}
	}

	log.Debug().
		Int("diagnostic_recommendations_added", recommendationCount).
		Msg("Diagnostics recommendations collected")

	// Execute diagnostic settings scan to find resources without diagnostic settings
	diagResults := diagnosticsScanner.Scan(ctx.ReportData.Resources)
	ctx.ReportData.Graph = append(ctx.ReportData.Graph, diagResults...)

	log.Debug().
		Int("diagnostic_graph_results", len(diagResults)).
		Int("total_graph_results", len(ctx.ReportData.Graph)).
		Msg("Diagnostics scan completed")

	return nil
}
