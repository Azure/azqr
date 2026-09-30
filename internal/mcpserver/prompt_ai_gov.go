// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleAIGovPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prompt := `Check AI/Cognitive Services throttling, usage, cost, and PTU suitability for Azure resources.

Please:
1. Use the scan-ai-gov tool to detect 429 throttling errors, and read its
   additional sheets: "AI Gov Usage" (hourly token counts), "AI Gov Cost"
   (estimated USD cost and pricing source), "AI Gov PTU" (Provisioned
   Throughput Unit eligibility and evidence), and "AI Gov Workloads"
   (cost rolled up per technical workload).
2. Calculate throttling rate for each Deployment/Model/Environment combination:
   - Throttling Rate (%) = (429 status count / total requests) × 100
   - HIGHLIGHT combinations where throttling rate > 1% as CRITICAL
   - Flag combinations with 0.1-1% as WARNING
   - Mark combinations with < 0.1% as HEALTHY
3. Analyze the results focusing on:
   - Instances experiencing throttling (group by account, deployment, model)
   - Deployments and models affected (calculate rates per combination)
   - Time patterns of throttling (identify peak hours and recurring patterns)
   - Spillover configuration status (check if spillover is enabled)
   - Estimated cost per deployment/model and per workload; note when a
     "Pricing Source" is "unavailable" (no verified rate could be resolved)
     rather than treating the estimate as exact
   - PTU eligibility per deployment: highlight deployments already
     "eligible" for PTU with sustained high TPM, and note deployments with
     "insufficient_evidence" or "capacity_data_required" where the
     recommendation should be treated as preliminary
4. Provide prioritized recommendations for:
   - Capacity planning and scaling (focus on CRITICAL combinations first)
   - Load distribution strategies using Azure APIM
   - Spillover configuration optimization
   - Cost optimization opportunities (e.g., moving high, steady-TPM
     deployments from pay-as-you-go to PTU when eligible)
   - Immediate actions vs. long-term improvements
`

		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompt))

		return mcp.NewGetPromptResult(
			"check AI governance",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
