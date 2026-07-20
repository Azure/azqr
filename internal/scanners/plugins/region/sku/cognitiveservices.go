// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"context"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// cognitiveServicesSKUItem is one entry from the Cognitive Services global SKU list.
// The endpoint returns all SKUs across all regions; each item lists its supported
// locations so we filter to the requested region in the extract function.
// Reference: GET /subscriptions/{sub}/providers/Microsoft.CognitiveServices/skus
type cognitiveServicesSKUItem struct {
	ResourceType string   `json:"resourceType"` // e.g. "accounts"
	Name         string   `json:"name"`          // e.g. "S0", "F0"
	Kind         string   `json:"kind"`          // e.g. "OpenAI", "SpeechServices", "Face"
	Locations    []string `json:"locations"`     // lowercase or display region names
}

// CognitiveServicesProvider checks Azure Cognitive Services (including Azure OpenAI)
// SKU availability per region.
//
// The API returns a subscription-scoped global list of all SKUs with their supported
// regions, similar to how Microsoft.Storage/skus works. We filter to the requested
// region in the extract function.
//
// Result keys are the lowercase SKU name (e.g. "s0", "f0") which matches the
// sku.name field returned by ARG for microsoft.cognitiveservices/accounts resources.
//
// Note: This reports tier availability, not per-model OpenAI deployment quota.
// TPM/capacity quota for specific model deployments requires a separate quota query.
type CognitiveServicesProvider struct{}

func (CognitiveServicesProvider) ResourceType() string {
	return "microsoft.cognitiveservices/accounts"
}

func (CognitiveServicesProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	url := "https://management.azure.com/subscriptions/" + subID + "/providers/Microsoft.CognitiveServices/skus?api-version=2024-10-01"
	target := types.NormalizeRegionName(region)

	return FetchPagedSKUs(ctx, url, client, func(item cognitiveServicesSKUItem) (string, types.SKUAvailability, bool) {
		if item.Name == "" {
			return "", types.SKUAvailability{}, false
		}
		// Filter: only keep items whose location list includes the target region.
		found := len(item.Locations) == 0 // empty locations = globally available
		for _, loc := range item.Locations {
			if strings.EqualFold(types.NormalizeRegionName(loc), target) {
				found = true
				break
			}
		}
		if !found {
			return "", types.SKUAvailability{}, false
		}
		return item.Name, types.SKUAvailability{State: types.SKUAvailable}, true
	})
}

func init() { Register(CognitiveServicesProvider{}) }
