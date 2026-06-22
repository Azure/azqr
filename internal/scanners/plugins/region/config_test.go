// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"strings"
	"testing"
)

func TestInit_LoadedEmbeddedConfig(t *testing.T) {
	// The package init() must successfully unmarshal both embedded JSON files.
	if len(skuConfigs) == 0 {
		t.Error("expected skuConfigs to be populated from embedded sku.json")
	}
	if len(propertyMapsConfig) == 0 {
		t.Fatal("expected propertyMapsConfig to be populated from embedded propertyMaps.json")
	}
	if len(propertyMapsIndex) != len(propertyMapsConfig) {
		t.Errorf("index size %d does not match config slice size %d", len(propertyMapsIndex), len(propertyMapsConfig))
	}

	// Every config must have a non-empty resource type and be reachable via the index.
	for i := range propertyMapsConfig {
		rt := propertyMapsConfig[i].ResourceType
		if rt == "" {
			t.Errorf("propertyMapsConfig[%d] has empty resourceType", i)
			continue
		}
		if getPropertyMapConfig(rt) == nil {
			t.Errorf("getPropertyMapConfig(%q) returned nil for a configured resource type", rt)
		}
	}
}

func TestGetPropertyMapConfig(t *testing.T) {
	// Use the first loaded entry as a known-present resource type.
	if len(propertyMapsConfig) == 0 {
		t.Skip("no property map configs loaded")
	}
	known := propertyMapsConfig[0].ResourceType

	t.Run("exact match", func(t *testing.T) {
		if got := getPropertyMapConfig(known); got == nil {
			t.Fatalf("expected config for %q, got nil", known)
		}
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		got := getPropertyMapConfig(strings.ToUpper(known))
		if got == nil {
			t.Fatalf("expected case-insensitive lookup of %q to succeed", known)
		}
		if !strings.EqualFold(got.ResourceType, known) {
			t.Errorf("expected resource type %q, got %q", known, got.ResourceType)
		}
	})

	t.Run("unknown resource type returns nil", func(t *testing.T) {
		if got := getPropertyMapConfig("microsoft.fake/doesnotexist"); got != nil {
			t.Errorf("expected nil for unknown resource type, got %+v", got)
		}
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		if got := getPropertyMapConfig(""); got != nil {
			t.Errorf("expected nil for empty resource type, got %+v", got)
		}
	})
}
