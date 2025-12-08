// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package zone

// zoneMappingResult represents a single logical-to-physical zone mapping for a location
type zoneMappingResult struct {
	subscriptionID   string
	subscriptionName string
	location         string
	displayName      string
	logicalZone      string
	physicalZone     string
}
