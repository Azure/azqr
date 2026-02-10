// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package registry provides centralized scanner registration.
// This package acts as a single import point for all scanner packages,
// eliminating the need for 51+ blank imports in the main scanner code.
package registry

import (
	"github.com/Azure/azqr/internal/models"
)

// GetAllScanners returns all registered scanners as a flat list
// This is the primary interface for accessing scanners
func GetAllScanners() []models.IAzureScanner {
	_, scanners := models.GetScanners()
	return scanners
}

// GetScannerKeys returns all registered scanner keys (service abbreviations)
func GetScannerKeys() []string {
	keys, _ := models.GetScanners()
	return keys
}

// GetScannerByKey returns scanners for a specific service abbreviation
// Returns nil if the key doesn't exist
func GetScannerByKey(key string) []models.IAzureScanner {
	return models.ScannerList[key]
}

// GetScannerCount returns the total number of registered scanner instances
func GetScannerCount() int {
	return len(GetAllScanners())
}

// GetScannerInfo returns metadata about all registered scanners
type ScannerInfo struct {
	Key           string
	ResourceTypes []string
	ScannerCount  int
}

// ListScannerInfo returns metadata about all registered scanners
func ListScannerInfo() []ScannerInfo {
	var info []ScannerInfo
	keys := GetScannerKeys()

	for _, key := range keys {
		scanners := GetScannerByKey(key)
		if len(scanners) > 0 {
			// Get resource types from first scanner (they should all be the same for a key)
			resourceTypes := scanners[0].ResourceTypes()
			info = append(info, ScannerInfo{
				Key:           key,
				ResourceTypes: resourceTypes,
				ScannerCount:  len(scanners),
			})
		}
	}

	return info
}
