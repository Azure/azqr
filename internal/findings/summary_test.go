// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package findings

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestBuildDeduplicatesAndSkipsSLA(t *testing.T) {
	definitions := map[string]map[string]*models.GraphRecommendation{
		"microsoft.test/widgets": {
			"rec-1": {
				RecommendationID: "rec-1",
				Recommendation:   "Fix widget",
				Category:         string(models.CategorySecurity),
				Impact:           string(models.ImpactHigh),
				ResourceType:     "Microsoft.Test/widgets",
			},
			"sla": {
				RecommendationID: "sla",
				Category:         string(models.CategorySLA),
				Impact:           string(models.ImpactLow),
			},
		},
	}
	results := []*models.GraphResult{
		{RecommendationID: "rec-1", ResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Test/widgets/one", Category: models.CategorySecurity, Impact: models.ImpactHigh},
		{RecommendationID: "REC-1", ResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Test/widgets/ONE", Category: models.CategorySecurity, Impact: models.ImpactHigh},
		{RecommendationID: "sla", ResourceID: "resource", Category: models.CategorySLA, Impact: models.ImpactLow},
	}

	summary := Build(definitions, results, 4)

	if len(summary.Recommendations) != 1 {
		t.Fatalf("got %d recommendations, want 1", len(summary.Recommendations))
	}
	if summary.Recommendations[0].ImpactedResources != 1 {
		t.Fatalf("got %d impacted resources, want 1", summary.Recommendations[0].ImpactedResources)
	}
	if summary.Resources != 4 || summary.ImpactedResources != 1 || summary.RecommendationsFound != 1 {
		t.Fatalf("unexpected summary totals: %+v", summary)
	}
	if summary.ImpactedByImpact[string(models.ImpactHigh)] != 1 {
		t.Fatalf("unexpected impact totals: %+v", summary.ImpactedByImpact)
	}
}

func TestBuildIncludesResultWithoutDefinition(t *testing.T) {
	summary := Build(nil, []*models.GraphResult{{
		RecommendationID: "external",
		Recommendation:   "External recommendation",
		ResourceID:       "resource",
		ResourceType:     "Microsoft.Test/widgets",
		Category:         models.CategoryGovernance,
		Impact:           models.ImpactMedium,
	}}, 1)

	if len(summary.Recommendations) != 1 || summary.Recommendations[0].ID != "external" {
		t.Fatalf("unexpected recommendations: %+v", summary.Recommendations)
	}
}
