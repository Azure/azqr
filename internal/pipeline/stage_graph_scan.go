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

// GraphScanStage executes ARG scanning.
type GraphScanStage struct {
	*BaseStage
}

func NewGraphScanStage() *GraphScanStage {
	return &GraphScanStage{
		BaseStage: NewBaseStage("Graph Scan", false),
	}
}

func (s *GraphScanStage) Skip(ctx *ScanContext) bool {
	// Graph stage is mandatory for regular scans
	return false
}

func (s *GraphScanStage) Execute(ctx *ScanContext) error {
	serviceScanners := ctx.Params.Filters.Azqr.Scanners
	log.Debug().
		Int("service_scanners_count", len(serviceScanners)).
		Int("subscriptions_count", len(ctx.Subscriptions)).
		Msg("Graph Stage ENTRY - starting execution")

	// Phase 1: Get initial recommendations with ALL scanners
	scanner := graph.NewScanner(serviceScanners, ctx.Params.Filters, ctx.Subscriptions)
	s.registerYamlPlugins(&scanner)

	recommendations, rules := scanner.ListRecommendations()
	ctx.ReportData.Recommendations = recommendations

	log.Debug().
		Int("recommendations_count", len(recommendations)).
		Int("rules_count", len(rules)).
		Msg("Graph Phase 1: Recommendations listed")

	// Get resource type counts for filtering
	resourceScanner := scanners.ResourceDiscovery{}
	resourceTypes := resourceScanner.GetCountPerResourceType(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	log.Debug().
		Int("resource_types_count", len(resourceTypes)).
		Msg("Graph Phase 1: Resource types retrieved")

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
		Msg("Graph Phase 1: Scanners filtered")

	if len(filteredScanners) == 0 {
		log.Warn().Msg("Graph: No filtered scanners - returning 0 results")
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

	// Phase 2: Create new ARG scanner with filtered scanners
	log.Debug().Msg("Graph Phase 2: Creating scanner with filtered scanners")
	scanner = graph.NewScanner(filteredScanners, ctx.Params.Filters, ctx.Subscriptions)
	s.registerYamlPlugins(&scanner)

	// Execute ARG scan
	log.Debug().Msg("Graph Phase 2: Executing scan")
	ctx.ReportData.Graph = scanner.Scan(ctx.Ctx, ctx.Cred)

	log.Debug().
		Int("graph_results", len(ctx.ReportData.Graph)).
		Msg("Graph scan completed")

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

func (s *GraphScanStage) registerYamlPlugins(aprlScanner *graph.GraphScanner) {
	yamlPluginRegistry := plugins.GetRegistry()
	for _, plugin := range yamlPluginRegistry.List() {
		if len(plugin.YamlRecommendations) > 0 {
			log.Info().
				Str("plugin", plugin.Metadata.Name).
				Int("queries", len(plugin.YamlRecommendations)).
				Msg("Registering YAML plugin queries with Graph scanner")
			for _, rec := range plugin.YamlRecommendations {
				aprlScanner.RegisterExternalQuery(rec.ResourceType, rec)
			}
		}
	}
}

func (s *GraphScanStage) filterServiceScanners(
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
				log.Debug().
					Str("resourceType", resourceType).
					Msgf("Skipping scanner")
				continue
			} else if add {
				filteredScanners = append(filteredScanners, scanner)
				add = false
				log.Info().
					Str("resourceType", resourceType).
					Msgf("Scanner will be used")
			}
		}
	}

	// Always include the generic resource scanner
	filteredScanners = append(filteredScanners, models.ScannerList["resource"][0])

	log.Debug().
		Int("original", len(serviceScanners)).
		Int("filtered", len(filteredScanners)).
		Msg("Filtered service scanners")

	return filteredScanners
}
