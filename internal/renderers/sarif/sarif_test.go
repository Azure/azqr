// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sarif

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/renderers"
)

func TestCreateReport(t *testing.T) {
	output := filepath.Join(t.TempDir(), "report")
	data := &renderers.ReportData{
		OutputFileName: output,
		ScopeID:       "scope",
		Summary: &findings.Summary{
			RecommendationsFound: 1,
			Recommendations: []findings.Recommendation{
				{ID: "rec-high", Recommendation: "Fix it", Impact: "High", Category: "Security", ResourceType: "Microsoft.Test/widgets", ImpactedResources: 2},
				{ID: "rec-zero", Recommendation: "No impact", Impact: "Low"},
			},
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
	results := run["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	result := results[0].(map[string]any)
	if result["level"] != "error" || result["ruleId"] != "rec-high" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestFingerprintIsStableAndScoped(t *testing.T) {
	first := fingerprint("scope", "REC")
	if first != fingerprint("scope", "rec") {
		t.Fatal("fingerprint should normalize recommendation ID case")
	}
	if first == fingerprint("other", "rec") {
		t.Fatal("fingerprint should include scope")
	}
}
