// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

// BaseScanner provides common implementation for scanners
type BaseScanner struct {
	serviceName   string
	resourceTypes []string
}

// NewBaseScanner creates a new base scanner
// This returns a ready-to-use IAzureScanner implementation without needing custom wrapper structs
func NewBaseScanner(serviceName string, resourceTypes ...string) IAzureScanner {
	return &BaseScanner{
		serviceName:   serviceName,
		resourceTypes: resourceTypes,
	}
}

// ServiceName implements IAzureScanner.ServiceName
func (b *BaseScanner) ServiceName() string {
	return b.serviceName
}

// ResourceTypes implements IAzureScanner.ResourceTypes
func (b *BaseScanner) ResourceTypes() []string {
	return b.resourceTypes
}
