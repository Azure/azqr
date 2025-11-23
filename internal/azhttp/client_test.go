// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestDo_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	// Create client
	cred := &mockCredential{token: "test-token"}
	client := NewClient(cred, 10*time.Second)

	// Execute request
	body, err := client.Do(context.Background(), server.URL, false, 3)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := `{"result": "success"}`
	if string(body) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestDo_WithAuth(t *testing.T) {
	// Create a test server that checks for auth header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": "authenticated"}`))
	}))
	defer server.Close()

	// Create client
	cred := &mockCredential{token: "test-token"}
	client := NewClient(cred, 10*time.Second)

	// Execute request with auth
	body, err := client.Do(context.Background(), server.URL, true, 3)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := `{"result": "authenticated"}`
	if string(body) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestDo_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	// Create client
	cred := &mockCredential{token: "test-token"}
	client := NewClient(cred, 10*time.Second)

	// Execute request
	_, err := client.Do(context.Background(), server.URL, false, 1)

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
	client := NewClient(cred, 30*time.Second)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.cred == nil {
		t.Fatal("Expected credential to be set, got nil")
	}
}

func TestThrottlingPolicy(t *testing.T) {
	policy := newThrottlingPolicy()
	if policy == nil {
		t.Fatal("Expected throttling policy to be created, got nil")
	}
}
