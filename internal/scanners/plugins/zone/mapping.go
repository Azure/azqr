// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package zone

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/rs/zerolog/log"
)

// ZoneMappingScanner is an internal plugin that retrieves availability zone mappings
type ZoneMappingScanner struct{}

// NewZoneMappingScanner creates a new zone mapping scanner
func NewZoneMappingScanner() *ZoneMappingScanner {
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
func (s *ZoneMappingScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Scanning availability zone mappings across subscriptions")

	// Initialize table with headers
	table := [][]string{
		{"Subscription", "Location", "Display Name", "Logical Zone", "Physical Zone"},
	}

	// Use parallel processing for performance
	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make([]zoneMappingResult, 0)

	// Process subscriptions in parallel with worker pool
	const maxWorkers = 5
	semaphore := make(chan struct{}, maxWorkers)

	for subID, subName := range subscriptions {
		wg.Add(1)
		go func(subscriptionID, subscriptionName string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Debug().Msgf("Fetching zone mappings for subscription: %s", subscriptionName)

			subResults, err := s.fetchZoneMappings(ctx, cred, subscriptionID, subscriptionName)
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

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "Zone Mapping",
		Description: "Logical-to-physical availability zone mappings for Azure regions by subscription",
		Table:       table,
	}, nil
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

// fetchZoneMappings retrieves zone mappings for a single subscription using direct REST API call
func (s *ZoneMappingScanner) fetchZoneMappings(ctx context.Context, cred azcore.TokenCredential, subscriptionID, subscriptionName string) ([]zoneMappingResult, error) {
	// Apply throttling to respect ARM API limits
	_ = throttling.WaitARM(ctx) // nolint:errcheck

	// Create pipeline with authentication
	clientOptions := policy.ClientOptions{
		PerCallPolicies: []policy.Policy{
			runtime.NewBearerTokenPolicy(cred, []string{"https://management.azure.com/.default"}, nil),
		},
	}
	pipeline := runtime.NewPipeline("azqr", "1.0.0", runtime.PipelineOptions{}, &clientOptions)

	// Construct the REST API URL
	// GET /subscriptions/{subscriptionId}/locations?api-version=2022-12-01
	endpoint := fmt.Sprintf("https://management.azure.com/subscriptions/%s/locations?api-version=2022-12-01", subscriptionID)

	// Create the request
	req, err := runtime.NewRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request through the pipeline
	resp, err := pipeline.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // nolint:errcheck
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var locationsResp locationResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &locationsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Debug().Msgf("Retrieved %d locations for subscription %s", len(locationsResp.Value), subscriptionName)

	// Process the results
	results := make([]zoneMappingResult, 0)
	locationsWithZones := 0
	for _, location := range locationsResp.Value {
		// Debug: log each location
		locationName := ""
		if location.Name != nil {
			locationName = *location.Name
		}

		// Only process locations that have availability zone mappings
		if len(location.AvailabilityZoneMappings) == 0 {
			log.Debug().Msgf("Location %s has no zone mappings", locationName)
			continue
		}

		locationsWithZones++
		log.Debug().Msgf("Location %s has %d zone mappings", locationName, len(location.AvailabilityZoneMappings))

		displayName := ""
		if location.DisplayName != nil {
			displayName = *location.DisplayName
		}

		// Extract each zone mapping for this location
		for _, mapping := range location.AvailabilityZoneMappings {
			logicalZone := ""
			if mapping.LogicalZone != nil {
				logicalZone = *mapping.LogicalZone
			}

			physicalZone := ""
			if mapping.PhysicalZone != nil {
				physicalZone = *mapping.PhysicalZone
			}

			results = append(results, zoneMappingResult{
				subscriptionID:   subscriptionID,
				subscriptionName: subscriptionName,
				location:         locationName,
				displayName:      displayName,
				logicalZone:      logicalZone,
				physicalZone:     physicalZone,
			})
		}
	}

	log.Debug().Msgf("Subscription %s: found %d locations with zones, extracted %d zone mappings", subscriptionName, locationsWithZones, len(results))

	return results, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("zone-mapping", NewZoneMappingScanner())
}
