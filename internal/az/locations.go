// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package az provides shared Azure REST API types and helpers.
// This file defines the ARM Locations API response types used by both the
// zone-mapping plugin and the region-selection plugin.
package az

import (
	"encoding/json"
	"fmt"
)

// LocationsResponse is the ARM JSON envelope for GET /subscriptions/{id}/locations.
type LocationsResponse struct {
	Value []LocationInfo `json:"value"`
}

// LocationInfo holds the name, display name, metadata, and AZ mappings for one Azure region.
type LocationInfo struct {
	Name                     string                    `json:"name"`
	DisplayName              string                    `json:"displayName"`
	Metadata                 LocationMetadata          `json:"metadata"`
	AvailabilityZoneMappings []AvailabilityZoneMapping `json:"availabilityZoneMappings"`
}

// LocationMetadata carries the region classification (Physical vs. Logical).
type LocationMetadata struct {
	RegionType string `json:"regionType"`
}

// AvailabilityZoneMapping is one logical → physical zone pair for a location.
// LogicalZone is the subscription-scoped zone number ("1", "2", "3").
// PhysicalZone is the datacenter label (e.g. "eastus-az1").
type AvailabilityZoneMapping struct {
	LogicalZone  string `json:"logicalZone"`
	PhysicalZone string `json:"physicalZone"`
}

// ParseLocations unmarshals the raw bytes from GET /subscriptions/{id}/locations and
// returns the full list of LocationInfo entries. Returns an error if the JSON is malformed.
func ParseLocations(body []byte) ([]LocationInfo, error) {
	var resp LocationsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse locations response: %w", err)
	}
	return resp.Value, nil
}

// ZoneMappingForRegion builds a logical → physical zone map for a single region name
// from a slice of LocationInfo. Returns nil when the region has no AZ mappings.
func ZoneMappingForRegion(locations []LocationInfo, regionName string) map[string]string {
	for _, loc := range locations {
		if loc.Name == regionName && len(loc.AvailabilityZoneMappings) > 0 {
			m := make(map[string]string, len(loc.AvailabilityZoneMappings))
			for _, zm := range loc.AvailabilityZoneMappings {
				m[zm.LogicalZone] = zm.PhysicalZone
			}
			return m
		}
	}
	return nil
}
