// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package notifications

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/history"
)

func TestBuildIncludesSameScopeChanges(t *testing.T) {
	summary := &findings.Summary{
		Resources:         4,
		ImpactedResources: 3,
		ImpactedByImpact:  map[string]int{"High": 3},
		Recommendations: []findings.Recommendation{
			{ID: "rec", Impact: "High", ImpactedResources: 3},
		},
	}
	previous := &history.Record{Recommendations: []history.RecommendationCount{{ID: "rec", Count: 1}}}
	model, err := Build(summary, previous, "scope", "failed", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !model.BaselineFound || len(model.Changes) != 1 || model.Changes[0].Delta != 2 {
		t.Fatalf("unexpected model: %+v", model)
	}
}

func TestSendSlackPayload(t *testing.T) {
	var body string
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data, _ := io.ReadAll(request.Body)
		body = string(data)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := Send(context.Background(), server.Client(), ProviderSlack, server.URL, Model{ScopeID: "scope", ByImpact: map[string]int{}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"blocks"`) || !strings.Contains(body, "scope") {
		t.Fatalf("unexpected Slack payload: %s", body)
	}
}

func TestSendRejectsAndRedactsURL(t *testing.T) {
	secretURL := "http://example.com/path?secret=value"
	err := Send(context.Background(), nil, ProviderSlack, secretURL, Model{})
	if err == nil {
		t.Fatal("expected URL validation error")
	}
	if strings.Contains(err.Error(), "secret=value") {
		t.Fatalf("error leaked webhook URL: %v", err)
	}
}

func TestSendReturnsNonSuccessStatus(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	err := Send(context.Background(), server.Client(), ProviderTeams, server.URL, Model{ByImpact: map[string]int{}})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("unexpected error: %v", err)
	}
}
