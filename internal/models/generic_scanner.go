// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"context"
)

// GenericScanner provides a reusable scanner implementation for Azure services
// TResource: The ARM resource type (e.g., *armcontainerservice.ManagedCluster)
// TClient: The ARM client type (e.g., *armcontainerservice.ManagedClustersClient)
type GenericScanner[TResource any, TClient any] struct {
	config              *ScannerConfig
	client              TClient
	resourceTypes       []string
	clientFactory       func(*ScannerConfig) (TClient, error)
	listResources       func(TClient, context.Context) ([]*TResource, error)
	getRecommendations  func() map[string]AzqrRecommendation
	extractResourceInfo func(*TResource) ResourceInfo
}

// GenericScannerConfig holds the configuration for creating a generic scanner
type GenericScannerConfig[TResource any, TClient any] struct {
	// ResourceTypes list of Azure resource types this scanner handles
	ResourceTypes []string
	// ClientFactory creates the ARM client from scanner config
	ClientFactory func(*ScannerConfig) (TClient, error)
	// ListResources retrieves all resources using the client
	ListResources func(TClient, context.Context) ([]*TResource, error)
	// GetRecommendations returns the recommendation rules for this service
	GetRecommendations func() map[string]AzqrRecommendation
	// ExtractResourceInfo extracts common fields from a resource for the service result
	ExtractResourceInfo func(*TResource) ResourceInfo
}

// ResourceInfo holds common resource information needed for reporting
type ResourceInfo struct {
	ID       string
	Name     string
	Location string
	Type     string
}

// ExtractStandardARMResourceInfo is a helper for resources with standard ARM structure
func ExtractStandardARMResourceInfo(id, name, location, resourceType *string) ResourceInfo {
	return ResourceInfo{
		ID:       derefString(id),
		Name:     derefString(name),
		Location: derefString(location),
		Type:     derefString(resourceType),
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// NewGenericScanner creates a new generic scanner with the provided configuration
func NewGenericScanner[TResource any, TClient any](
	config GenericScannerConfig[TResource, TClient],
) *GenericScanner[TResource, TClient] {
	return &GenericScanner[TResource, TClient]{
		resourceTypes:       config.ResourceTypes,
		clientFactory:       config.ClientFactory,
		listResources:       config.ListResources,
		getRecommendations:  config.GetRecommendations,
		extractResourceInfo: config.ExtractResourceInfo,
	}
}

// Init initializes the generic scanner with configuration and creates the ARM client
func (s *GenericScanner[TResource, TClient]) Init(config *ScannerConfig) error {
	s.config = config

	client, err := s.clientFactory(config)
	if err != nil {
		return err
	}
	s.client = client

	return nil
}

// Scan executes the scan for all resources of this type
func (s *GenericScanner[TResource, TClient]) Scan(scanContext *ScanContext) ([]*AzqrServiceResult, error) {
	LogSubscriptionScan(s.config.SubscriptionID, s.resourceTypes[0])

	// List all resources
	resources, err := s.listResources(s.client, s.config.Ctx)
	if err != nil {
		return nil, err
	}

	// Get recommendation rules
	rules := s.getRecommendations()
	engine := RecommendationEngine{}
	results := []*AzqrServiceResult{}

	// Evaluate recommendations for each resource
	for _, resource := range resources {
		rr := engine.EvaluateRecommendations(rules, resource, scanContext)

		// Extract common fields from resource using type assertion or reflection
		serviceResult := s.buildServiceResult(resource, rr)
		results = append(results, serviceResult)
	}

	return results, nil
}

// buildServiceResult creates an AzqrServiceResult from a resource
func (s *GenericScanner[TResource, TClient]) buildServiceResult(
	resource *TResource,
	recommendations map[string]*AzqrResult,
) *AzqrServiceResult {
	// Extract common fields using the provided extractor function
	info := s.extractResourceInfo(resource)

	return &AzqrServiceResult{
		SubscriptionID:   s.config.SubscriptionID,
		SubscriptionName: s.config.SubscriptionName,
		ResourceGroup:    GetResourceGroupFromResourceID(info.ID),
		Location:         info.Location,
		Type:             info.Type,
		ServiceName:      info.Name,
		Recommendations:  recommendations,
	}
}

// ResourceTypes returns the list of resource types this scanner handles
func (s *GenericScanner[TResource, TClient]) ResourceTypes() []string {
	return s.resourceTypes
}

// GetRecommendations returns the recommendation rules for this scanner
func (s *GenericScanner[TResource, TClient]) GetRecommendations() map[string]AzqrRecommendation {
	return s.getRecommendations()
}
