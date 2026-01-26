// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package registry provides centralized scanner registration.
// This package acts as a single import point for all scanner packages,
// eliminating the need for 51+ blank imports in the main scanner code.
package registry

import (
	"sort"

	"github.com/Azure/azqr/internal/models"

	// Scanner packages - automatically register themselves via init()
	_ "github.com/Azure/azqr/internal/scanners/aa"
	_ "github.com/Azure/azqr/internal/scanners/adf"
	_ "github.com/Azure/azqr/internal/scanners/afd"
	_ "github.com/Azure/azqr/internal/scanners/afw"
	_ "github.com/Azure/azqr/internal/scanners/agw"
	_ "github.com/Azure/azqr/internal/scanners/aif"
	_ "github.com/Azure/azqr/internal/scanners/aks"
	_ "github.com/Azure/azqr/internal/scanners/amg"
	_ "github.com/Azure/azqr/internal/scanners/apim"
	_ "github.com/Azure/azqr/internal/scanners/appcs"
	_ "github.com/Azure/azqr/internal/scanners/appi"
	_ "github.com/Azure/azqr/internal/scanners/arc"
	_ "github.com/Azure/azqr/internal/scanners/as"
	_ "github.com/Azure/azqr/internal/scanners/asp"
	_ "github.com/Azure/azqr/internal/scanners/avail"
	_ "github.com/Azure/azqr/internal/scanners/avd"
	_ "github.com/Azure/azqr/internal/scanners/avs"
	_ "github.com/Azure/azqr/internal/scanners/ba"
	_ "github.com/Azure/azqr/internal/scanners/ca"
	_ "github.com/Azure/azqr/internal/scanners/cae"
	_ "github.com/Azure/azqr/internal/scanners/ci"
	_ "github.com/Azure/azqr/internal/scanners/conn"
	_ "github.com/Azure/azqr/internal/scanners/cosmos"
	_ "github.com/Azure/azqr/internal/scanners/cr"
	_ "github.com/Azure/azqr/internal/scanners/dbw"
	_ "github.com/Azure/azqr/internal/scanners/dec"
	_ "github.com/Azure/azqr/internal/scanners/disk"
	_ "github.com/Azure/azqr/internal/scanners/erc"
	_ "github.com/Azure/azqr/internal/scanners/evgd"
	_ "github.com/Azure/azqr/internal/scanners/evh"
	_ "github.com/Azure/azqr/internal/scanners/fabric"
	_ "github.com/Azure/azqr/internal/scanners/fdfp"
	_ "github.com/Azure/azqr/internal/scanners/gal"
	_ "github.com/Azure/azqr/internal/scanners/hpc"
	_ "github.com/Azure/azqr/internal/scanners/hub"
	_ "github.com/Azure/azqr/internal/scanners/iot"
	_ "github.com/Azure/azqr/internal/scanners/it"
	_ "github.com/Azure/azqr/internal/scanners/kv"
	_ "github.com/Azure/azqr/internal/scanners/lb"
	_ "github.com/Azure/azqr/internal/scanners/log"
	_ "github.com/Azure/azqr/internal/scanners/logic"
	_ "github.com/Azure/azqr/internal/scanners/maria"
	_ "github.com/Azure/azqr/internal/scanners/mysql"
	_ "github.com/Azure/azqr/internal/scanners/netapp"
	_ "github.com/Azure/azqr/internal/scanners/ng"
	_ "github.com/Azure/azqr/internal/scanners/nic"
	_ "github.com/Azure/azqr/internal/scanners/nsg"
	_ "github.com/Azure/azqr/internal/scanners/nw"
	_ "github.com/Azure/azqr/internal/scanners/odb"
	_ "github.com/Azure/azqr/internal/scanners/pdnsz"
	_ "github.com/Azure/azqr/internal/scanners/pep"
	_ "github.com/Azure/azqr/internal/scanners/pip"
	_ "github.com/Azure/azqr/internal/scanners/psql"
	_ "github.com/Azure/azqr/internal/scanners/redis"
	_ "github.com/Azure/azqr/internal/scanners/rg"
	_ "github.com/Azure/azqr/internal/scanners/rsv"
	_ "github.com/Azure/azqr/internal/scanners/rt"
	_ "github.com/Azure/azqr/internal/scanners/sap"
	_ "github.com/Azure/azqr/internal/scanners/sb"
	_ "github.com/Azure/azqr/internal/scanners/sigr"
	_ "github.com/Azure/azqr/internal/scanners/sql"
	_ "github.com/Azure/azqr/internal/scanners/srch"
	_ "github.com/Azure/azqr/internal/scanners/st"
	_ "github.com/Azure/azqr/internal/scanners/synw"
	_ "github.com/Azure/azqr/internal/scanners/traf"
	_ "github.com/Azure/azqr/internal/scanners/vdpool"
	_ "github.com/Azure/azqr/internal/scanners/vgw"
	_ "github.com/Azure/azqr/internal/scanners/vm"
	_ "github.com/Azure/azqr/internal/scanners/vmss"
	_ "github.com/Azure/azqr/internal/scanners/vnet"
	_ "github.com/Azure/azqr/internal/scanners/vwan"
	_ "github.com/Azure/azqr/internal/scanners/wps"
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

// GetScannersByKeys returns scanners for the specified service abbreviations
// If keys is empty, returns all scanners
func GetScannersByKeys(keys []string) []models.IAzureScanner {
	if len(keys) == 0 {
		return GetAllScanners()
	}

	var scanners []models.IAzureScanner
	for _, key := range keys {
		if s := GetScannerByKey(key); s != nil {
			scanners = append(scanners, s...)
		}
	}
	return scanners
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

// GetResourceTypeToScannerMap returns a map of resource type to scanner keys
// Useful for finding which scanner handles a specific resource type
func GetResourceTypeToScannerMap() map[string][]string {
	resourceTypeMap := make(map[string][]string)

	for _, key := range GetScannerKeys() {
		scanners := GetScannerByKey(key)
		if len(scanners) > 0 {
			for _, resourceType := range scanners[0].ResourceTypes() {
				resourceTypeMap[resourceType] = append(resourceTypeMap[resourceType], key)
			}
		}
	}

	// Sort the scanner keys for each resource type
	for resourceType := range resourceTypeMap {
		sort.Strings(resourceTypeMap[resourceType])
	}

	return resourceTypeMap
}
