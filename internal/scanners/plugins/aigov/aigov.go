// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
	"github.com/rs/zerolog/log"
)

// maxConcurrentDeploymentFetches limits parallel ARM calls when listing deployments.
const maxConcurrentDeploymentFetches = 5

// maxResourcesPerBatch is the Azure Monitor batch API limit on resources per call.
const maxResourcesPerBatch = 50

// AIGovScanner is an internal plugin that monitors AI governance
type AIGovScanner struct{}

// NewScanner creates a new AI governance scanner
func NewScanner() *AIGovScanner {
	return &AIGovScanner{}
}

// GetMetadata returns plugin metadata
func (s *AIGovScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "ai-gov",
		Version:        "1.2.0",
		Description:    "Checks AI Governance",
		Author:         "Azure Quick Review Team",
		License:        "MIT",
		Type:           plugins.PluginTypeInternal,
		ColumnMetadata: ptuSheetMetadata().ColumnMetadata,
	}
}

// deploymentInfo holds pre-extracted deployment metadata to avoid repeated field access.
type deploymentInfo struct {
	ModelVersion         string
	ModelFormat          string
	SKUCapacity          string
	VersionUpgradeOption string
	SpilloverEnabled     string
	SpilloverDeployment  string
	SKUName              string
	Mode                 string
}

// newDefaultDeploymentInfo returns display defaults for unknown deployments.
func newDefaultDeploymentInfo() deploymentInfo {
	return deploymentInfo{
		ModelVersion:         "N/A",
		ModelFormat:          "N/A",
		SKUCapacity:          "N/A",
		VersionUpgradeOption: "N/A",
		SpilloverEnabled:     "No",
		SpilloverDeployment:  "N/A",
		SKUName:              "N/A",
		Mode:                 modelinfo.ModeUnknown,
	}
}

// extractDeploymentInfo pre-computes display values from a Deployment once.
func extractDeploymentInfo(deployment *armcognitiveservices.Deployment) deploymentInfo {
	info := newDefaultDeploymentInfo()

	if deployment.Properties != nil {
		if deployment.Properties.SpilloverDeploymentName != nil {
			info.SpilloverEnabled = "Yes"
			info.SpilloverDeployment = *deployment.Properties.SpilloverDeploymentName
		}
		if deployment.Properties.Model != nil {
			if deployment.Properties.Model.Version != nil {
				info.ModelVersion = *deployment.Properties.Model.Version
			}
			if deployment.Properties.Model.Format != nil {
				info.ModelFormat = *deployment.Properties.Model.Format
			}
		}
		if deployment.Properties.VersionUpgradeOption != nil {
			info.VersionUpgradeOption = string(*deployment.Properties.VersionUpgradeOption)
		}
	}
	if deployment.SKU != nil {
		if deployment.SKU.Capacity != nil {
			info.SKUCapacity = fmt.Sprintf("%d", *deployment.SKU.Capacity)
		}
		if deployment.SKU.Name != nil && *deployment.SKU.Name != "" {
			info.SKUName = *deployment.SKU.Name
			info.Mode = modelinfo.DeploymentModeForSKU(*deployment.SKU.Name)
		}
	}

	return info
}

// Scan executes the plugin and returns table data
func (s *AIGovScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	// Use a targeted Resource Graph query instead of fetching all resources.
	// This avoids downloading the entire resource inventory when only
	// CognitiveServices/accounts with OpenAI or AI Services kind are needed.
	openAIResources, err := s.discoverOpenAIResources(ctx, cred, subscriptions, params.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OpenAI resources: %w", err)
	}
	log.Debug().Msgf("Discovered %d OpenAI/AI Services accounts", len(openAIResources))

	httpClient := az.NewHttpClient(cred, nil)

	// Group resources by subscription and region for batch processing
	resourceGroups := s.groupResourcesForBatch(openAIResources)
	log.Debug().Msgf("Created %d resource batch groups", len(resourceGroups))

	// Create client options once and cache DeploymentsClient per subscription.
	clientOptions := az.NewDefaultClientOptions()
	deploymentsClients := make(map[string]*armcognitiveservices.DeploymentsClient)
	var collected []*collectedAccount

	for groupIdx, group := range resourceGroups {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("scan cancelled: %w", err)
		}

		log.Debug().Int("group", groupIdx+1).Int("total", len(resourceGroups)).Str("subscription", group.SubscriptionID).Str("region", group.Region).Int("resources", len(group.Resources)).Msg("Processing group")

		deploymentsClient, exists := deploymentsClients[group.SubscriptionID]
		if !exists {
			var clientErr error
			deploymentsClient, clientErr = armcognitiveservices.NewDeploymentsClient(group.SubscriptionID, cred, clientOptions)
			if clientErr != nil {
				log.Debug().Err(clientErr).Str("subscription", group.SubscriptionID).Msg("Failed to create deployments client, skipping group")
				continue
			}
			deploymentsClients[group.SubscriptionID] = deploymentsClient
		}

		subscriptionName := subscriptions[group.SubscriptionID]
		if subscriptionName == "" {
			subscriptionName = group.SubscriptionID
		}

		for i := 0; i < len(group.Resources); i += maxResourcesPerBatch {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("scan cancelled: %w", err)
			}

			end := min(i+maxResourcesPerBatch, len(group.Resources))
			batch := group.Resources[i:end]
			log.Debug().Int("start", i).Int("end", end).Int("total", len(group.Resources)).Msg("Processing batch")

			accounts := s.collectAccounts(ctx, cred, deploymentsClient, subscriptionName, group.SubscriptionID, group.Region, batch)
			collected = append(collected, accounts...)
		}
	}

	// Usage, Cost, PTU, Insights and Workloads sheets share the same
	// collected buckets so every sheet describes the same observation window.
	return s.collectExtendedSheets(ctx, httpClient, collected), nil
}

// discoverOpenAIResources queries Azure Resource Graph for CognitiveServices accounts
// with an OpenAI or AI Services kind, avoiding a full resource inventory download.
func (s *AIGovScanner) discoverOpenAIResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) ([]*models.Resource, error) {
	graphClient := graph.NewGraphQuery(cred)

	query := `resources
| where type =~ "Microsoft.CognitiveServices/accounts"
| where isempty(kind) or kind contains "openai" or kind contains "aiservices"
| project id, subscriptionId, resourceGroup, location, type, name, sku.name, sku.tier, kind
| order by subscriptionId, resourceGroup`

	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query openai resources from resource graph: %w", err)
	}

	resources := make([]*models.Resource, 0, len(result.Data))
	type openAIRow struct {
		ID             string `json:"id"`
		SubscriptionID string `json:"subscriptionId"`
		ResourceGroup  string `json:"resourceGroup"`
		Location       string `json:"location"`
		Type           string `json:"type"`
		Name           string `json:"name"`
		SkuName        string `json:"sku_name"`
		SkuTier        string `json:"sku_tier"`
		Kind           string `json:"kind"`
	}
	for _, raw := range result.Data {
		var r openAIRow
		if err := json.Unmarshal(raw, &r); err != nil {
			continue
		}

		if filters.Azqr.IsServiceExcluded(r.ID) {
			continue
		}

		resources = append(resources, &models.Resource{
			ID:             r.ID,
			SubscriptionID: r.SubscriptionID,
			ResourceGroup:  r.ResourceGroup,
			Location:       r.Location,
			Type:           r.Type,
			Name:           r.Name,
			SkuName:        r.SkuName,
			SkuTier:        r.SkuTier,
			Kind:           r.Kind,
		})
	}

	return resources, nil
}

// ResourceBatchGroup groups resources by subscription and region for batch processing
type ResourceBatchGroup struct {
	SubscriptionID string
	Region         string
	Resources      []*models.Resource
}

// groupResourcesForBatch groups resources that can be queried together in batch API
func (s *AIGovScanner) groupResourcesForBatch(resources []*models.Resource) []ResourceBatchGroup {
	groups := make(map[string]*ResourceBatchGroup)

	for _, resource := range resources {
		// Batch API requires: same subscription and region
		key := fmt.Sprintf("%s|%s", resource.SubscriptionID, resource.Location)

		group, exists := groups[key]
		if !exists {
			group = &ResourceBatchGroup{SubscriptionID: resource.SubscriptionID, Region: resource.Location}
			groups[key] = group
		}
		group.Resources = append(group.Resources, resource)
	}

	// Convert map to slice
	result := make([]ResourceBatchGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	return result
}

// getDeployments retrieves all deployments for a cognitive services account
func (s *AIGovScanner) getDeployments(ctx context.Context, client *armcognitiveservices.DeploymentsClient, resourceGroup, accountName string) ([]*armcognitiveservices.Deployment, error) {
	pager := client.NewListPager(resourceGroup, accountName, nil)
	var deployments []*armcognitiveservices.Deployment

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list deployments: %w", err)
		}

		deployments = append(deployments, page.Value...)
	}

	return deployments, nil
}

// init registers this plugin with the internal plugin registry
func init() {
	scanner := NewScanner()
	plugins.RegisterInternalPlugin("ai-gov", scanner)
}
