// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package zone

import (
	"context"
	"encoding/json"
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

// locationResponse represents the API response structure
type locationResponse struct {
	Value []locationInfo `json:"value"`
}

// locationInfo represents a single location with zone mappings
type locationInfo struct {
	Name                     *string                   `json:"name"`
	DisplayName              *string                   `json:"displayName"`
	AvailabilityZoneMappings []availabilityZoneMapping `json:"availabilityZoneMappings"`
}

// availabilityZoneMapping represents a logical-to-physical zone mapping
type availabilityZoneMapping struct {
	LogicalZone  *string `json:"logicalZone"`
	PhysicalZone *string `json:"physicalZone"`
}

// fetchZoneMappings retrieves zone mappings for a single subscription using az.HttpClient
func (s *ZoneMappingScanner) fetchZoneMappings(ctx context.Context, httpClient *az.HttpClient, subscriptionID, subscriptionName string) ([]zoneMappingResult, error) {
	// Construct the REST API URL
	// GET /subscriptions/{subscriptionId}/locations?api-version=2022-12-01
	endpoint := fmt.Sprintf("https://management.azure.com/subscriptions/%s/locations?api-version=2022-12-01", subscriptionID)

	// Make the request with authentication
	body, err := httpClient.Do(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}

	return parseZoneMappings(body, subscriptionID, subscriptionName)
}

// derefStr safely dereferences a *string, returning "" for nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// parseZoneMappings parses a locations REST response body and converts it into
// zoneMappingResult records. Locations without availability zone mappings are
// skipped, and nil (optional) fields are normalized to empty strings.
func parseZoneMappings(body []byte, subscriptionID, subscriptionName string) ([]zoneMappingResult, error) {
	var locationsResp locationResponse
	if err := json.Unmarshal(body, &locationsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Debug().Msgf("Retrieved %d locations for subscription %s", len(locationsResp.Value), subscriptionName)

	// Process the results
	results := make([]zoneMappingResult, 0, len(locationsResp.Value)*3)
	locationsWithZones := 0
	for _, location := range locationsResp.Value {
		locationName := derefStr(location.Name)

		// Only process locations that have availability zone mappings
		if len(location.AvailabilityZoneMappings) == 0 {
			log.Debug().Msgf("Location %s has no zone mappings", locationName)
			continue
		}

		locationsWithZones++
		log.Debug().Msgf("Location %s has %d zone mappings", locationName, len(location.AvailabilityZoneMappings))

		displayName := derefStr(location.DisplayName)

		// Extract each zone mapping for this location
		for _, mapping := range location.AvailabilityZoneMappings {
			results = append(results, zoneMappingResult{
				subscriptionID:   subscriptionID,
				subscriptionName: subscriptionName,
				location:         locationName,
				displayName:      displayName,
				logicalZone:      derefStr(mapping.LogicalZone),
				physicalZone:     derefStr(mapping.PhysicalZone),
			})
		}
	}

	log.Debug().Msgf("Subscription %s: found %d locations with zones, extracted %d zone mappings", subscriptionName, locationsWithZones, len(results))

	return results, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("zone-mapping", NewScanner())
}
