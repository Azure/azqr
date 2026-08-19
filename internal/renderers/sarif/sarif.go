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
	Tool       tool     `json:"tool"`
	Results    []result `json:"results"`
	Automation automation `json:"automationDetails,omitempty"`
}

type automation struct {
	ID string `json:"id,omitempty"`
}

type tool struct {
	Driver driver `json:"driver"`
}

type driver struct {
	Name            string `json:"name"`
	Version         string `json:"version,omitempty"`
	InformationURI  string `json:"informationUri"`
	Rules           []rule `json:"rules"`
}

type rule struct {
	ID               string            `json:"id"`
	ShortDescription message           `json:"shortDescription"`
	FullDescription  message           `json:"fullDescription,omitempty"`
	HelpURI           string            `json:"helpUri,omitempty"`
	Properties        map[string]string `json:"properties"`
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

// CreateReport writes a SARIF report for impacted recommendations.
func CreateReport(data *renderers.ReportData, version string) error {
	if data == nil || data.Summary == nil {
		return fmt.Errorf("SARIF rendering requires a findings summary")
	}

	rules := make([]rule, 0, data.Summary.RecommendationsFound)
	results := make([]result, 0, data.Summary.RecommendationsFound)
	for _, recommendation := range data.Summary.Recommendations {
		if recommendation.ImpactedResources == 0 {
			continue
		}
		rules = append(rules, buildRule(recommendation))
		results = append(results, buildResult(recommendation, data.ScopeID))
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

func buildRule(recommendation findings.Recommendation) rule {
	return rule{
		ID:               recommendation.ID,
		ShortDescription: message{Text: recommendation.Recommendation},
		FullDescription:  message{Text: recommendation.LongDescription},
		HelpURI:           recommendation.LearnURL,
		Properties: map[string]string{
			"category":     recommendation.Category,
			"impact":       recommendation.Impact,
			"resourceType": recommendation.ResourceType,
			"source":       recommendation.Source,
		},
	}
}

func buildResult(recommendation findings.Recommendation, scopeID string) result {
	return result{
		RuleID: recommendation.ID,
		Level:  level(recommendation.Impact),
		Message: message{Text: fmt.Sprintf(
			"%s impacts %d Azure resource(s).",
			recommendation.Recommendation,
			recommendation.ImpactedResources,
		)},
		PartialFingerprints: map[string]string{
			"azqrRecommendationScope/v1": fingerprint(scopeID, recommendation.ID),
		},
		Locations: []location{{
			LogicalLocations: []logicalLocation{{
				Name:               recommendation.ID,
				FullyQualifiedName: recommendation.ResourceType + "/" + recommendation.ID,
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

func fingerprint(scopeID, recommendationID string) string {
	sum := sha256.Sum256([]byte(scopeID + "\x00" + strings.ToLower(recommendationID)))
	return hex.EncodeToString(sum[:16])
}
