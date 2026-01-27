// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// mockCredential is a mock implementation of azcore.TokenCredential for testing
type mockCredential struct {
	token string
	err   error
}

func (m *mockCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if m.err != nil {
		return azcore.AccessToken{}, m.err
	}
	return azcore.AccessToken{
		Token:     m.token,
		ExpiresOn: time.Now().Add(1 * time.Hour),
	}, nil
}

// testHttpClientOptions returns options optimized for fast test execution
func testHttpClientOptions() *HttpClientOptions {
	return &HttpClientOptions{
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		OperationTimeout: 10 * time.Second, // Enough for test operations
		Scope:            "https://management.azure.com/.default",
	}
}

func TestDo_Success(t *testing.T) {
	// Create a test server (TLS required for Azure SDK auth policy)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	// Create client with test server's HTTP client (trusts self-signed cert)
	cred := &mockCredential{token: "test-token"}
	opts := testHttpClientOptions()
	opts.Transport = server.Client()
	client := NewHttpClient(cred, opts)

	// Execute request
	body, err := client.Do(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := `{"result": "success"}`
	if string(body) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestDo_WithAuth(t *testing.T) {
	// Create a test server that checks for auth header (TLS required)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": "authenticated"}`))
	}))
	defer server.Close()

	// Create client with test server's HTTP client (trusts self-signed cert)
	cred := &mockCredential{token: "test-token"}
	opts := testHttpClientOptions()
	opts.Transport = server.Client()
	client := NewHttpClient(cred, opts)

	// Execute request with auth (empty string uses default scope)
	body, err := client.Do(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := `{"result": "authenticated"}`
	if string(body) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestDo_HTTPError(t *testing.T) {
	// Create a test server that returns an error (TLS required)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	// Create client with fast retry options and test server's client
	cred := &mockCredential{token: "test-token"}
	opts := testHttpClientOptions()
	opts.Transport = server.Client()
	client := NewHttpClient(cred, opts)

	// Execute request
	_, err := client.Do(context.Background(), server.URL)

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, httpErr.StatusCode)
	}
}

func TestNewClient(t *testing.T) {
	cred := &mockCredential{token: "test-token"}
	client := NewHttpClient(cred, testHttpClientOptions())

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
		return
	}
}

func TestNewClient_NilOptions(t *testing.T) {
	cred := &mockCredential{token: "test-token"}
	client := NewHttpClient(cred, nil)

	if client == nil {
		t.Fatal("Expected client to be created with default options, got nil")
	}
}

func TestDefaultHttpClientOptions(t *testing.T) {
	opts := DefaultHttpClientOptions(45 * time.Second)

	if opts.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", opts.Timeout)
	}
	if opts.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", opts.MaxRetries)
	}
	if opts.OperationTimeout != 450*time.Second {
		t.Errorf("Expected OperationTimeout 450s, got %v", opts.OperationTimeout)
	}
	if opts.Scope != "https://management.azure.com/.default" {
		t.Errorf("Expected default ARM scope, got %s", opts.Scope)
	}
}

func TestThrottlingPolicy(t *testing.T) {
	policy := throttling.NewThrottlingPolicy()
	if policy == nil {
		t.Fatal("Expected throttling policy to be created, got nil")
		return
	}
}
