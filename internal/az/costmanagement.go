// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
)

// costManagementAPIVersion matches armcostmanagement.QueryClient's generated Query API version.
const costManagementAPIVersion = "2021-10-01"

// QueryCostManagementPages executes a Cost Management usage query for scope and follows
// pagination via raw HTTP POST requests through httpClient.
//
// armcostmanagement.QueryClient has no pager for its Usage method: the Cost Management query
// operation isn't modeled as paged in its API spec, so the generated SDK only ever sends one
// request. Azure's own Az.CostManagement PowerShell module hits the same gap and works around
// it identically — a raw REST call for every page, including the first — so this follows the
// same approach instead of mixing a typed SDK call with a raw HTTP fallback.
//
// pageFn is invoked once per non-empty page; return an error from pageFn to stop early.
func QueryCostManagementPages(ctx context.Context, httpClient *HttpClient, scope string, definition armcostmanagement.QueryDefinition, pageFn func(*armcostmanagement.QueryProperties) error) error {
	bodyBytes, err := json.Marshal(definition)
	if err != nil {
		return fmt.Errorf("failed to marshal Cost Management query definition: %w", err)
	}

	url := fmt.Sprintf("%s%s/providers/Microsoft.CostManagement/query?api-version=%s", GetResourceManagerEndpoint(), scope, costManagementAPIVersion)

	for url != "" {
		responseBody, resp, err := httpClient.DoPost(ctx, url, NopReadSeekCloser{Reader: bytes.NewReader(bodyBytes)})
		if err != nil {
			return fmt.Errorf("failed to query Cost Management API: %w", err)
		}
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		var result armcostmanagement.QueryResult
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return fmt.Errorf("failed to unmarshal Cost Management results: %w", err)
		}
		if result.Properties == nil {
			return nil
		}
		if err := pageFn(result.Properties); err != nil {
			return err
		}
		if result.Properties.NextLink == nil || *result.Properties.NextLink == "" {
			return nil
		}
		url = *result.Properties.NextLink
	}
	return nil
}
