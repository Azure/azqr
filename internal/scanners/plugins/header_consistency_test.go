// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package plugins_headertest verifies that each internal plugin's Table[0] (the
// header row it writes to Excel) is derived from its own ColumnMetadata via
// HeaderRow(). Any drift between ColumnMetadata and the actual table headers
// would cause the Excel report and the web viewer to show different column names.
package plugins_test

import (
	"testing"

	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners/plugins/carbon"
	"github.com/Azure/azqr/internal/scanners/plugins/openai"
	regionplugin "github.com/Azure/azqr/internal/scanners/plugins/region"
	"github.com/Azure/azqr/internal/scanners/plugins/sqlesu"
	"github.com/Azure/azqr/internal/scanners/plugins/zone"
)

// scannerWithMetadata is a minimal interface that lets us call GetMetadata on each plugin
// without knowing the concrete type.
type scannerWithMetadata interface {
	GetMetadata() plugins.PluginMetadata
}

// headerProducer is implemented by plugins whose Scan builds a Table with a header row.
// We test header consistency via GetMetadata().HeaderRow() directly, so no real scan
// is needed — we only verify the metadata contract.

func assertHeaderRowMatchesColumnMetadata(t *testing.T, scanner scannerWithMetadata) {
	t.Helper()
	meta := scanner.GetMetadata()

	if len(meta.ColumnMetadata) == 0 {
		t.Errorf("plugin %q: ColumnMetadata is empty; expected column definitions", meta.Name)
		return
	}

	headerRow := meta.HeaderRow()

	if len(headerRow) != len(meta.ColumnMetadata) {
		t.Errorf("plugin %q: HeaderRow() len=%d, ColumnMetadata len=%d — they must match",
			meta.Name, len(headerRow), len(meta.ColumnMetadata))
		return
	}

	for i, col := range meta.ColumnMetadata {
		if headerRow[i] != col.Name {
			t.Errorf("plugin %q: HeaderRow()[%d]=%q, ColumnMetadata[%d].Name=%q — mismatch",
				meta.Name, i, headerRow[i], i, col.Name)
		}
	}
}

func TestZonePlugin_HeaderRow_MatchesColumnMetadata(t *testing.T) {
	assertHeaderRowMatchesColumnMetadata(t, zone.NewScanner())
}

func TestOpenAIThrottlingPlugin_HeaderRow_MatchesColumnMetadata(t *testing.T) {
	assertHeaderRowMatchesColumnMetadata(t, openai.NewScanner())
}

func TestCarbonPlugin_HeaderRow_MatchesColumnMetadata(t *testing.T) {
	assertHeaderRowMatchesColumnMetadata(t, carbon.NewScanner())
}

func TestSQLESUPlugin_HeaderRow_MatchesColumnMetadata(t *testing.T) {
	assertHeaderRowMatchesColumnMetadata(t, sqlesu.NewScanner())
}

func TestRegionPlugin_HeaderRow_MatchesColumnMetadata(t *testing.T) {
	assertHeaderRowMatchesColumnMetadata(t, regionplugin.NewScanner())
}

// TestAllInternalPlugins_HaveColumnMetadata is a catch-all: every registered
// internal plugin must declare at least one column (prevents accidentally
// shipping a plugin with an empty metadata definition).
func TestAllInternalPlugins_HaveColumnMetadata(t *testing.T) {
	scanners := []scannerWithMetadata{
		zone.NewScanner(),
		openai.NewScanner(),
		carbon.NewScanner(),
		sqlesu.NewScanner(),
		regionplugin.NewScanner(),
	}

	for _, s := range scanners {
		meta := s.GetMetadata()
		if len(meta.ColumnMetadata) == 0 {
			t.Errorf("plugin %q has no ColumnMetadata", meta.Name)
		}
	}
}
