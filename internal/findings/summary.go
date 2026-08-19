// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package findings provides the canonical summary of azqr recommendations.
package findings

import (
	"sort"
	"strings"

	"github.com/Azure/azqr/internal/models"
)

// Recommendation is the normalized summary of one recommendation.
type Recommendation struct {
	ID                string `json:"id"`
	Recommendation    string `json:"recommendation"`
	Source            string `json:"source"`
	Category          string `json:"category"`
	Impact            string `json:"impact"`
	ResourceType      string `json:"resourceType"`
	LongDescription   string `json:"longDescription,omitempty"`
	LearnURL          string `json:"learnUrl,omitempty"`
	ImpactedResources int    `json:"impactedResources"`
}

// Summary contains deterministic recommendation and finding aggregates.
type Summary struct {
	Recommendations      []Recommendation `json:"recommendations"`
	Resources            int              `json:"resources"`
	ImpactedResources    int              `json:"impactedResources"`
	ImpactedByImpact     map[string]int   `json:"impactedByImpact"`
	ImpactedByCategory   map[string]int   `json:"impactedByCategory"`
	RecommendationsFound int              `json:"recommendationsFound"`
}

type findingKey struct {
	recommendationID string
	resourceID       string
}

// Build creates a canonical summary from recommendation definitions and graph results.
func Build(
	definitions map[string]map[string]*models.GraphRecommendation,
	results []*models.GraphResult,
	resourceCount int,
) *Summary {
	records := make(map[string]Recommendation)

	for _, byID := range definitions {
		for _, definition := range byID {
			if definition == nil || definition.Category == string(models.CategorySLA) {
				continue
			}
			records[normalize(definition.RecommendationID)] = recommendationFromDefinition(definition)
		}
	}

	counts := make(map[string]int)
	seen := make(map[findingKey]struct{}, len(results))
	for _, result := range results {
		if result == nil || result.Category == models.CategorySLA {
			continue
		}

		id := normalize(result.RecommendationID)
		if id == "" {
			continue
		}
		if _, ok := records[id]; !ok {
			records[id] = recommendationFromResult(result)
		}

		key := findingKey{
			recommendationID: id,
			resourceID:       normalize(result.ResourceID),
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		counts[id]++
	}

	recommendations := make([]Recommendation, 0, len(records))
	summary := &Summary{
		Resources:          resourceCount,
		ImpactedByImpact:   make(map[string]int),
		ImpactedByCategory: make(map[string]int),
	}
	for id, record := range records {
		record.ImpactedResources = counts[id]
		recommendations = append(recommendations, record)
		if record.ImpactedResources == 0 {
			continue
		}
		summary.RecommendationsFound++
		summary.ImpactedResources += record.ImpactedResources
		summary.ImpactedByImpact[record.Impact] += record.ImpactedResources
		summary.ImpactedByCategory[record.Category] += record.ImpactedResources
	}

	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].ResourceType != recommendations[j].ResourceType {
			return recommendations[i].ResourceType < recommendations[j].ResourceType
		}
		return recommendations[i].ID < recommendations[j].ID
	})
	summary.Recommendations = recommendations
	return summary
}

func recommendationFromDefinition(definition *models.GraphRecommendation) Recommendation {
	learnURL := ""
	if len(definition.LearnMoreLink) > 0 {
		learnURL = definition.LearnMoreLink[0].Url
	}
	return Recommendation{
		ID:              definition.RecommendationID,
		Recommendation:  definition.Recommendation,
		Source:          definition.Source,
		Category:        definition.Category,
		Impact:          definition.Impact,
		ResourceType:    definition.ResourceType,
		LongDescription: definition.LongDescription,
		LearnURL:        learnURL,
	}
}

func recommendationFromResult(result *models.GraphResult) Recommendation {
	return Recommendation{
		ID:              result.RecommendationID,
		Recommendation:  result.Recommendation,
		Source:          result.Source,
		Category:        string(result.Category),
		Impact:          string(result.Impact),
		ResourceType:    result.ResourceType,
		LongDescription: result.LongDescription,
		LearnURL:        result.Learn,
	}
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
