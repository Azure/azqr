// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package metrics

import (
	"reflect"
	"strings"
	"testing"
)

func defs(entries map[string][]string) []Definition {
	var out []Definition
	for name, dims := range entries {
		out = append(out, Definition{Name: name, Dimensions: dims})
	}
	return out
}

func TestResolveMetricPlan(t *testing.T) {
	deploy := []string{"ModelDeploymentName", "ModelName"}
	tests := []struct {
		name      string
		defs      []Definition
		want      Plan
		wantNotes string
	}{
		{
			name: "primary azure openai metrics",
			defs: defs(map[string][]string{
				"InputTokens": append([]string{"ServiceTierRequest", "ContextLength"}, deploy...), "OutputTokens": append([]string{"ServiceTierRequest", "ContextLength"}, deploy...),
				"AzureOpenAIRequests": append([]string{"StatusCode"}, deploy...), "AzureOpenAITimeToResponse": deploy,
			}),
			want: Plan{Input: "InputTokens", Output: "OutputTokens", Requests: "AzureOpenAIRequests", Latency: "AzureOpenAITimeToResponse", StatusSplit: true, TierDim: true, ContextDim: true},
		},
		{
			name: "fallback names and ModelRequests priority",
			defs: defs(map[string][]string{
				"ProcessedPromptTokens": deploy, "GeneratedTokens": deploy,
				"ModelRequests": deploy, "AzureOpenAIRequests": append([]string{"StatusCode"}, deploy...),
				"ThrottledCalls": deploy,
			}),
			want:      Plan{Input: "ProcessedPromptTokens", Output: "GeneratedTokens", Requests: "ModelRequests", Throttled: "ThrottledCalls"},
			wantNotes: "latency: no supported metric",
		},
		{
			name: "metric not filterable by deployment is excluded",
			defs: defs(map[string][]string{
				"InputTokens": deploy, "OutputTokens": deploy,
				"ModelRequests": {"Region"}, "AzureOpenAIRequests": deploy,
			}),
			want:      Plan{Input: "InputTokens", Output: "OutputTokens", Requests: "AzureOpenAIRequests"},
			wantNotes: "ModelRequests skipped",
		},
		{
			name:      "missing everything",
			defs:      nil,
			want:      Plan{},
			wantNotes: "requests: no supported metric",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePlan(tt.defs)
			notes := strings.Join(got.Notes, ";")
			got.Notes = nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("plan = %+v, want %+v", got, tt.want)
			}
			if tt.wantNotes != "" && !strings.Contains(notes, tt.wantNotes) {
				t.Errorf("notes %q missing %q", notes, tt.wantNotes)
			}
		})
	}
}

func TestClassifyStatus(t *testing.T) {
	tests := map[string]string{
		"200": statusSuccess, "204": statusSuccess, "429": statusRateLimited,
		"400": statusFailed, "500": statusFailed, "abc": statusFailed,
		"": StatusUnknown, "Unknown": StatusUnknown,
	}
	for in, want := range tests {
		if got := ClassifyStatus(in); got != want {
			t.Errorf("classifyStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestContextBandAndTier(t *testing.T) {
	if contextBand("1000") != "short" || contextBand("128000") != "long" || contextBand("") != StatusUnknown || contextBand("x") != StatusUnknown {
		t.Error("unexpected context band")
	}
	if normalizeTier(" Standard ") != "standard" || normalizeTier("") != StatusUnknown {
		t.Error("unexpected tier normalisation")
	}
}

func TestResolvePlanPrefersTierCapableTokens(t *testing.T) {
	deploy := []string{"ModelDeploymentName", "ModelName"}
	plan := ResolvePlan(defs(map[string][]string{
		"InputTokens": deploy, "OutputTokens": deploy,
		"ProcessedPromptTokens": append([]string{"ServiceTierRequest"}, deploy...),
		"GeneratedTokens":       append([]string{"ServiceTierRequest"}, deploy...),
		"AzureOpenAIRequests":   deploy,
	}))
	if plan.Input != "ProcessedPromptTokens" || plan.Output != "GeneratedTokens" || !plan.TierDim {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
