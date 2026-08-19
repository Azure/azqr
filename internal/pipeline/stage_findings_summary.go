// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import "github.com/Azure/azqr/internal/findings"

// FindingsSummaryStage builds the canonical recommendation summary.
type FindingsSummaryStage struct {
	*BaseStage
}

// NewFindingsSummaryStage creates a findings summary stage.
func NewFindingsSummaryStage() *FindingsSummaryStage {
	return &FindingsSummaryStage{BaseStage: NewBaseStage("Findings Summary", true)}
}

// Execute builds the summary after all recommendation-producing stages finish.
func (s *FindingsSummaryStage) Execute(ctx *ScanContext) error {
	ctx.ReportData.Summary = findings.Build(
		ctx.ReportData.Recommendations,
		ctx.ReportData.Graph,
		len(ctx.ReportData.Resources),
	)
	return nil
}
