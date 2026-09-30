// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"

func testPlan() metrics.Plan {
	return metrics.Plan{Input: "InputTokens", Output: "OutputTokens", Requests: "AzureOpenAIRequests", StatusSplit: true}
}
