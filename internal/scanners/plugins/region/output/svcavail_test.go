// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package output

import (
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

func TestBuildSvcAvailSheets_RanksUnknownBetweenAvailableAndMissing(t *testing.T) {
	inv := &types.ResourceInventory{
		ResourceTypes:         map[string]int{"microsoft.compute/disks": 1, "microsoft.web/sites": 1, "microsoft.sql/servers": 1},
		SKUsByType:            map[string]map[string]int{},
		ResourceTypesByRegion: map[string]map[string]int{"westeurope": {"microsoft.compute/disks": 1, "microsoft.web/sites": 1, "microsoft.sql/servers": 1}},
	}
	results := []types.RegionComparison{
		{TargetRegion: "swedencentral", UnknownResourceTypes: []string{"microsoft.compute/disks", "microsoft.web/sites"}},
		{TargetRegion: "swedencentral", MissingResourceTypes: []string{"microsoft.web/sites"}},
	}
	sheets := BuildSvcAvailSheets(results, inv)
	if len(sheets) != 1 {
		t.Fatalf("got %d sheets", len(sheets))
	}
	got := map[string]string{}
	for _, row := range sheets[0].Table[1:] {
		got[row[0]] = row[len(row)-1]
	}
	want := map[string]string{
		"microsoft.compute/disks": "Unknown",
		"microsoft.web/sites":     "Not available",
		"microsoft.sql/servers":   "Available",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestBuildSvcAvailSheets_SKUListedByAnySubscriptionIsAvailable(t *testing.T) {
	rt := "microsoft.cognitiveservices/accounts"
	inv := &types.ResourceInventory{
		ResourceTypes:         map[string]int{rt: 1},
		SKUsByType:            map[string]map[string]int{rt: {"S1": 1}},
		ResourceTypesByRegion: map[string]map[string]int{"westeurope": {rt: 1}},
	}
	results := []types.RegionComparison{
		{TargetRegion: "swedencentral", MissingSKUs: []string{rt + ":S1"}},
		{TargetRegion: "swedencentral", OfferedSKUs: []string{rt + ":S1"}},
	}
	rows := BuildSvcAvailSheets(results, inv)[0].Table
	if got := rows[1][5]; got != "Available" {
		t.Errorf("SKU available = %q, want Available", got)
	}
}
