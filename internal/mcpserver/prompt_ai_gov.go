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
1. Use the scan-ai-gov tool and read its sheets: "AI Gov Usage" (hourly tokens,
   requests, 429 rate-limited and failed requests, and latency), "AI Gov Cost"
   (estimated USD cost, pricing source and assumptions), "AI Gov PTU"
   (deployment details, Provisioned Throughput Unit recommendation and four dimension verdicts),
   "AI Gov Workloads" (cost per technical workload) and "AI Gov Insights"
   (evidence-grounded findings and data gaps).
2. Calculate throttling rate for each Deployment/Model/Environment combination:
   - Throttling Rate (%) = (Rate-Limited Requests / Request Count) × 100
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
     rather than treating the estimate as exact. Costs are list USD retail
     prices for token-billed standard deployments only; check the Workloads
     "Cost Status" and "Priced Token Share (%)" columns before comparing
     workloads, and do not read "N/A" as zero
   - PTU recommendation per deployment: highlight "PTU recommended"
     deployments, and treat "Borderline", "Insufficient evidence" and
     "Capacity data required" as preliminary. Prefer the Window Throttle Rate
     column; it is N/A when request outcomes are incomplete. TPM columns are
     PTU-weighted, and the PTU "Notes" and "Observed Span (days)" columns say
     when monthly estimates extrapolate a short window or use fallback rates.
     Also review the "AI Gov Insights" collection-gap findings before trusting
     any conclusion
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
