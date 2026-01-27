// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// ResourceDiscoveryStage fetches all resources from Azure Resource Graph.
type ResourceDiscoveryStage struct {
	*BaseStage
}

func NewResourceDiscoveryStage() *ResourceDiscoveryStage {
	return &ResourceDiscoveryStage{
		BaseStage: NewBaseStage("Resource Discovery", true),
	}
}

func (s *ResourceDiscoveryStage) Execute(ctx *ScanContext) error {
	scanner := scanners.ResourceDiscovery{}

	resources, excludedResources := scanner.GetAllResources(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	ctx.ReportData.Resources = resources
	ctx.ReportData.ExludedResources = excludedResources

	// Check Excel row limit
	const excelMaxRows = 1048566
	if len(resources) > excelMaxRows {
		log.Fatal().
			Msgf("Number of resources (%d) exceeds Excel's maximum row limit (%d). Aborting scan.",
				len(resources), excelMaxRows)
	}

	log.Info().
		Int("resources", len(resources)).
		Int("excluded", len(excludedResources)).
		Msg("Discovered resources")

	return nil
}
