// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
	"github.com/rs/zerolog/log"
)

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
			{Name: "Spillover Enabled", DataKey: "spilloverEnabled", FilterType: plugins.FilterTypeDropdown},
			{Name: "Spillover Deployment", DataKey: "spilloverDeployment", FilterType: plugins.FilterTypeSearch},
			{Name: "Hour", DataKey: "hour", FilterType: plugins.FilterTypeSearch},
			{Name: "Status Code", DataKey: "statusCode", FilterType: plugins.FilterTypeDropdown},
			{Name: "Request Count", DataKey: "requestCount", FilterType: plugins.FilterTypeNone},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *ThrottlingScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	// Initialize table with headers
	table := [][]string{
		{"Subscription", "Resource Group", "Account Name", "Kind", "SKU", "Deployment Name", "Model Name", "Spillover Enabled", "Spillover Deployment", "Hour", "Status Code", "Request Count"},
	}

	// Use Resource Graph to get all Cognitive Services/OpenAI accounts efficiently
	resourceScanner := scanners.ResourceScanner{}
	allResources, _ := resourceScanner.GetAllResources(ctx, cred, subscriptions, filters)
	log.Debug().Msgf("Found %d total resources", len(allResources))

	// Filter for OpenAI and AI Services accounts
	openAIResources := make([]*models.Resource, 0)
	for _, resource := range allResources {
		if strings.EqualFold(resource.Type, "Microsoft.CognitiveServices/accounts") {
			// Filter by kind if available
			if resource.Kind != "" {
				kindLower := strings.ToLower(resource.Kind)
				if strings.Contains(kindLower, "openai") || strings.Contains(kindLower, "aiservices") {
					openAIResources = append(openAIResources, resource)
				}
			} else {
				// If no kind info, include it (will be filtered later when getting deployments)
				openAIResources = append(openAIResources, resource)
			}
		}
	}
	log.Debug().Msgf("Filtered to %d OpenAI/AI Services accounts", len(openAIResources))

	// Group resources by subscription and region for batch processing
	resourceGroups := s.groupResourcesForBatch(openAIResources)
	log.Debug().Msgf("Created %d resource batch groups", len(resourceGroups))

	// Process each batch group
	for groupIdx, group := range resourceGroups {
		log.Debug().Int("group", groupIdx+1).Int("total", len(resourceGroups)).Str("subscription", group.SubscriptionID).Str("region", group.Region).Int("resources", len(group.Resources)).Msg("Processing group")

		// Process resources in batches of up to 50
		for i := 0; i < len(group.Resources); i += 50 {
			end := i + 50
			if end > len(group.Resources) {
				end = len(group.Resources)
			}
			batch := group.Resources[i:end]
			log.Debug().Int("start", i).Int("end", end).Int("total", len(group.Resources)).Msg("Processing batch")

			// Collect throttling data for batch
			results, err := s.processBatch(ctx, cred, batch, subscriptions)
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

// processBatch processes a batch of resources using batch metrics API
func (s *ThrottlingScanner) processBatch(ctx context.Context, cred azcore.TokenCredential, resources []*models.Resource, subscriptions map[string]string) ([][]string, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	subscriptionID := resources[0].SubscriptionID
	region := resources[0].Location
	subscriptionName := subscriptions[subscriptionID]
	if subscriptionName == "" {
		subscriptionName = subscriptionID
	}

	// Create client options
	clientOptions := az.NewDefaultClientOptions()

	// Create Deployments client for getting deployment info
	deploymentsClient, err := armcognitiveservices.NewDeploymentsClient(subscriptionID, cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployments client: %w", err)
	}

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

		// Get deployments for this account
		deployments, err := s.getDeployments(ctx, deploymentsClient, resourceGroup, accountName)
		if err != nil {
			log.Debug().Err(err).Str("account", accountName).Msg("Failed to get deployments")
			deployments = []*armcognitiveservices.Deployment{}
		}

		// Create a map of deployment names to their properties
		deploymentMap := make(map[string]*armcognitiveservices.Deployment)
		for _, deployment := range deployments {
			if deployment.Name != nil {
				deploymentMap[*deployment.Name] = deployment
			}
		}

		// Create rows for each hour/deployment/model/status code combination
		for hourKey, deploymentMetrics := range hourlyMetrics {
			for metricDeploymentName, models := range deploymentMetrics {
				for modelName, statusCodes := range models {
					for statusCode, count := range statusCodes {
						// Look up deployment properties
						spilloverEnabled := "No"
						spilloverDeployment := "N/A"

						if deployment, exists := deploymentMap[metricDeploymentName]; exists {
							if deployment.Properties != nil && deployment.Properties.SpilloverDeploymentName != nil {
								spilloverEnabled = "Yes"
								spilloverDeployment = *deployment.Properties.SpilloverDeploymentName
							}
						}

						results = append(results, []string{
							subscriptionName,
							resourceGroup,
							accountName,
							kind,
							skuName,
							metricDeploymentName,
							modelName,
							spilloverEnabled,
							spilloverDeployment,
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
