// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"sort"
)

// ScannerList is a map of service abbreviation to scanner
var ScannerList = map[string][]IAzureScanner{}

// GetScanners returns a list of all scanners in ScannerList
func GetScanners() ([]string, []IAzureScanner) {
	var scanners []IAzureScanner
	keys := make([]string, 0, len(ScannerList))
	for key := range ScannerList {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		scanners = append(scanners, ScannerList[key]...)
	}
	return keys, scanners
}
