// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package sarif renders azqr findings as SARIF 2.1.0.
package sarif

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

const schemaURL = "https://json.schemastore.org/sarif-2.1.0.json"

type logFile struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []run  `json:"runs"`
}

type run struct {
	Tool       tool       `json:"tool"`
	Results    []result   `json:"results"`
	Automation automation `json:"automationDetails,omitempty"`
}

type automation struct {
	ID string `json:"id,omitempty"`
}

type tool struct {
	Driver driver `json:"driver"`
}

type driver struct {
	Name           string `json:"name"`
	Version        string `json:"version,omitempty"`
	InformationURI string `json:"informationUri"`
	Rules          []rule `json:"rules"`
}

type rule struct {
	ID               string            `json:"id"`
	ShortDescription message           `json:"shortDescription"`
	FullDescription  message           `json:"fullDescription,omitempty"`
	HelpURI          string            `json:"helpUri,omitempty"`
	Properties       map[string]string `json:"properties"`
}

type result struct {
	RuleID              string            `json:"ruleId"`
	Level               string            `json:"level"`
	Message             message           `json:"message"`
	PartialFingerprints map[string]string `json:"partialFingerprints"`
	Locations           []location        `json:"locations"`
}

type message struct {
	Text string `json:"text"`
}

type location struct {
	LogicalLocations []logicalLocation `json:"logicalLocations"`
}

type logicalLocation struct {
	Name               string `json:"name"`
	FullyQualifiedName string `json:"fullyQualifiedName"`
	Kind               string `json:"kind"`
}

// CreateReport writes a SARIF report with one result per impacted resource.
func CreateReport(data *renderers.ReportData, version string) error {
	if data == nil || data.Summary == nil {
		return fmt.Errorf("SARIF rendering requires a findings summary")
	}

	// Build rules and an index of which recommendation IDs have impacted resources.
	impacted := make(map[string]findings.Recommendation, data.Summary.RecommendationsFound)
	rules := make([]rule, 0, data.Summary.RecommendationsFound)
	for _, rec := range data.Summary.Recommendations {
		if rec.ImpactedResources == 0 {
			continue
		}
		impacted[strings.ToLower(rec.ID)] = rec
		rules = append(rules, buildRule(rec))
	}

	// One result per (recommendation, resource) pair, deduplicated.
	type seenKey struct{ ruleID, resourceID string }
	seen := make(map[seenKey]struct{}, data.Summary.ImpactedResources)
	results := make([]result, 0, data.Summary.ImpactedResources)
	for _, g := range data.Graph {
		if g == nil || g.Category == models.CategorySLA {
			continue
		}
		ruleID := strings.ToLower(g.RecommendationID)
		if _, ok := impacted[ruleID]; !ok {
			continue
		}
		key := seenKey{ruleID, strings.ToLower(g.ResourceID)}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, buildResult(g))
	}

	output := logFile{
		Version: "2.1.0",
		Schema:  schemaURL,
		Runs: []run{{
			Tool: tool{Driver: driver{
				Name:           "azqr",
				Version:        version,
				InformationURI: "https://github.com/Azure/azqr",
				Rules:          rules,
			}},
			Results:    results,
			Automation: automation{ID: "azqr/" + data.ScopeID},
		}},
	}

	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal SARIF report: %w", err)
	}
	filename := data.OutputFileName + ".sarif"
	if err := os.WriteFile(filename, encoded, 0600); err != nil {
		return fmt.Errorf("write SARIF report %s: %w", filename, err)
	}
	return nil
}

func buildRule(rec findings.Recommendation) rule {
	return rule{
		ID:               rec.ID,
		ShortDescription: message{Text: rec.Recommendation},
		FullDescription:  message{Text: rec.LongDescription},
		HelpURI:          rec.LearnURL,
		Properties: map[string]string{
			"category":     rec.Category,
			"impact":       rec.Impact,
			"resourceType": rec.ResourceType,
			"source":       rec.Source,
		},
	}
}

func buildResult(g *models.GraphResult) result {
	return result{
		RuleID: g.RecommendationID,
		Level:  level(string(g.Impact)),
		Message: message{Text: fmt.Sprintf(
			"%s: %s",
			g.Recommendation,
			g.ResourceID,
		)},
		PartialFingerprints: map[string]string{
			"azqrFinding/v1": fingerprint(g.RecommendationID, g.ResourceID),
		},
		Locations: []location{{
			LogicalLocations: []logicalLocation{{
				Name:               g.Name,
				FullyQualifiedName: g.ResourceID,
				Kind:               "resource",
			}},
		}},
	}
}

func level(impact string) string {
	rank, _ := models.SeverityRank(impact)
	switch rank {
	case 3:
		return "error"
	case 2:
		return "warning"
	default:
		return "note"
	}
}

func fingerprint(recommendationID, resourceID string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(recommendationID) + "\x00" + strings.ToLower(resourceID)))
	return hex.EncodeToString(sum[:16])
}
