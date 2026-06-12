// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"testing"

	"github.com/Azure/azqr/internal/renderers"
)

// findHeaderIndex returns the 1-based column index of name in the first row of
// table(), or 0 if not found. Used to verify hyperlinkCol constants.
func findHeaderIndex(table [][]string, name string) int {
	if len(table) == 0 {
		return 0
	}
	for i, h := range table[0] {
		if h == name {
			return i + 1 // 1-based
		}
	}
	return 0
}

func emptyReportData() *renderers.ReportData {
	return &renderers.ReportData{}
}

// TestHyperlinkColRecommendations verifies that the constant points at "Read More".
func TestHyperlinkColRecommendations_PointsAtReadMore(t *testing.T) {
	rd := emptyReportData()
	table := rd.RecommendationsTable()
	idx := findHeaderIndex(table, "Read More")
	if idx == 0 {
		t.Fatal("column 'Read More' not found in RecommendationsTable headers")
	}
	if idx != hyperlinkColRecommendations {
		t.Errorf("hyperlinkColRecommendations=%d but 'Read More' is at col %d",
			hyperlinkColRecommendations, idx)
	}
}

// TestHyperlinkColImpacted verifies that the constant points at "Learn".
func TestHyperlinkColImpacted_PointsAtLearn(t *testing.T) {
	rd := emptyReportData()
	table := rd.ImpactedTable()
	idx := findHeaderIndex(table, "Learn")
	if idx == 0 {
		t.Fatal("column 'Learn' not found in ImpactedTable headers")
	}
	if idx != hyperlinkColImpacted {
		t.Errorf("hyperlinkColImpacted=%d but 'Learn' is at col %d",
			hyperlinkColImpacted, idx)
	}
}

// TestHyperlinkColDefenderRecommendations verifies the constant points at "AzPortal Link".
func TestHyperlinkColDefenderRecommendations_PointsAtAzPortalLink(t *testing.T) {
	rd := emptyReportData()
	table := rd.DefenderRecommendationsTable()
	idx := findHeaderIndex(table, "AzPortal Link")
	if idx == 0 {
		t.Fatal("column 'AzPortal Link' not found in DefenderRecommendationsTable headers")
	}
	if idx != hyperlinkColDefenderRecommendations {
		t.Errorf("hyperlinkColDefenderRecommendations=%d but 'AzPortal Link' is at col %d",
			hyperlinkColDefenderRecommendations, idx)
	}
}

// TestHyperlinkColResources_DocumentedNoOp verifies that hyperlinkColResources
// (col 12) is intentionally beyond the 10-column ResourcesTable. If someone
// ever extends ResourcesTable to 12+ columns and adds a URL column at col 12,
// this test will remind them to update the constant and add a real hyperlink.
func TestHyperlinkColResources_BeyondTableWidth(t *testing.T) {
	rd := emptyReportData()
	table := rd.ResourcesTable()
	if len(table) == 0 {
		t.Fatal("ResourcesTable returned empty slice")
	}
	tableWidth := len(table[0])
	if hyperlinkColResources <= tableWidth {
		// The table has grown to include a URL column — update the constant and
		// the comment in excel.go to point at the correct column name.
		t.Errorf("hyperlinkColResources=%d is now within the %d-column ResourcesTable: "+
			"update the constant to point at the URL column",
			hyperlinkColResources, tableWidth)
	}
}
