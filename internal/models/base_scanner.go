// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

// BaseScanner provides common implementation for scanners with minimal or no recommendations
type BaseScanner struct {
	config        *ScannerConfig
	resourceTypes []string
}

// NewBaseScanner creates a new base scanner with specified resource types
// This returns a ready-to-use IAzureScanner implementation without needing custom wrapper structs
func NewBaseScanner(resourceTypes ...string) IAzureScanner {
	return &BaseScanner{
		resourceTypes: resourceTypes,
	}
}

// Init implements IAzureScanner.Init
func (b *BaseScanner) Init(config *ScannerConfig) error {
	b.config = config
	return nil
}

// Scan implements IAzureScanner.Scan - returns empty results
func (b *BaseScanner) Scan(scanContext *ScanContext) ([]*AzqrServiceResult, error) {
	LogSubscriptionScan(b.config.SubscriptionID, b.resourceTypes[0])
	return []*AzqrServiceResult{}, nil
}

// ResourceTypes implements IAzureScanner.ResourceTypes
func (b *BaseScanner) ResourceTypes() []string {
	return b.resourceTypes
}

// GetRecommendations implements IAzureScanner.GetRecommendations - returns empty map
func (b *BaseScanner) GetRecommendations() map[string]AzqrRecommendation {
	return map[string]AzqrRecommendation{}
}

// GetConfig returns the scanner configuration
func (b *BaseScanner) GetConfig() *ScannerConfig {
	return b.config
}
