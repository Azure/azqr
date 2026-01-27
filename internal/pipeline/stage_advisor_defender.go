// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// AdvisorDefenderStage executes Advisor, Defender, Policy, and Arc scans.
type AdvisorDefenderStage struct {
	*BaseStage
}

func NewAdvisorDefenderStage() *AdvisorDefenderStage {
	return &AdvisorDefenderStage{
		BaseStage: NewBaseStage("Advisor & Defender Scan", true),
	}
}

func (s *AdvisorDefenderStage) Execute(ctx *ScanContext) error {
	advisorScanner := scanners.AdvisorScanner{}
	defenderScanner := scanners.DefenderScanner{}
	azurePolicyScanner := scanners.AzurePolicyScanner{}
	arcSQLScanner := scanners.ArcSQLScanner{}

	// Scan advisor
	if ctx.Params.Advisor {
		ctx.ReportData.Advisor = advisorScanner.Scan(
			ctx.Ctx,
			ctx.Params.Defender,
			ctx.Cred,
			ctx.Subscriptions,
			ctx.Params.Filters,
		)
	}

	// Scan defender
	if ctx.Params.Defender {
		ctx.ReportData.Defender = defenderScanner.Scan(
			ctx.Ctx,
			ctx.Params.Defender,
			ctx.Cred,
			ctx.Subscriptions,
			ctx.Params.Filters,
		)

		ctx.ReportData.DefenderRecommendations = defenderScanner.GetRecommendations(
			ctx.Ctx,
			ctx.Params.Defender,
			ctx.Cred,
			ctx.Subscriptions,
			ctx.Params.Filters,
		)
	}

	// Scan Azure Policy
	if ctx.Params.Policy {
		ctx.ReportData.AzurePolicy = azurePolicyScanner.Scan(
			ctx.Ctx,
			ctx.Cred,
			ctx.Subscriptions,
			ctx.Params.Filters,
		)
	}

	// Scan Arc-enabled SQL
	if ctx.Params.Arc {
		ctx.ReportData.ArcSQL = arcSQLScanner.Scan(
			ctx.Ctx,
			ctx.Cred,
			ctx.Subscriptions,
			ctx.Params.Filters,
		)
	}

	log.Info().Msg("Advisor & Defender scans completed")

	return nil
}
