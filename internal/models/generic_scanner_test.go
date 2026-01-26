// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// Mock types for testing
type MockResource struct {
	ID       *string
	Name     *string
	Location *string
	Type     *string
}

type MockClient struct {
	resources []*MockResource
	err       error
}

func (m *MockClient) NewListPager() *MockPager {
	return &MockPager{
		resources: m.resources,
		err:       m.err,
	}
}

type MockPager struct {
	resources []*MockResource
	err       error
	called    bool
}

func (p *MockPager) More() bool {
	return !p.called
}

func (p *MockPager) NextPage(ctx context.Context) (*MockPageResponse, error) {
	p.called = true
	if p.err != nil {
		return nil, p.err
	}
	return &MockPageResponse{Value: p.resources}, nil
}

type MockPageResponse struct {
	Value []*MockResource
}

// Test successful scanner initialization
func TestGenericScanner_Init_Success(t *testing.T) {
	mockClient := &MockClient{}

	scanner := NewGenericScanner(
		GenericScannerConfig[MockResource, *MockClient]{
			ResourceTypes: []string{"Microsoft.Test/resources"},
			ClientFactory: func(config *ScannerConfig) (*MockClient, error) {
				return mockClient, nil
			},
			ListResources: func(client *MockClient, ctx context.Context) ([]*MockResource, error) {
				return nil, nil
			},
			GetRecommendations: func() map[string]AzqrRecommendation {
				return map[string]AzqrRecommendation{}
			},
			ExtractResourceInfo: func(r *MockResource) ResourceInfo {
				return ExtractStandardARMResourceInfo(r.ID, r.Name, r.Location, r.Type)
			},
		},
	)

	config := &ScannerConfig{
		Ctx:              context.Background(),
		SubscriptionID:   "00000000-0000-0000-0000-000000000000",
		SubscriptionName: "Test Subscription",
	}

	err := scanner.Init(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if scanner.config == nil {
		t.Error("Expected config to be set")
	}

	if scanner.client == nil {
		t.Error("Expected client to be initialized")
	}
}

// Test scanner initialization failure
func TestGenericScanner_Init_Failure(t *testing.T) {
	expectedError := errors.New("client creation failed")

	scanner := NewGenericScanner(
		GenericScannerConfig[MockResource, *MockClient]{
			ResourceTypes: []string{"Microsoft.Test/resources"},
			ClientFactory: func(config *ScannerConfig) (*MockClient, error) {
				return nil, expectedError
			},
			ListResources: func(client *MockClient, ctx context.Context) ([]*MockResource, error) {
				return nil, nil
			},
			GetRecommendations: func() map[string]AzqrRecommendation {
				return map[string]AzqrRecommendation{}
			},
			ExtractResourceInfo: func(r *MockResource) ResourceInfo {
				return ResourceInfo{}
			},
		},
	)

	config := &ScannerConfig{
		Ctx:              context.Background(),
		SubscriptionID:   "00000000-0000-0000-0000-000000000000",
		SubscriptionName: "Test Subscription",
	}

	err := scanner.Init(config)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

// Test successful scan with resources
func TestGenericScanner_Scan_Success(t *testing.T) {
	mockResources := []*MockResource{
		{
			ID:       to.Ptr("/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Test/resources/resource1"),
			Name:     to.Ptr("resource1"),
			Location: to.Ptr("eastus"),
			Type:     to.Ptr("Microsoft.Test/resources"),
		},
		{
			ID:       to.Ptr("/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Test/resources/resource2"),
			Name:     to.Ptr("resource2"),
			Location: to.Ptr("westus"),
			Type:     to.Ptr("Microsoft.Test/resources"),
		},
	}

	mockClient := &MockClient{resources: mockResources}

	scanner := NewGenericScanner(
		GenericScannerConfig[MockResource, *MockClient]{
			ResourceTypes: []string{"Microsoft.Test/resources"},
			ClientFactory: func(config *ScannerConfig) (*MockClient, error) {
				return mockClient, nil
			},
			ListResources: func(client *MockClient, ctx context.Context) ([]*MockResource, error) {
				pager := client.NewListPager()
				resources := make([]*MockResource, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					resources = append(resources, resp.Value...)
				}
				return resources, nil
			},
			GetRecommendations: func() map[string]AzqrRecommendation {
				return map[string]AzqrRecommendation{
					"test-001": {
						RecommendationID: "test-001",
						ResourceType:     "Microsoft.Test/resources",
						Category:         CategorySecurity,
						Impact:           ImpactHigh,
						Eval: func(target interface{}, ctx *ScanContext) (bool, string) {
							// Mock evaluation - always returns non-compliant
							return true, "Test recommendation"
						},
					},
				}
			},
			ExtractResourceInfo: func(r *MockResource) ResourceInfo {
				return ExtractStandardARMResourceInfo(r.ID, r.Name, r.Location, r.Type)
			},
		},
	)

	config := &ScannerConfig{
		Ctx:              context.Background(),
		SubscriptionID:   "00000000-0000-0000-0000-000000000000",
		SubscriptionName: "Test Subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	scanContext := &ScanContext{
		Filters:             NewFilters(),
		PrivateEndpoints:    map[string]bool{},
		DiagnosticsSettings: map[string]bool{},
		PublicIPs:           map[string]*armnetwork.PublicIPAddress{},
	}

	results, err := scanner.Scan(scanContext)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify first result
	if results[0].ServiceName != "resource1" {
		t.Errorf("Expected ServiceName 'resource1', got '%s'", results[0].ServiceName)
	}

	if results[0].Location != "eastus" {
		t.Errorf("Expected Location 'eastus', got '%s'", results[0].Location)
	}

	if results[0].SubscriptionID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("Expected SubscriptionID '00000000-0000-0000-0000-000000000000', got '%s'", results[0].SubscriptionID)
	}

	// Verify recommendations were evaluated
	if len(results[0].Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(results[0].Recommendations))
	}
}

// Test scan with list resources error
func TestGenericScanner_Scan_ListResourcesError(t *testing.T) {
	expectedError := errors.New("list resources failed")
	mockClient := &MockClient{err: expectedError}

	scanner := NewGenericScanner(
		GenericScannerConfig[MockResource, *MockClient]{
			ResourceTypes: []string{"Microsoft.Test/resources"},
			ClientFactory: func(config *ScannerConfig) (*MockClient, error) {
				return mockClient, nil
			},
			ListResources: func(client *MockClient, ctx context.Context) ([]*MockResource, error) {
				pager := client.NewListPager()
				resources := make([]*MockResource, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					resources = append(resources, resp.Value...)
				}
				return resources, nil
			},
			GetRecommendations: func() map[string]AzqrRecommendation {
				return map[string]AzqrRecommendation{}
			},
			ExtractResourceInfo: func(r *MockResource) ResourceInfo {
				return ResourceInfo{}
			},
		},
	)

	config := &ScannerConfig{
		Ctx:              context.Background(),
		SubscriptionID:   "00000000-0000-0000-0000-000000000000",
		SubscriptionName: "Test Subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	scanContext := &ScanContext{
		Filters:             NewFilters(),
		PrivateEndpoints:    map[string]bool{},
		DiagnosticsSettings: map[string]bool{},
	}

	_, err = scanner.Scan(scanContext)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

// Test ResourceTypes method
func TestGenericScanner_ResourceTypes(t *testing.T) {
	expectedTypes := []string{"Microsoft.Test/resources", "Microsoft.Test/otherResources"}

	scanner := NewGenericScanner(
		GenericScannerConfig[MockResource, *MockClient]{
			ResourceTypes: expectedTypes,
			ClientFactory: func(config *ScannerConfig) (*MockClient, error) {
				return &MockClient{}, nil
			},
			ListResources: func(client *MockClient, ctx context.Context) ([]*MockResource, error) {
				return nil, nil
			},
			GetRecommendations: func() map[string]AzqrRecommendation {
				return map[string]AzqrRecommendation{}
			},
			ExtractResourceInfo: func(r *MockResource) ResourceInfo {
				return ResourceInfo{}
			},
		},
	)

	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) != len(expectedTypes) {
		t.Errorf("Expected %d resource types, got %d", len(expectedTypes), len(resourceTypes))
	}

	for i, rt := range resourceTypes {
		if rt != expectedTypes[i] {
			t.Errorf("Expected resource type '%s', got '%s'", expectedTypes[i], rt)
		}
	}
}

// Test GetRecommendations method
func TestGenericScanner_GetRecommendations(t *testing.T) {
	expectedRecommendations := map[string]AzqrRecommendation{
		"test-001": {
			RecommendationID: "test-001",
			ResourceType:     "Microsoft.Test/resources",
		},
		"test-002": {
			RecommendationID: "test-002",
			ResourceType:     "Microsoft.Test/resources",
		},
	}

	scanner := NewGenericScanner(
		GenericScannerConfig[MockResource, *MockClient]{
			ResourceTypes: []string{"Microsoft.Test/resources"},
			ClientFactory: func(config *ScannerConfig) (*MockClient, error) {
				return &MockClient{}, nil
			},
			ListResources: func(client *MockClient, ctx context.Context) ([]*MockResource, error) {
				return nil, nil
			},
			GetRecommendations: func() map[string]AzqrRecommendation {
				return expectedRecommendations
			},
			ExtractResourceInfo: func(r *MockResource) ResourceInfo {
				return ResourceInfo{}
			},
		},
	)

	recommendations := scanner.GetRecommendations()

	if len(recommendations) != len(expectedRecommendations) {
		t.Errorf("Expected %d recommendations, got %d", len(expectedRecommendations), len(recommendations))
	}

	for id := range expectedRecommendations {
		if _, exists := recommendations[id]; !exists {
			t.Errorf("Expected recommendation '%s' to exist", id)
		}
	}
}

// Test ExtractStandardARMResourceInfo helper
func TestExtractStandardARMResourceInfo(t *testing.T) {
	id := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Test/resources/resource1"
	name := "resource1"
	location := "eastus"
	resourceType := "Microsoft.Test/resources"

	info := ExtractStandardARMResourceInfo(
		to.Ptr(id),
		to.Ptr(name),
		to.Ptr(location),
		to.Ptr(resourceType),
	)

	if info.ID != id {
		t.Errorf("Expected ID '%s', got '%s'", id, info.ID)
	}

	if info.Name != name {
		t.Errorf("Expected Name '%s', got '%s'", name, info.Name)
	}

	if info.Location != location {
		t.Errorf("Expected Location '%s', got '%s'", location, info.Location)
	}

	if info.Type != resourceType {
		t.Errorf("Expected Type '%s', got '%s'", resourceType, info.Type)
	}
}

// Test ExtractStandardARMResourceInfo with nil values
func TestExtractStandardARMResourceInfo_NilValues(t *testing.T) {
	info := ExtractStandardARMResourceInfo(nil, nil, nil, nil)

	if info.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", info.ID)
	}

	if info.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", info.Name)
	}

	if info.Location != "" {
		t.Errorf("Expected empty Location, got '%s'", info.Location)
	}

	if info.Type != "" {
		t.Errorf("Expected empty Type, got '%s'", info.Type)
	}
}
