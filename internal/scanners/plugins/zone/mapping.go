// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package zone

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// ZoneMappingScanner is an internal plugin that retrieves availability zone mappings
type ZoneMappingScanner struct{}

// NewScanner creates a new zone mapping scanner
func NewScanner() *ZoneMappingScanner {
	return &ZoneMappingScanner{}
}

// GetMetadata returns plugin metadata
func (s *ZoneMappingScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "zone-mapping",
		Version:     "1.0.0",
		Description: "Retrieves logical-to-physical availability zone mappings for all Azure regions in each subscription",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription", DataKey: "subscription", FilterType: plugins.FilterTypeSearch},
			{Name: "Location", DataKey: "location", FilterType: plugins.FilterTypeDropdown},
			{Name: "Display Name", DataKey: "displayName", FilterType: plugins.FilterTypeDropdown},
			{Name: "Logical Zone", DataKey: "logicalZone", FilterType: plugins.FilterTypeDropdown},
			{Name: "Physical Zone", DataKey: "physicalZone", FilterType: plugins.FilterTypeSearch},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *ZoneMappingScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Scanning availability zone mappings across subscriptions")

	// Build header row from ColumnMetadata (single source of truth).
	table := [][]string{s.GetMetadata().HeaderRow()}

	// Create a single HTTP client to share across all goroutines.
	// az.HttpClient is safe for concurrent use.
	httpClient := az.NewHttpClient(cred, az.DefaultHttpClientOptions(30*time.Second))

	// Use parallel processing for performance
	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make([]zoneMappingResult, 0, len(subscriptions)*10)

	// Process subscriptions in parallel with worker pool
	const maxWorkers = 5
	semaphore := make(chan struct{}, maxWorkers)

	for subID, subName := range subscriptions {
		wg.Add(1)
		go func(subscriptionID, subscriptionName string) {
			defer wg.Done()

			// Check for context cancellation before acquiring the semaphore
			if ctx.Err() != nil {
				return
			}

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Debug().Msgf("Fetching zone mappings for subscription: %s", subscriptionName)

			subResults, err := s.fetchZoneMappings(ctx, httpClient, subscriptionID, subscriptionName)
			if err != nil {
				log.Error().
					Err(err).
					Str("subscription", subscriptionName).
					Msg("Failed to fetch zone mappings")
				return
			}

			// Thread-safe append
			mu.Lock()
			results = append(results, subResults...)
			mu.Unlock()

			log.Debug().Msgf("Fetched %d zone mappings for subscription: %s", len(subResults), subscriptionName)
		}(subID, subName)
	}

	wg.Wait()

	log.Info().Msgf("Total zone mappings retrieved: %d", len(results))

	// Sort results for consistent output
	sort.Slice(results, func(i, j int) bool {
		if results[i].subscriptionName != results[j].subscriptionName {
			return results[i].subscriptionName < results[j].subscriptionName
		}
		if results[i].location != results[j].location {
			return results[i].location < results[j].location
		}
		return results[i].logicalZone < results[j].logicalZone
	})

	// Convert results to table format
	for _, result := range results {
		table = append(table, []string{
			result.subscriptionName,
			result.location,
			result.displayName,
			result.logicalZone,
			result.physicalZone,
		})
	}

	return []plugins.ExternalPluginOutput{{
		Metadata:    s.GetMetadata(),
		SheetName:   "Zone Mapping",
		Description: "Logical-to-physical availability zone mappings for Azure regions by subscription",
		Table:       table,
	}}, nil
}

// fetchZoneMappings retrieves zone mappings for a single subscription using az.HttpClient.
// JSON parsing is delegated to parseZoneMappings (also used by tests).
func (s *ZoneMappingScanner) fetchZoneMappings(ctx context.Context, httpClient *az.HttpClient, subscriptionID, subscriptionName string) ([]zoneMappingResult, error) {
	endpoint := fmt.Sprintf("https://management.azure.com/subscriptions/%s/locations?api-version=2022-12-01", subscriptionID)

	body, err := httpClient.Do(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}

	results, err := parseZoneMappings(body, subscriptionID, subscriptionName)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Subscription %s: found %d zone mappings", subscriptionName, len(results))
	return results, nil
}

// parseZoneMappings is a thin testable wrapper around az.ParseLocations that converts
// the shared LocationInfo slice to zone-plugin-specific zoneMappingResult records.
func parseZoneMappings(body []byte, subscriptionID, subscriptionName string) ([]zoneMappingResult, error) {
	locations, err := az.ParseLocations(body)
	if err != nil {
		return nil, err
	}
	results := make([]zoneMappingResult, 0)
	for _, location := range locations {
		if len(location.AvailabilityZoneMappings) == 0 {
			continue
		}
		for _, mapping := range location.AvailabilityZoneMappings {
			results = append(results, zoneMappingResult{
				subscriptionID:   subscriptionID,
				subscriptionName: subscriptionName,
				location:         location.Name,
				displayName:      location.DisplayName,
				logicalZone:      mapping.LogicalZone,
				physicalZone:     mapping.PhysicalZone,
			})
		}
	}
	return results, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("zone-mapping", NewScanner())
}
