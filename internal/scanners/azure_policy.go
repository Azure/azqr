// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// AzurePolicyScanner scans for non-compliant resources based on Azure Policy states.
type AzurePolicyScanner struct{}

// Scan queries Azure Resource Graph for non-compliant policy states across the specified subscriptions
func (s *AzurePolicyScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.AzurePolicyResult {
	models.LogResourceTypeScan("Azure Policy (Non Compliant Resources)")

	graphClient := graph.NewGraphQuery(cred)
	query := `
		PolicyResources
		| where type == 'microsoft.policyinsights/policystates'
		| extend 
			resourceId = tostring(properties.resourceId),
			subscriptionId = tostring(properties.subscriptionId),
			policyAssignmentId = tostring(properties.policyAssignmentId),
			policyAssignmentName = tostring(properties.policyAssignmentName),
			policyDefinitionId = tostring(properties.policyDefinitionId),
			policyDefinitionName = tostring(properties.policyDefinitionName),
			timestamp = todatetime(properties.timestamp),
			complianceState = tostring(properties.complianceState)
		| where complianceState == 'NonCompliant'
		| join kind=leftouter (
			PolicyResources
			| where type == 'microsoft.authorization/policydefinitions'
			| extend policyDefinitionId = tolower(id)
			| project policyDefinitionId, policyDescription = tostring(properties.description), policyDefinitionDisplayName = properties.displayName
		) on policyDefinitionId
		| join kind=leftouter (
			ResourceContainers
			| where type == 'microsoft.resources/subscriptions'
			| project subscriptionId = tolower(subscriptionId), subscriptionName = name
		) on subscriptionId
		| project subscriptionId, subscriptionName, resourceId, policyAssignmentId, policyAssignmentName, policyDefinitionId, policyDefinitionName, timestamp, policyDefinitionDisplayName, policyDescription, complianceState
		`

	log.Debug().Msg(query)
	result, err := graphClient.Query(ctx, query, subscriptions, graph.QueryOptions{ManagementGroupScope: true})
	// Composite key type for deduplication - avoids string concatenation allocations
	type policyKey struct {
		resourceID         string
		policyDefinitionID string
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for Azure Policy non-compliant resources")
		return nil
	}
	resources := []*models.AzurePolicyResult{}
	seen := make(map[policyKey]struct{}, len(result.Data))

	type policyRow struct {
		SubscriptionID          string `json:"subscriptionId"`
		SubscriptionName        string `json:"subscriptionName"`
		ResourceID              string `json:"resourceId"`
		PolicyDefinitionDisplay string `json:"policyDefinitionDisplayName"`
		PolicyDescription       string `json:"policyDescription"`
		Timestamp               string `json:"timestamp"`
		PolicyDefinitionName    string `json:"policyDefinitionName"`
		PolicyDefinitionID      string `json:"policyDefinitionId"`
		PolicyAssignmentName    string `json:"policyAssignmentName"`
		PolicyAssignmentID      string `json:"policyAssignmentId"`
		ComplianceState         string `json:"complianceState"`
	}
	for _, r := range graph.UnmarshalRows[policyRow](result.Data, "Azure Policy") {
		if filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
			continue
		}

		if filters.Azqr.IsServiceExcluded(r.ResourceID) {
			continue
		}

		rec := &models.AzurePolicyResult{
			SubscriptionID:       r.SubscriptionID,
			SubscriptionName:     r.SubscriptionName,
			Type:                 models.GetResourceTypeFromResourceID(r.ResourceID),
			ResourceGroupName:    models.GetResourceGroupFromResourceID(r.ResourceID),
			Name:                 models.GetResourceNameFromResourceID(r.ResourceID),
			PolicyDisplayName:    r.PolicyDefinitionDisplay,
			PolicyDescription:    r.PolicyDescription,
			ResourceID:           r.ResourceID,
			TimeStamp:            r.Timestamp,
			PolicyDefinitionName: r.PolicyDefinitionName,
			PolicyDefinitionID:   r.PolicyDefinitionID,
			PolicyAssignmentName: r.PolicyAssignmentName,
			PolicyAssignmentID:   r.PolicyAssignmentID,
			ComplianceState:      r.ComplianceState,
		}

		// Create unique composite key - avoids string concatenation
		key := policyKey{
			resourceID:         rec.ResourceID,
			policyDefinitionID: rec.PolicyDefinitionID,
		}

		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			resources = append(resources, rec)
		}
	}

	return resources
}
