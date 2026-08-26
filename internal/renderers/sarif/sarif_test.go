// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sarif

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

func TestCreateReport(t *testing.T) {
	output := filepath.Join(t.TempDir(), "report")
	data := &renderers.ReportData{
		OutputFileName: output,
		ScopeID:        "sub-1,sub-2",
		Summary: &findings.Summary{
			RecommendationsFound: 1,
			ImpactedResources:    2,
			Recommendations: []findings.Recommendation{
				{ID: "rec-high", Recommendation: "Fix it", Impact: "High", Category: "Security", ResourceType: "Microsoft.Test/widgets", ImpactedResources: 2},
				{ID: "rec-zero", Recommendation: "No impact", Impact: "Low"},
			},
		},
		Graph: []*models.GraphResult{
			{RecommendationID: "rec-high", ResourceID: "/subscriptions/sub-1/resourceGroups/rg/providers/Microsoft.Test/widgets/a", Name: "a", Recommendation: "Fix it", Impact: "High"},
			{RecommendationID: "rec-high", ResourceID: "/subscriptions/sub-1/resourceGroups/rg/providers/Microsoft.Test/widgets/b", Name: "b", Recommendation: "Fix it", Impact: "High"},
			{RecommendationID: "rec-high", ResourceID: "/subscriptions/sub-1/resourceGroups/rg/providers/Microsoft.Test/widgets/a", Name: "a", Recommendation: "Fix it", Impact: "High"}, // duplicate — should be deduplicated
			{RecommendationID: "rec-zero", ResourceID: "/subscriptions/sub-1/resourceGroups/rg/providers/Microsoft.Test/widgets/c", Name: "c"},                                            // not impacted — should be excluded
		},
	}
	if err := CreateReport(data, "test"); err != nil {
		t.Fatal(err)
	}

	encoded, err := os.ReadFile(output + ".sarif")
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(encoded, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	runs := report["runs"].([]any)
	run := runs[0].(map[string]any)

	// Two unique (rec-high, resource) pairs; duplicate and zero-impact filtered out.
	results := run["results"].([]any)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		res := r.(map[string]any)
		if res["level"] != "error" || res["ruleId"] != "rec-high" {
			t.Fatalf("unexpected result: %+v", res)
		}
		locs := res["locations"].([]any)[0].(map[string]any)["logicalLocations"].([]any)
		fqn := locs[0].(map[string]any)["fullyQualifiedName"].(string)
		if fqn == "" {
			t.Fatal("fullyQualifiedName should be the resource ID")
		}
	}

	// automationDetails should carry a stable scope digest.
	automation := run["automationDetails"].(map[string]any)
	id := automation["id"].(string)
	if id == "" || !strings.HasPrefix(id, "azqr/") {
		t.Fatalf("unexpected automationDetails.id: %v", id)
	}
}

func TestFingerprintIsStableAndUnique(t *testing.T) {
	fp := fingerprint("REC-001", "/subscriptions/sub/resourceGroups/rg/providers/Type/name")
	// Case-insensitive on recommendation ID.
	if fp != fingerprint("rec-001", "/subscriptions/sub/resourceGroups/rg/providers/Type/name") {
		t.Fatal("fingerprint should normalize recommendation ID case")
	}
	// Different resource → different fingerprint.
	if fp == fingerprint("rec-001", "/subscriptions/sub/resourceGroups/rg/providers/Type/other") {
		t.Fatal("fingerprint should differ for different resources")
	}
	// Different recommendation → different fingerprint.
	if fp == fingerprint("rec-002", "/subscriptions/sub/resourceGroups/rg/providers/Type/name") {
		t.Fatal("fingerprint should differ for different recommendations")
	}
}
