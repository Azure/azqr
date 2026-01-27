// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// AprlScanStage executes APRL (Azure Proactive Resiliency Library) scanning.
type AprlScanStage struct {
	*BaseStage
}

func NewAprlScanStage() *AprlScanStage {
	return &AprlScanStage{
		BaseStage: NewBaseStage("APRL Scan", false),
	}
}

func (s *AprlScanStage) CanSkip(ctx *ScanContext) bool {
	shouldSkip := !ctx.Params.UseAprlRecommendations
	log.Debug().
		Bool("use_aprl", ctx.Params.UseAprlRecommendations).
		Bool("will_skip", shouldSkip).
		Msg("APRL stage skip check")
	return shouldSkip
}

func (s *AprlScanStage) Execute(ctx *ScanContext) error {
	// Force INFO level for APRL stage debugging
	serviceScanners := ctx.Params.Filters.Azqr.Scanners
	log.Debug().
		Int("service_scanners_count", len(serviceScanners)).
		Int("subscriptions_count", len(ctx.Subscriptions)).
		Msg("APRL Stage ENTRY - starting execution")

	// Phase 1: Get initial recommendations with ALL scanners
	aprlScanner := graph.NewAprlScanner(serviceScanners, ctx.Params.Filters, ctx.Subscriptions)
	s.registerYamlPlugins(&aprlScanner)

	recommendations, rules := aprlScanner.ListRecommendations()
	ctx.ReportData.Recommendations = recommendations

	log.Debug().
		Int("recommendations_count", len(recommendations)).
		Int("rules_count", len(rules)).
		Msg("APRL Phase 1: Recommendations listed")

	// Get resource type counts for filtering
	resourceScanner := scanners.ResourceScanner{}
	resourceTypes := resourceScanner.GetCountPerResourceType(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	log.Debug().
		Int("resource_types_count", len(resourceTypes)).
		Msg("APRL Phase 1: Resource types retrieved")

	// Normalize resource types to lowercase
	normalizedResourceTypes := make(map[string]float64)
	for rt, count := range resourceTypes {
		normalizedResourceTypes[strings.ToLower(rt)] = count
	}

	// Filter scanners based on resource types
	filteredScanners := s.filterServiceScanners(serviceScanners, normalizedResourceTypes)

	log.Debug().
		Int("original_scanners", len(serviceScanners)).
		Int("filtered_scanners", len(filteredScanners)).
		Int("normalized_types", len(normalizedResourceTypes)).
		Msg("APRL Phase 1: Scanners filtered")

	if len(filteredScanners) == 0 {
		log.Warn().Msg("APRL: No filtered scanners - returning 0 results")
		// Still need to get resource type counts
		ctx.ReportData.ResourceTypeCount = resourceScanner.GetCountPerResourceTypeAndSubscription(
			ctx.Ctx,
			ctx.Cred,
			ctx.Subscriptions,
			ctx.ReportData.Recommendations,
			ctx.Params.Filters,
		)
		return nil
	}

	// Phase 2: Create new APRL scanner with filtered scanners
	log.Debug().Msg("APRL Phase 2: Creating scanner with filtered scanners")
	aprlScanner = graph.NewAprlScanner(filteredScanners, ctx.Params.Filters, ctx.Subscriptions)
	s.registerYamlPlugins(&aprlScanner)

	// Execute APRL scan
	log.Debug().Msg("APRL Phase 2: Executing scan")
	ctx.ReportData.Aprl = aprlScanner.Scan(ctx.Ctx, ctx.Cred)

	log.Debug().
		Int("aprl_results", len(ctx.ReportData.Aprl)).
		Msg("APRL scan completed")

	// Get resource type counts per subscription
	ctx.ReportData.ResourceTypeCount = resourceScanner.GetCountPerResourceTypeAndSubscription(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.ReportData.Recommendations,
		ctx.Params.Filters,
	)

	return nil
}

func (s *AprlScanStage) registerYamlPlugins(aprlScanner *graph.AprlScanner) {
	yamlPluginRegistry := plugins.GetRegistry()
	for _, plugin := range yamlPluginRegistry.List() {
		if len(plugin.YamlRecommendations) > 0 {
			log.Info().
				Str("plugin", plugin.Metadata.Name).
				Int("queries", len(plugin.YamlRecommendations)).
				Msg("Registering YAML plugin queries with APRL scanner")
			for _, rec := range plugin.YamlRecommendations {
				aprlScanner.RegisterExternalQuery(rec.ResourceType, rec)
			}
		}
	}
}

func (s *AprlScanStage) filterServiceScanners(
	serviceScanners []models.IAzureScanner,
	resourceTypes map[string]float64,
) []models.IAzureScanner {
	var filteredScanners []models.IAzureScanner

	for _, scanner := range serviceScanners {
		add := true
		for _, resourceType := range scanner.ResourceTypes() {
			resourceType = strings.ToLower(resourceType)

			// Check if the resource type exists across any subscription
			if count, exists := resourceTypes[resourceType]; !exists || count <= 0 {
				log.Debug().Msgf("Skipping scanner for resource type %s as it has no resources", resourceType)
				continue
			} else if add {
				filteredScanners = append(filteredScanners, scanner)
				add = false
				log.Info().Msgf("Scanner for resource type %s will be used", resourceType)
			}
		}
	}

	log.Debug().
		Int("original", len(serviceScanners)).
		Int("filtered", len(filteredScanners)).
		Msg("Filtered service scanners")

	return filteredScanners
}
