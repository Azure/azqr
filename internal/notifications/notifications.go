// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package notifications builds and sends scan summaries to supported webhooks.
package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/history"
	"github.com/Azure/azqr/internal/models"
)

const maxTopItems = 20

// Provider identifies a webhook payload contract.
type Provider string

const (
	// ProviderSlack sends Slack Block Kit payloads.
	ProviderSlack Provider = "slack"
	// ProviderTeams sends Teams Workflows Adaptive Card payloads.
	ProviderTeams Provider = "teams"
)

// HTTPClient is the subset of http.Client used by webhook delivery.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Item is one ranked recommendation.
type Item struct {
	ID     string `json:"id"`
	Impact string `json:"impact"`
	Count  int    `json:"count"`
	Delta  int    `json:"delta,omitempty"`
}

// Model is the provider-neutral notification summary.
type Model struct {
	ScopeID       string         `json:"scopeId"`
	Resources     int            `json:"resources"`
	Findings      int            `json:"findings"`
	ByImpact      map[string]int `json:"byImpact"`
	GateStatus    string         `json:"gateStatus,omitempty"`
	Top           []Item         `json:"top"`
	Changes       []Item         `json:"changes"`
	BaselineFound bool           `json:"baselineFound"`
}

// ParseProvider validates a provider CLI value.
func ParseProvider(value string) (Provider, error) {
	switch Provider(strings.ToLower(strings.TrimSpace(value))) {
	case ProviderSlack:
		return ProviderSlack, nil
	case ProviderTeams:
		return ProviderTeams, nil
	default:
		return "", fmt.Errorf("invalid notify provider %q: supported values are teams and slack", value)
	}
}

// Build creates a provider-neutral notification model.
func Build(summary *findings.Summary, previous *history.Record, scopeID, gateStatus string, top int) (Model, error) {
	if top < 1 || top > maxTopItems {
		return Model{}, fmt.Errorf("notify-top must be between 1 and %d", maxTopItems)
	}
	model := Model{
		ScopeID:    scopeID,
		ByImpact:   map[string]int{},
		GateStatus: gateStatus,
		Top:        []Item{},
		Changes:    []Item{},
	}
	if summary == nil {
		return model, nil
	}
	model.Resources = summary.Resources
	model.Findings = summary.ImpactedResources
	for key, value := range summary.ImpactedByImpact {
		model.ByImpact[key] = value
	}

	current := make(map[string]Item)
	for _, recommendation := range summary.Recommendations {
		if recommendation.ImpactedResources == 0 {
			continue
		}
		item := Item{
			ID:     recommendation.ID,
			Impact: recommendation.Impact,
			Count:  recommendation.ImpactedResources,
		}
		current[strings.ToLower(recommendation.ID)] = item
		model.Top = append(model.Top, item)
	}
	sort.Slice(model.Top, func(i, j int) bool {
		if model.Top[i].Count != model.Top[j].Count {
			return model.Top[i].Count > model.Top[j].Count
		}
		return model.Top[i].ID < model.Top[j].ID
	})
	model.Top = limit(model.Top, top)

	if previous == nil {
		return model, nil
	}
	model.BaselineFound = true
	previousCounts := make(map[string]history.RecommendationCount, len(previous.Recommendations))
	for _, recommendation := range previous.Recommendations {
		previousCounts[strings.ToLower(recommendation.ID)] = recommendation
	}

	for key, item := range current {
		item.Delta = item.Count - previousCounts[key].Count
		if item.Delta != 0 {
			model.Changes = append(model.Changes, item)
		}
		delete(previousCounts, key)
	}
	for _, recommendation := range previousCounts {
		model.Changes = append(model.Changes, Item{
			ID:     recommendation.ID,
			Impact: recommendation.Impact,
			Delta:  -recommendation.Count,
		})
	}
	sort.Slice(model.Changes, func(i, j int) bool {
		left := abs(model.Changes[i].Delta)
		right := abs(model.Changes[j].Delta)
		if left != right {
			return left > right
		}
		return model.Changes[i].ID < model.Changes[j].ID
	})
	model.Changes = limit(model.Changes, top)
	return model, nil
}

// Send posts a provider-specific webhook payload.
func Send(ctx context.Context, client HTTPClient, provider Provider, webhookURL string, model Model) error {
	parsed, err := url.Parse(webhookURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("notify-webhook must be a valid HTTPS URL")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	payload, err := render(provider, model)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create %s notification request: invalid request", provider)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send %s notification to %s: request failed", provider, parsed.Hostname())
	}
	if _, err := io.Copy(io.Discard, io.LimitReader(response.Body, 4096)); err != nil {
		_ = response.Body.Close()
		return fmt.Errorf("read %s notification response from %s", provider, parsed.Hostname())
	}
	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("close %s notification response from %s", provider, parsed.Hostname())
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("send %s notification to %s: unexpected HTTP status %d", provider, parsed.Hostname(), response.StatusCode)
	}
	return nil
}

func render(provider Provider, model Model) ([]byte, error) {
	text := summaryText(model)
	switch provider {
	case ProviderSlack:
		return json.Marshal(map[string]any{
			"text": text,
			"blocks": []map[string]any{{
				"type": "section",
				"text": map[string]string{"type": "mrkdwn", "text": text},
			}},
		})
	case ProviderTeams:
		return json.Marshal(map[string]any{
			"type": "message",
			"attachments": []map[string]any{{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]any{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body": []map[string]any{{
						"type": "TextBlock",
						"text": text,
						"wrap": true,
					}},
				},
			}},
		})
	default:
		return nil, fmt.Errorf("unsupported notification provider %q", provider)
	}
}

func summaryText(model Model) string {
	var text strings.Builder
	fmt.Fprintf(
		&text,
		"*azqr scan* | scope `%s`\nResources: %d | Findings: %d | High: %d | Medium: %d | Low: %d",
		model.ScopeID,
		model.Resources,
		model.Findings,
		model.ByImpact[string(models.ImpactHigh)],
		model.ByImpact[string(models.ImpactMedium)],
		model.ByImpact[string(models.ImpactLow)],
	)
	if model.GateStatus != "" {
		fmt.Fprintf(&text, "\nGate: %s", truncate(model.GateStatus, 160))
	}
	if len(model.Top) > 0 {
		text.WriteString("\nTop recommendations:")
		for _, item := range model.Top {
			fmt.Fprintf(&text, "\n- %s (%s): %d", truncate(item.ID, 100), item.Impact, item.Count)
		}
	}
	if !model.BaselineFound {
		text.WriteString("\nChanges: no same-scope history baseline")
	} else if len(model.Changes) == 0 {
		text.WriteString("\nChanges: none")
	} else {
		text.WriteString("\nLargest changes:")
		for _, item := range model.Changes {
			fmt.Fprintf(&text, "\n- %s: %+d", truncate(item.ID, 100), item.Delta)
		}
	}
	return text.String()
}

func limit(items []Item, maximum int) []Item {
	if len(items) > maximum {
		return items[:maximum]
	}
	return items
}

func truncate(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	return value[:maximum-3] + "..."
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
