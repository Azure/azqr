// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type testCredential struct{}

func (testCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     "test-token",
		ExpiresOn: time.Now().Add(time.Hour),
	}, nil
}

type rewriteTransport struct {
	client *http.Client
	target *url.URL
}

func (r *rewriteTransport) Do(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = r.target.Scheme
	clone.URL.Host = r.target.Host
	clone.Host = r.target.Host
	return r.client.Do(clone)
}

func newTestPricingClient(t *testing.T, handler http.HandlerFunc) *az.HttpClient {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(server.URL) error = %v", err)
	}

	return az.NewHttpClient(testCredential{}, &az.HttpClientOptions{
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		OperationTimeout: 10 * time.Second,
		Scope:            "https://management.azure.com/.default",
		Transport: &rewriteTransport{
			client: server.Client(),
			target: target,
		},
	})
}

func TestResolveAzureOpenAIPrice_ExactMeterMatch(t *testing.T) {
	client := newTestPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.03, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.06, "unitOfMeasure": "1M", "meterName": "gpt-4o Outp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 999, "unitOfMeasure": "1M", "meterName": "gpt-35-turbo Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4o", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to find an exact meter match")
	}
	if price.Source != "exact_retail" {
		t.Fatalf("expected Source %q, got %q", "exact_retail", price.Source)
	}
	if price.InputPerMTokUSD != 0.03 {
		t.Fatalf("expected InputPerMTokUSD 0.03, got %v", price.InputPerMTokUSD)
	}
	if price.OutputPerMTokUSD != 0.06 {
		t.Fatalf("expected OutputPerMTokUSD 0.06, got %v", price.OutputPerMTokUSD)
	}
}

func TestResolveAzureOpenAIPrice_NoMatchIsUnresolved(t *testing.T) {
	client := newTestPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Items": [], "NextPageLink": null}`))
	})

	_, ok := resolveAzureOpenAIPrice(context.Background(), client, "some-unlisted-model", "eastus")
	if ok {
		t.Fatal("expected resolveAzureOpenAIPrice to report no match when the API returns no items")
	}
}

func TestResolveAzureOpenAIPrice_PartialMeterPairIsUnresolved(t *testing.T) {
	client := newTestPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Only the input meter is present; output is missing, so the pair is incomplete.
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.03, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	_, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4o", "eastus")
	if ok {
		t.Fatal("expected resolveAzureOpenAIPrice to report no match with only one side of the meter pair")
	}
}

func TestResolveAzureOpenAIPrice_MessyRealWorldMeterNames(t *testing.T) {
	// Real Azure Retail Prices meter names for "Foundry Models" are
	// inconsistently formatted and encode several billing axes at once.
	// This fixture mirrors meters observed live for gpt-6-sol: dropped
	// "gpt" prefix, short/long context, cached vs. standard, priority
	// processing vs. standard, and global vs. data-zone tier.
	client := newTestPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.4,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Cd Inp PP Gl 1M Tokens",  "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 2.0,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Inp Std Gl 1M Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 2.2,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Inp Std DZ 1M Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 4.4,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Inp PP Gl 1M Tokens",    "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 10.0, "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Opt Std Gl 1M Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 4.0,  "unitOfMeasure": "1M", "meterName": "6-sol LongCo Inp Std Gl 1M Tokens",    "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 15.0, "unitOfMeasure": "1M", "meterName": "6-sol LongCo Opt Std Gl 1M Tokens",    "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-6-sol", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to resolve gpt-6-sol despite the messy meter names")
	}
	if price.Source != "exact_retail" {
		t.Fatalf("expected Source %q, got %q", "exact_retail", price.Source)
	}
	// Should prefer the Global, Standard (non-cached, non-priority-processing),
	// short-context pair: 2.0 input / 10.0 output.
	if price.InputPerMTokUSD != 2.0 {
		t.Fatalf("expected InputPerMTokUSD 2.0 (Std/Gl/ShortCo, not cached/PP/DZ), got %v", price.InputPerMTokUSD)
	}
	if price.OutputPerMTokUSD != 10.0 {
		t.Fatalf("expected OutputPerMTokUSD 10.0, got %v", price.OutputPerMTokUSD)
	}
}

func TestResolveAzureOpenAIPrice_EmbeddingModelHasNoOutputMeter(t *testing.T) {
	// Embedding models only ever bill on input tokens; there is no
	// "Outp"/"Opt" meter to find, so the output rate should resolve to
	// zero rather than being reported unavailable.
	client := newTestPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.00013,  "unitOfMeasure": "1K", "meterName": "text-embedding-3-large-glbl Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.000143, "unitOfMeasure": "1K", "meterName": "text embedding 3 large DZ Tokens",     "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.00013,  "unitOfMeasure": "1K", "meterName": "text-embedding-3-large-grader Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "text-embedding-3-large", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to resolve the embedding model from its single input meter")
	}
	if price.OutputPerMTokUSD != 0 {
		t.Fatalf("expected OutputPerMTokUSD 0 for an embedding model, got %v", price.OutputPerMTokUSD)
	}
	// unitOfMeasure "1K": retailPrice 0.00013/1K tokens -> 0.13/1M tokens; prefers Global tier.
	if got, want := price.InputPerMTokUSD, 0.13; got < want-1e-9 || got > want+1e-9 {
		t.Fatalf("expected InputPerMTokUSD %.5f (normalized from 1K to 1M and preferring Global tier), got %.5f", want, got)
	}
}

func TestResolveAzureOpenAIPrice_ExcludesBatchAndPTUMeters(t *testing.T) {
	// Only batch/provisioned meters exist for the model; since standard
	// pay-as-you-go pricing can't be found, the result must stay
	// unresolved rather than silently reporting a batch/PTU rate.
	client := newTestPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.015, "unitOfMeasure": "1M", "meterName": "gpt-4-8K-Batch-Inp-glbl Tokens",  "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.03,  "unitOfMeasure": "1M", "meterName": "gpt-4-8K-Batch-Outp-glbl Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 5.0,   "unitOfMeasure": "1M", "meterName": "gpt-4-8K PTU Inp Gl Tokens",      "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	_, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4-8K", "eastus")
	if ok {
		t.Fatal("expected batch/PTU-only meters to leave the price unresolved")
	}
}

func TestResolveAzureOpenAIPrice_FollowsPagination(t *testing.T) {
	mux := http.NewServeMux()
	var page2URL string
	mux.HandleFunc("/api/retail/prices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.03, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": "` + page2URL + `"
		}`))
	})
	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.06, "unitOfMeasure": "1M", "meterName": "gpt-4o Outp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	page2URL = server.URL + "/page2"

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(server.URL) error = %v", err)
	}
	client := az.NewHttpClient(testCredential{}, &az.HttpClientOptions{
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		OperationTimeout: 10 * time.Second,
		Scope:            "https://management.azure.com/.default",
		Transport: &rewriteTransport{
			client: server.Client(),
			target: target,
		},
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4o", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to combine meters found across pages")
	}
	if price.InputPerMTokUSD != 0.03 || price.OutputPerMTokUSD != 0.06 {
		t.Fatalf("expected combined price 0.03/0.06, got %v/%v", price.InputPerMTokUSD, price.OutputPerMTokUSD)
	}
}
