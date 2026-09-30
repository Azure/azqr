// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics"
	"github.com/rs/zerolog/log"
)

// hourlyTokenUsage holds the input/output token counts observed for one
// deployment/model in one hourly bucket.
type hourlyTokenUsage struct {
	InputTokens  float64
	OutputTokens float64
}

// usageMetricNames are the Azure Monitor metric names for Azure OpenAI token
// counters. These are the primary, documented metric names (see Azure OpenAI
// monitoring data reference); TokenLens's metrics_catalog.py additionally
// tracks lower-priority fallback names (ProcessedPromptTokens/GeneratedTokens,
// PromptTokenCount/CompletionTokenCount) for non-Azure-OpenAI-branded
// deployments, which is left as a documented gap rather than guessed at here.
var usageMetricNames = []string{"InputTokens", "OutputTokens"}

// getBatchUsageMetrics queries Azure Monitor batch metrics for token usage,
// mirroring getBatchMetricsWithStatusCodeSplit's shape but for the "AI Gov
// Usage"/"AI Gov Cost"/"AI Gov PTU" sheets instead of throttling.
// Returns: resourceID -> hourKey -> deploymentName -> modelName -> usage.
func (s *AIGovScanner) getBatchUsageMetrics(ctx context.Context, cred azcore.TokenCredential, region string, resourceIDs []string, hours int) (map[string]map[string]map[string]map[string]hourlyTokenUsage, error) {
	endpoint := fmt.Sprintf("https://%s.metrics.monitor.azure.com", region)
	client, err := azmetrics.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client for region %s: %w", region, err)
	}

	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	namespace := "Microsoft.CognitiveServices/accounts"
	options := &azmetrics.QueryResourcesOptions{
		StartTime:   to.Ptr(startTime.Format(time.RFC3339)),
		EndTime:     to.Ptr(endTime.Format(time.RFC3339)),
		Interval:    to.Ptr("PT1H"),
		Aggregation: to.Ptr("Total"),
		Filter:      to.Ptr("ModelDeploymentName eq '*' and ModelName eq '*'"),
	}

	response, err := client.QueryResources(ctx, resourceIDsSubscriptionID(resourceIDs), namespace, usageMetricNames, azmetrics.ResourceIDList{ResourceIDs: resourceIDs}, options)
	if err != nil {
		log.Debug().Err(err).Msg("batch query error for AI Gov usage metrics")
		return nil, fmt.Errorf("failed to query batch usage metrics: %w", err)
	}

	result := make(map[string]map[string]map[string]map[string]hourlyTokenUsage)

	for _, metricData := range response.Values {
		if metricData.ResourceID == nil {
			continue
		}
		resourceID := *metricData.ResourceID

		for _, metric := range metricData.Values {
			if metric.Name == nil || metric.TimeSeries == nil {
				continue
			}
			metricName := *metric.Name.Value

			for _, timeseries := range metric.TimeSeries {
				deploymentName := "Unknown"
				modelName := "Unknown"
				if timeseries.MetadataValues != nil {
					for _, meta := range timeseries.MetadataValues {
						if meta.Name != nil && meta.Name.Value != nil && meta.Value != nil {
							switch strings.ToLower(*meta.Name.Value) {
							case "modeldeploymentname":
								deploymentName = *meta.Value
							case "modelname":
								modelName = *meta.Value
							}
						}
					}
				}

				if timeseries.Data == nil {
					continue
				}
				for _, data := range timeseries.Data {
					if data.TimeStamp == nil || data.Total == nil {
						continue
					}
					hourKey := data.TimeStamp.Format("2006-01-02 15:00")

					if result[resourceID] == nil {
						result[resourceID] = make(map[string]map[string]map[string]hourlyTokenUsage)
					}
					if result[resourceID][hourKey] == nil {
						result[resourceID][hourKey] = make(map[string]map[string]hourlyTokenUsage)
					}
					if result[resourceID][hourKey][deploymentName] == nil {
						result[resourceID][hourKey][deploymentName] = make(map[string]hourlyTokenUsage)
					}

					usage := result[resourceID][hourKey][deploymentName][modelName]
					switch metricName {
					case "InputTokens":
						usage.InputTokens += *data.Total
					case "OutputTokens":
						usage.OutputTokens += *data.Total
					}
					result[resourceID][hourKey][deploymentName][modelName] = usage
				}
			}
		}
	}

	return result, nil
}

// resourceIDsSubscriptionID extracts the subscription ID from the first
// resource ID in the list; all resource IDs in a batch share the same
// subscription (see groupResourcesForBatch).
func resourceIDsSubscriptionID(resourceIDs []string) string {
	if len(resourceIDs) == 0 {
		return ""
	}
	parts := strings.Split(resourceIDs[0], "/")
	for i, p := range parts {
		if strings.EqualFold(p, "subscriptions") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
