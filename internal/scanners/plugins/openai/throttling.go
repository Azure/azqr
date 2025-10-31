// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
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
			{Name: "Hour", DataKey: "hour", FilterType: plugins.FilterTypeSearch},
			{Name: "Model Name", DataKey: "modelName", FilterType: plugins.FilterTypeDropdown},
			{Name: "Status Code", DataKey: "statusCode", FilterType: plugins.FilterTypeDropdown},
			{Name: "Request Count", DataKey: "requestCount", FilterType: plugins.FilterTypeNone},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *ThrottlingScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	// Initialize table with headers
	table := [][]string{
		{"Subscription", "Resource Group", "Account Name", "Kind", "SKU", "Hour", "Model Name", "Status Code", "Request Count"},
	}

	// Scan all subscriptions
	for subID, subName := range subscriptions {
		results, err := s.scanSubscription(ctx, cred, subID, subName)
		if err != nil {
			// Skip subscriptions with errors but don't fail the entire scan
			continue
		}
		table = append(table, results...)
	}

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "OpenAI Throttling",
		Description: "Analysis of throttling errors for OpenAI/Cognitive Services accounts by hour, model, and status code",
		Table:       table,
	}, nil
}

// scanSubscription scans a single subscription for OpenAI accounts
func (s *ThrottlingScanner) scanSubscription(ctx context.Context, cred azcore.TokenCredential, subscriptionID, subscriptionName string) ([][]string, error) {
	clientOptions := &policy.ClientOptions{}

	// Create Cognitive Services client
	cognitiveClient, err := armcognitiveservices.NewAccountsClient(subscriptionID, cred, &arm.ClientOptions{
		ClientOptions: *clientOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cognitive services client: %w", err)
	}

	// Create Monitor metrics client
	metricsClient, err := armmonitor.NewMetricsClient(subscriptionID, cred, &arm.ClientOptions{
		ClientOptions: *clientOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// List all Cognitive Services accounts
	pager := cognitiveClient.NewListPager(nil)
	var results [][]string

	for pager.More() {
		// Wait for rate limiter
		<-throttling.ARMLimiter

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list accounts: %w", err)
		}

		for _, account := range page.Value {
			if account.ID == nil || account.Name == nil {
				continue
			}

			// Only check OpenAI accounts
			if account.Kind == nil || !strings.Contains(strings.ToLower(*account.Kind), "openai") {
				continue
			}

			// Get metrics for the last 7 days
			hourlyMetrics, err := s.getMetricsWithStatusCodeSplit(ctx, metricsClient, *account.ID, 167)
			if err != nil {
				// Skip accounts with metrics errors but don't fail the entire scan
				continue
			}

			// Extract resource group from ID
			resourceGroup := extractResourceGroup(*account.ID)

			// Get SKU name
			skuName := "Unknown"
			if account.SKU != nil && account.SKU.Name != nil {
				skuName = string(*account.SKU.Name)
			}

			kind := "Unknown"
			if account.Kind != nil {
				kind = *account.Kind
			}

			// Create one row per hour per model per status code
			for hourKey, models := range hourlyMetrics {
				for modelName, statusCodes := range models {
					for statusCode, count := range statusCodes {
						results = append(results, []string{
							subscriptionName,
							resourceGroup,
							*account.Name,
							kind,
							skuName,
							hourKey,
							modelName,
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

// getMetricsWithStatusCodeSplit queries Azure Monitor for request metrics split by status code and model
func (s *ThrottlingScanner) getMetricsWithStatusCodeSplit(ctx context.Context, client *armmonitor.MetricsClient, resourceID string, hours int) (map[string]map[string]map[string]float64, error) {
	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)
	timespan := fmt.Sprintf("%s/%s",
		startTime.Format("2006-01-02T15:04:05Z"),
		endTime.Format("2006-01-02T15:04:05Z"))

	metricNames := "AzureOpenAIRequests"
	interval := "PT1H"

	// Use wildcard filter to get all status codes and model names
	filter := "StatusCode eq '*' and ModelName eq '*'"

	result, err := client.List(ctx, resourceID, &armmonitor.MetricsClientListOptions{
		Timespan:    &timespan,
		Interval:    &interval,
		Metricnames: &metricNames,
		Filter:      &filter,
		Aggregation: to.Ptr("Count"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Aggregate by hour, model name, and status code: map[hour]map[modelName]map[statusCode]count
	hourlyData := make(map[string]map[string]map[string]float64)

	for _, metric := range result.Value {
		if metric.Timeseries == nil {
			continue
		}
		for _, timeseries := range metric.Timeseries {
			// Extract status code and model name from metadata
			statusCode := "Unknown"
			modelName := "Unknown"
			if len(timeseries.Metadatavalues) > 0 {
				for _, meta := range timeseries.Metadatavalues {
					if meta.Name != nil && meta.Value != nil {
						if meta.Name.Value != nil {
							if strings.EqualFold(*meta.Name.Value, "StatusCode") {
								statusCode = *meta.Value
							} else if strings.EqualFold(*meta.Name.Value, "ModelName") {
								modelName = *meta.Value
							}
						}
					}
				}
			}

			// Process each data point (hourly)
			if timeseries.Data != nil {
				for _, data := range timeseries.Data {
					if data.TimeStamp != nil {
						// Format hour as readable string
						hourKey := data.TimeStamp.Format("2006-01-02 15:00")

						if hourlyData[hourKey] == nil {
							hourlyData[hourKey] = make(map[string]map[string]float64)
						}
						if hourlyData[hourKey][modelName] == nil {
							hourlyData[hourKey][modelName] = make(map[string]float64)
						}
						// Use Count field (which we requested via Aggregation)
						if data.Count != nil {
							hourlyData[hourKey][modelName][statusCode] += *data.Count
						}
					}
				}
			}
		}
	}

	return hourlyData, nil
}

// extractResourceGroup extracts the resource group name from an Azure resource ID
func extractResourceGroup(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "Unknown"
}

// init registers this plugin with the internal plugin registry
func init() {
	scanner := NewThrottlingScanner()
	plugins.RegisterInternalPlugin("openai-throttling", scanner)
}
