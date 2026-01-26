// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package registry

import (
	"testing"
)

func TestGetAllScanners(t *testing.T) {
	scanners := GetAllScanners()
	if len(scanners) == 0 {
		t.Error("Expected scanners to be registered, got 0")
	}
	t.Logf("Total scanners registered: %d", len(scanners))
}

func TestGetScannerKeys(t *testing.T) {
	keys := GetScannerKeys()
	if len(keys) == 0 {
		t.Error("Expected scanner keys, got 0")
	}

	// Check that keys are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Errorf("Keys are not sorted: %s > %s", keys[i-1], keys[i])
		}
	}

	t.Logf("Total scanner keys: %d", len(keys))
}

func TestGetScannerByKey(t *testing.T) {
	// Test with known scanner key
	scanners := GetScannerByKey("aa")
	if len(scanners) == 0 {
		t.Error("Expected to find 'aa' scanner")
	}

	// Test with non-existent key
	scanners = GetScannerByKey("nonexistent")
	if len(scanners) > 0 {
		t.Error("Expected nil for non-existent key")
	}
}

func TestGetScannerCount(t *testing.T) {
	count := GetScannerCount()
	if count == 0 {
		t.Error("Expected positive scanner count")
	}
	t.Logf("Scanner count: %d", count)
}

func TestGetScannersByKeys(t *testing.T) {
	// Test with empty keys (should return all)
	allScanners := GetScannersByKeys([]string{})
	if len(allScanners) == 0 {
		t.Error("Expected all scanners when keys is empty")
	}

	// Test with specific keys
	specificScanners := GetScannersByKeys([]string{"aa", "kv"})
	if len(specificScanners) == 0 {
		t.Error("Expected scanners for aa and kv")
	}

	// Verify we got fewer scanners than all
	if len(specificScanners) >= len(allScanners) {
		t.Error("Expected fewer scanners when filtering by keys")
	}
}

func TestListScannerInfo(t *testing.T) {
	info := ListScannerInfo()
	if len(info) == 0 {
		t.Error("Expected scanner info")
	}

	// Verify each info entry has required fields
	for _, si := range info {
		if si.Key == "" {
			t.Error("Scanner info missing key")
		}
		if len(si.ResourceTypes) == 0 {
			t.Errorf("Scanner %s has no resource types", si.Key)
		}
		if si.ScannerCount == 0 {
			t.Errorf("Scanner %s has zero scanner count", si.Key)
		}
	}

	t.Logf("Scanner info entries: %d", len(info))
}

func TestGetResourceTypeToScannerMap(t *testing.T) {
	resourceMap := GetResourceTypeToScannerMap()
	if len(resourceMap) == 0 {
		t.Error("Expected resource type to scanner mapping")
	}

	// Verify scanners are sorted for each resource type
	for resourceType, scannerKeys := range resourceMap {
		if len(scannerKeys) == 0 {
			t.Errorf("Resource type %s has no scanners", resourceType)
		}
		for i := 1; i < len(scannerKeys); i++ {
			if scannerKeys[i-1] > scannerKeys[i] {
				t.Errorf("Scanner keys not sorted for resource type %s", resourceType)
			}
		}
	}

	t.Logf("Unique resource types: %d", len(resourceMap))
}

func TestScannerResourceTypes(t *testing.T) {
	// Verify all scanners have at least one resource type
	keys := GetScannerKeys()
	for _, key := range keys {
		scanners := GetScannerByKey(key)
		if len(scanners) == 0 {
			continue
		}

		scanner := scanners[0]
		resourceTypes := scanner.ResourceTypes()
		if len(resourceTypes) == 0 {
			t.Errorf("Scanner %s has no resource types", key)
		}
	}
}
