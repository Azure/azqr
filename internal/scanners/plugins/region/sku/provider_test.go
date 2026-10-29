// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

func TestMergeBestSKU_KeepsBestState(t *testing.T) {
	m := map[string]types.SKUAvailability{}
	mergeBestSKU(m, "premium_lrs", types.SKUAvailability{State: types.SKUUnavailable})
	mergeBestSKU(m, "premium_lrs", types.SKUAvailability{State: types.SKUAvailable})
	mergeBestSKU(m, "premium_lrs", types.SKUAvailability{State: types.SKURestricted})
	if m["premium_lrs"].State != types.SKUAvailable {
		t.Errorf("state = %v, want available", m["premium_lrs"].State)
	}
}

func TestExtractDiskSKU_IgnoresSnapshots(t *testing.T) {
	if _, _, ok := extractDiskSKU(diskSKUItem{ResourceType: "snapshots", Name: "Premium_LRS"}); ok {
		t.Error("snapshot entry should be ignored")
	}
	if name, _, ok := extractDiskSKU(diskSKUItem{ResourceType: "disks", Name: "Premium_LRS"}); !ok || name != "Premium_LRS" {
		t.Errorf("disk entry = %q, %v", name, ok)
	}
}
