// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
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
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, to.Ptr(s))
	}
	// Composite key type for deduplication - avoids string concatenation allocations
	type policyKey struct {
		resourceID         string
		policyDefinitionID string
	}

	result := graphClient.Query(ctx, query, subs)
	resources := []*models.AzurePolicyResult{}
	seen := make(map[policyKey]struct{}, len(result.Data))

	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			if filters.Azqr.IsSubscriptionExcluded(to.String(m["subscriptionId"])) {
				continue
			}

			resourceId := to.String(m["resourceId"])
			if filters.Azqr.IsServiceExcluded(resourceId) {
				continue
			}

			rec := &models.AzurePolicyResult{
				SubscriptionID:       to.String(m["subscriptionId"]),
				SubscriptionName:     to.String(m["subscriptionName"]),
				Type:                 models.GetResourceTypeFromResourceID(resourceId),
				ResourceGroupName:    models.GetResourceGroupFromResourceID(resourceId),
				Name:                 models.GetResourceNameFromResourceID(resourceId),
				PolicyDisplayName:    to.String(m["policyDefinitionDisplayName"]),
				PolicyDescription:    to.String(m["policyDescription"]),
				ResourceID:           resourceId,
				TimeStamp:            to.String(m["timestamp"]),
				PolicyDefinitionName: to.String(m["policyDefinitionName"]),
				PolicyDefinitionID:   to.String(m["policyDefinitionId"]),
				PolicyAssignmentName: to.String(m["policyAssignmentName"]),
				PolicyAssignmentID:   to.String(m["policyAssignmentId"]),
				ComplianceState:      to.String(m["complianceState"]),
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
	}

	return resources
}
