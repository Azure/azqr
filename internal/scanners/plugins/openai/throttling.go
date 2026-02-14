// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package openai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
	"github.com/rs/zerolog/log"
)

// maxConcurrentDeploymentFetches limits parallel ARM calls when listing deployments.
const maxConcurrentDeploymentFetches = 5

// ThrottlingScanner is an internal plugin that monitors OpenAI throttling
type ThrottlingScanner struct{}

// NewThrottlingScanner creates a new OpenAI throttling scanner
func NewThrottlingScanner() *ThrottlingScanner {
	return &ThrottlingScanner{}
}

// GetMetadata returns plugin metadata
func (s *ThrottlingScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "openai-throttling",
		Version:     "1.0.0",
		Description: "Checks OpenAI/Cognitive Services accounts for 429 throttling errors",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription", DataKey: "subscription", FilterType: plugins.FilterTypeSearch},
			{Name: "Resource Group", DataKey: "resourceGroup", FilterType: plugins.FilterTypeSearch},
			{Name: "Account Name", DataKey: "accountName", FilterType: plugins.FilterTypeSearch},
			{Name: "Kind", DataKey: "kind", FilterType: plugins.FilterTypeDropdown},
			{Name: "SKU", DataKey: "sku", FilterType: plugins.FilterTypeDropdown},
			{Name: "Deployment Name", DataKey: "deploymentName", FilterType: plugins.FilterTypeSearch},
			{Name: "Model Name", DataKey: "modelName", FilterType: plugins.FilterTypeDropdown},
			{Name: "Model Version", DataKey: "modelVersion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Model Format", DataKey: "modelFormat", FilterType: plugins.FilterTypeDropdown},
			{Name: "SKU Capacity", DataKey: "skuCapacity", FilterType: plugins.FilterTypeNone},
			{Name: "Version Upgrade Option", DataKey: "versionUpgradeOption", FilterType: plugins.FilterTypeDropdown},
			{Name: "Spillover Enabled", DataKey: "spilloverEnabled", FilterType: plugins.FilterTypeDropdown},
			{Name: "Spillover Deployment", DataKey: "spilloverDeployment", FilterType: plugins.FilterTypeSearch},
			{Name: "Hour", DataKey: "hour", FilterType: plugins.FilterTypeSearch},
			{Name: "Status Code", DataKey: "statusCode", FilterType: plugins.FilterTypeDropdown},
			{Name: "Request Count", DataKey: "requestCount", FilterType: plugins.FilterTypeNone},
		},
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
}

// extractDeploymentInfo pre-computes display values from a Deployment once.
func extractDeploymentInfo(deployment *armcognitiveservices.Deployment) deploymentInfo {
	info := deploymentInfo{
		ModelVersion:         "N/A",
		ModelFormat:          "N/A",
		SKUCapacity:          "N/A",
		VersionUpgradeOption: "N/A",
		SpilloverEnabled:     "No",
		SpilloverDeployment:  "N/A",
	}

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
	if deployment.SKU != nil && deployment.SKU.Capacity != nil {
		info.SKUCapacity = fmt.Sprintf("%d", *deployment.SKU.Capacity)
	}

	return info
}

// Scan executes the plugin and returns table data
func (s *ThrottlingScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	// Initialize table with headers
	table := [][]string{
		{"Subscription", "Resource Group", "Account Name", "Kind", "SKU", "Deployment Name", "Model Name", "Model Version", "Model Format", "SKU Capacity", "Version Upgrade Option", "Spillover Enabled", "Spillover Deployment", "Hour", "Status Code", "Request Count"},
	}

	// Use a targeted Resource Graph query instead of fetching all resources.
	// This avoids downloading the entire resource inventory when only
	// CognitiveServices/accounts with OpenAI or AI Services kind are needed.
	openAIResources, err := s.discoverOpenAIResources(ctx, cred, subscriptions, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OpenAI resources: %w", err)
	}
	log.Debug().Msgf("Discovered %d OpenAI/AI Services accounts", len(openAIResources))

	if len(openAIResources) == 0 {
		return &plugins.ExternalPluginOutput{
			Metadata:    s.GetMetadata(),
			SheetName:   "OpenAI Throttling",
			Description: "Analysis of throttling errors for OpenAI/Cognitive Services accounts by hour, model, and status code",
			Table:       table,
		}, nil
	}

	// Group resources by subscription and region for batch processing
	resourceGroups := s.groupResourcesForBatch(openAIResources)
	log.Debug().Msgf("Created %d resource batch groups", len(resourceGroups))

	// Create client options once and cache DeploymentsClient per subscription.
	clientOptions := az.NewDefaultClientOptions()
	deploymentsClients := make(map[string]*armcognitiveservices.DeploymentsClient)

	// Process each batch group
	for groupIdx, group := range resourceGroups {
		// Check for context cancellation between groups
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("scan cancelled: %w", err)
		}

		log.Debug().Int("group", groupIdx+1).Int("total", len(resourceGroups)).Str("subscription", group.SubscriptionID).Str("region", group.Region).Int("resources", len(group.Resources)).Msg("Processing group")

		// Reuse DeploymentsClient for the same subscription across batches/regions
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

		// Process resources in batches of up to 50
		for i := 0; i < len(group.Resources); i += 50 {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("scan cancelled: %w", err)
			}

			end := min(i+50, len(group.Resources))
			batch := group.Resources[i:end]
			log.Debug().Int("start", i).Int("end", end).Int("total", len(group.Resources)).Msg("Processing batch")

			// Collect throttling data for batch
			results, err := s.processBatch(ctx, cred, deploymentsClient, batch, subscriptionName)
			if err != nil {
				log.Debug().Err(err).Msg("Error processing batch")
				continue
			}

			log.Debug().Msgf("Collected %d results from batch", len(results))
			table = append(table, results...)
		}
	}

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "OpenAI Throttling",
		Description: "Analysis of throttling errors for OpenAI/Cognitive Services accounts by hour, model, and status code",
		Table:       table,
	}, nil
}

// discoverOpenAIResources queries Azure Resource Graph for CognitiveServices accounts
// with an OpenAI or AI Services kind, avoiding a full resource inventory download.
func (s *ThrottlingScanner) discoverOpenAIResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) ([]*models.Resource, error) {
	graphClient := graph.NewGraphQuery(cred)

	query := `resources
| where type =~ "Microsoft.CognitiveServices/accounts"
| where isempty(kind) or kind contains "openai" or kind contains "aiservices"
| project id, subscriptionId, resourceGroup, location, type, name, sku.name, sku.tier, kind
| order by subscriptionId, resourceGroup`

	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, to.Ptr(s))
	}

	result := graphClient.Query(ctx, query, subs)

	resources := make([]*models.Resource, 0, len(result.Data))
	for _, row := range result.Data {
		m, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := m["id"].(string)
		if filters.Azqr.IsServiceExcluded(id) {
			continue
		}

		skuName, _ := m["sku_name"].(string)
		skuTier, _ := m["sku_tier"].(string)
		kind, _ := m["kind"].(string)
		resourceGroup, _ := m["resourceGroup"].(string)
		location, _ := m["location"].(string)

		resources = append(resources, &models.Resource{
			ID:             id,
			SubscriptionID: m["subscriptionId"].(string),
			ResourceGroup:  resourceGroup,
			Location:       location,
			Type:           m["type"].(string),
			Name:           m["name"].(string),
			SkuName:        skuName,
			SkuTier:        skuTier,
			Kind:           kind,
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
func (s *ThrottlingScanner) groupResourcesForBatch(resources []*models.Resource) []ResourceBatchGroup {
	groups := make(map[string]*ResourceBatchGroup)

	for _, resource := range resources {
		// Batch API requires: same subscription and region
		key := fmt.Sprintf("%s|%s", resource.SubscriptionID, resource.Location)

		if group, exists := groups[key]; exists {
			group.Resources = append(group.Resources, resource)
		} else {
			groups[key] = &ResourceBatchGroup{
				SubscriptionID: resource.SubscriptionID,
				Region:         resource.Location,
				Resources:      []*models.Resource{resource},
			}
		}
	}

	// Convert map to slice
	result := make([]ResourceBatchGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	return result
}

// accountDeployments holds the result of a concurrent deployment fetch.
type accountDeployments struct {
	ResourceID    string
	DeploymentMap map[string]deploymentInfo
	Err           error
}

// processBatch processes a batch of resources using batch metrics API
func (s *ThrottlingScanner) processBatch(ctx context.Context, cred azcore.TokenCredential, deploymentsClient *armcognitiveservices.DeploymentsClient, resources []*models.Resource, subscriptionName string) ([][]string, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	subscriptionID := resources[0].SubscriptionID
	region := resources[0].Location

	// Build resource ID list for batch query
	resourceIDs := make([]string, 0, len(resources))
	resourceMap := make(map[string]*models.Resource)
	for _, resource := range resources {
		resourceIDs = append(resourceIDs, resource.ID)
		resourceMap[resource.ID] = resource
	}

	// Query metrics using batch API
	batchMetrics, err := s.getBatchMetricsWithStatusCodeSplit(ctx, cred, subscriptionID, region, resourceIDs, 167)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch metrics: %w", err)
	}

	// Fetch deployments concurrently for all resources that have metrics.
	// Use a semaphore to bound the number of parallel ARM calls.
	sem := make(chan struct{}, maxConcurrentDeploymentFetches)
	deploymentsCh := make(chan accountDeployments, len(batchMetrics))
	var wg sync.WaitGroup

	for resourceID := range batchMetrics {
		resource, exists := resourceMap[resourceID]
		if !exists {
			continue
		}

		wg.Add(1)
		go func(res *models.Resource) {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			deployments, fetchErr := s.getDeployments(ctx, deploymentsClient, res.ResourceGroup, res.Name)
			result := accountDeployments{ResourceID: res.ID, Err: fetchErr}
			if fetchErr == nil {
				result.DeploymentMap = make(map[string]deploymentInfo, len(deployments))
				for _, d := range deployments {
					if d.Name != nil {
						result.DeploymentMap[*d.Name] = extractDeploymentInfo(d)
					}
				}
			}
			deploymentsCh <- result
		}(resource)
	}

	wg.Wait()
	close(deploymentsCh)

	// Collect deployment maps keyed by resource ID.
	allDeployments := make(map[string]map[string]deploymentInfo, len(batchMetrics))
	for ad := range deploymentsCh {
		if ad.Err != nil {
			log.Debug().Err(ad.Err).Str("resourceID", ad.ResourceID).Msg("Failed to get deployments")
			continue
		}
		allDeployments[ad.ResourceID] = ad.DeploymentMap
	}

	// Default deployment info for unmatched deployments.
	defaultInfo := deploymentInfo{
		ModelVersion:         "N/A",
		ModelFormat:          "N/A",
		SKUCapacity:          "N/A",
		VersionUpgradeOption: "N/A",
		SpilloverEnabled:     "No",
		SpilloverDeployment:  "N/A",
	}

	var results [][]string

	// Process metrics for each resource
	for resourceID, hourlyMetrics := range batchMetrics {
		resource, exists := resourceMap[resourceID]
		if !exists {
			continue
		}

		resourceGroup := resource.ResourceGroup
		accountName := resource.Name
		kind := resource.Kind
		if kind == "" {
			kind = "Unknown"
		}

		skuName := resource.SkuName
		if skuName == "" {
			skuName = "Unknown"
		}

		depMap := allDeployments[resourceID] // may be nil if fetch failed

		// Create rows for each hour/deployment/model/status code combination
		for hourKey, deploymentMetrics := range hourlyMetrics {
			for metricDeploymentName, mdls := range deploymentMetrics {
				// Look up deployment metadata once per deployment name, not per metric row.
				info := defaultInfo
				if depMap != nil {
					if di, ok := depMap[metricDeploymentName]; ok {
						info = di
					}
				}

				for modelName, statusCodes := range mdls {
					for statusCode, count := range statusCodes {
						results = append(results, []string{
							subscriptionName,
							resourceGroup,
							accountName,
							kind,
							skuName,
							metricDeploymentName,
							modelName,
							info.ModelVersion,
							info.ModelFormat,
							info.SKUCapacity,
							info.VersionUpgradeOption,
							info.SpilloverEnabled,
							info.SpilloverDeployment,
							hourKey,
							statusCode,
							fmt.Sprintf("%.0f", count),
						})
					}
				}
			}
		}
	}

	return results, nil
}

// getBatchMetricsWithStatusCodeSplit queries Azure Monitor batch API for request metrics
func (s *ThrottlingScanner) getBatchMetricsWithStatusCodeSplit(ctx context.Context, cred azcore.TokenCredential, subscriptionID, region string, resourceIDs []string, hours int) (map[string]map[string]map[string]map[string]map[string]float64, error) {
	// Create regional metrics client
	endpoint := fmt.Sprintf("https://%s.metrics.monitor.azure.com", region)
	client, err := azmetrics.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client for region %s: %w", region, err)
	}

	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	log.Debug().Str("subscription", subscriptionID).Int("resources", len(resourceIDs)).Msg("Querying batch metrics for throttling")

	metricNames := []string{"AzureOpenAIRequests"}
	namespace := "Microsoft.CognitiveServices/accounts"

	// Build options with filters for status code, deployment name, and model name
	options := &azmetrics.QueryResourcesOptions{
		StartTime:   to.Ptr(startTime.Format(time.RFC3339)),
		EndTime:     to.Ptr(endTime.Format(time.RFC3339)),
		Interval:    to.Ptr("PT1H"),
		Aggregation: to.Ptr("Count"),
		Filter:      to.Ptr("StatusCode eq '*' and ModelDeploymentName eq '*' and ModelName eq '*'"),
	}

	// Create ResourceIDList
	resourceIDList := azmetrics.ResourceIDList{
		ResourceIDs: resourceIDs,
	}

	// Query batch metrics
	response, err := client.QueryResources(ctx, subscriptionID, namespace, metricNames, resourceIDList, options)
	if err != nil {
		log.Debug().Err(err).Msg("Batch query error for throttling metrics")
		return nil, fmt.Errorf("failed to query batch metrics: %w", err)
	}

	log.Debug().Msgf("Batch query successful, returned %d metric data entries", len(response.Values))

	// Parse response: map[resourceID]map[hour]map[deploymentName]map[modelName]map[statusCode]count
	result := make(map[string]map[string]map[string]map[string]map[string]float64)

	for _, metricData := range response.Values {
		if metricData.ResourceID == nil {
			continue
		}
		resourceID := *metricData.ResourceID

		for _, metric := range metricData.Values {
			if metric.TimeSeries == nil {
				continue
			}

			for _, timeseries := range metric.TimeSeries {
				// Extract metadata
				statusCode := "Unknown"
				deploymentName := "Unknown"
				modelName := "Unknown"

				if timeseries.MetadataValues != nil {
					for _, meta := range timeseries.MetadataValues {
						if meta.Name != nil && meta.Name.Value != nil && meta.Value != nil {
							switch strings.ToLower(*meta.Name.Value) {
							case "statuscode":
								statusCode = *meta.Value
							case "modelname":
								modelName = *meta.Value
							case "modeldeploymentname":
								deploymentName = *meta.Value
							}
						}
					}
				}

				// Process each data point (hourly)
				if timeseries.Data != nil {
					for _, data := range timeseries.Data {
						if data.TimeStamp == nil || data.Count == nil {
							continue
						}

						hourKey := data.TimeStamp.Format("2006-01-02 15:00")

						if result[resourceID] == nil {
							result[resourceID] = make(map[string]map[string]map[string]map[string]float64)
						}
						if result[resourceID][hourKey] == nil {
							result[resourceID][hourKey] = make(map[string]map[string]map[string]float64)
						}
						if result[resourceID][hourKey][deploymentName] == nil {
							result[resourceID][hourKey][deploymentName] = make(map[string]map[string]float64)
						}
						if result[resourceID][hourKey][deploymentName][modelName] == nil {
							result[resourceID][hourKey][deploymentName][modelName] = make(map[string]float64)
						}

						result[resourceID][hourKey][deploymentName][modelName][statusCode] += *data.Count
					}
				}
			}
		}
	}

	return result, nil
}

// getDeployments retrieves all deployments for a cognitive services account
func (s *ThrottlingScanner) getDeployments(ctx context.Context, client *armcognitiveservices.DeploymentsClient, resourceGroup, accountName string) ([]*armcognitiveservices.Deployment, error) {
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
	scanner := NewThrottlingScanner()
	plugins.RegisterInternalPlugin("openai-throttling", scanner)
}
