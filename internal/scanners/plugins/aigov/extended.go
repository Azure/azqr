// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"fmt"
	"sort"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// usageHoursLookback matches the throttling collector's 167-hour (~7 day)
// window so all AI Gov sheets describe the same observation period.
const usageHoursLookback = 167

// deploymentModelKey identifies one deployment/model pair within an account.
type deploymentModelKey struct {
	Deployment string
	Model      string
}

// workloadRollup accumulates cost and evidence across all deployments that
// map to the same workload (see defaultWorkloadID).
type workloadRollup struct {
	CostUSD       float64
	ActiveBuckets int
}

// collectExtendedSheets builds the "AI Gov Usage", "AI Gov Cost", "AI Gov
// PTU", and "AI Gov Workloads" sheets. This is the ported subset of
// TokenLens-for-Azure's (https://github.com/zakarel/tokenlens-for-azure)
// analytical capability — discovery, usage aggregation, verified-pricing
// cost resolution, and PTU suitability — deliberately excluding its HTML
// report/dashboard, interactive wizard/CLI, and client-side instrumentation
// layers. azqr's existing Excel/CSV/JSON renderers remain the only reporting
// surface for this data.
func (s *AIGovScanner) collectExtendedSheets(ctx context.Context, cred azcore.TokenCredential, httpClient *az.HttpClient, subscriptions map[string]string, resources []*models.Resource) []plugins.ExternalPluginOutput {
	usageTable := [][]string{usageSheetMetadata().HeaderRow()}
	costTable := [][]string{costSheetMetadata().HeaderRow()}
	ptuTable := [][]string{ptuSheetMetadata().HeaderRow()}
	workloadTable := [][]string{workloadsSheetMetadata().HeaderRow()}

	sheets := func() []plugins.ExternalPluginOutput {
		return []plugins.ExternalPluginOutput{
			{Metadata: usageSheetMetadata(), SheetName: "AI Gov Usage", Description: "Hourly token usage by deployment/model", Table: usageTable},
			{Metadata: costSheetMetadata(), SheetName: "AI Gov Cost", Description: "Estimated cost by deployment/model using verified pricing only", Table: costTable},
			{Metadata: ptuSheetMetadata(), SheetName: "AI Gov PTU", Description: "PTU suitability and evidence quality by deployment", Table: ptuTable},
			{Metadata: workloadsSheetMetadata(), SheetName: "AI Gov Workloads", Description: "Estimated cost rolled up per technical workload", Table: workloadTable},
		}
	}

	if len(resources) == 0 {
		return sheets()
	}

	resourceGroups := s.groupResourcesForBatch(resources)
	workloadRollups := map[string]*workloadRollup{}

	for _, group := range resourceGroups {
		if err := ctx.Err(); err != nil {
			break
		}

		subscriptionName := subscriptions[group.SubscriptionID]
		if subscriptionName == "" {
			subscriptionName = group.SubscriptionID
		}

		resourceMap := make(map[string]*models.Resource, len(group.Resources))
		resourceIDs := make([]string, 0, len(group.Resources))
		for _, r := range group.Resources {
			resourceMap[r.ID] = r
			resourceIDs = append(resourceIDs, r.ID)
		}

		for i := 0; i < len(resourceIDs); i += 50 {
			end := min(i+50, len(resourceIDs))
			batch := resourceIDs[i:end]

			usage, err := s.getBatchUsageMetrics(ctx, cred, group.Region, batch, usageHoursLookback)
			if err != nil {
				log.Debug().Err(err).Msg("failed to collect AI Gov usage metrics for batch")
				continue
			}

			for resourceID, hourly := range usage {
				resource := resourceMap[resourceID]
				if resource == nil {
					continue
				}

				perDeployModel := map[deploymentModelKey][]hourlyTokenUsage{}
				for hourKey, deployMap := range hourly {
					for deployment, modelMap := range deployMap {
						for model, u := range modelMap {
							key := deploymentModelKey{Deployment: deployment, Model: model}
							perDeployModel[key] = append(perDeployModel[key], u)
							usageTable = append(usageTable, []string{
								subscriptionName, resource.ResourceGroup, resource.Name,
								deployment, model, hourKey,
								fmt.Sprintf("%.0f", u.InputTokens), fmt.Sprintf("%.0f", u.OutputTokens),
							})
						}
					}
				}

				for key, samples := range perDeployModel {
					var totalInput, totalOutput float64
					for _, s := range samples {
						totalInput += s.InputTokens
						totalOutput += s.OutputTokens
					}

					price, priceOK := resolvePrice(ctx, httpClient, key.Model, resource.Location)
					var cost float64
					priceSource := "unavailable"
					if priceOK {
						cost = estimatedCostUSD(price, totalInput, totalOutput)
						priceSource = price.Source
					}

					costTable = append(costTable, []string{
						subscriptionName, resource.ResourceGroup, resource.Name,
						key.Deployment, key.Model,
						fmt.Sprintf("%.0f", totalInput), fmt.Sprintf("%.0f", totalOutput),
						priceSource, fmt.Sprintf("%.2f", cost),
						costRemediation(priceOK, priceSource),
					})

					avgTPM, p95TPM, activeBuckets := tpmStats(samples)
					assessment := assessPTU(key.Model, avgTPM, p95TPM, price, activeBuckets)
					ptuTable = append(ptuTable, []string{
						subscriptionName, resource.ResourceGroup, resource.Name,
						key.Deployment, key.Model,
						fmt.Sprintf("%.1f", avgTPM), fmt.Sprintf("%.1f", p95TPM),
						assessment.Eligibility, fmt.Sprintf("%d", assessment.RecommendedPTUs),
						fmt.Sprintf("%.2f", assessment.PAYGMonthlyUSD), fmt.Sprintf("%.2f", assessment.PTUMonthlyUSD),
						fmt.Sprintf("%.0f", assessment.BreakEvenTPM),
						ptuRemediation(assessment),
					})

					workloadID := defaultWorkloadID(resource.Name, key.Deployment)
					rollup := workloadRollups[workloadID]
					if rollup == nil {
						rollup = &workloadRollup{}
						workloadRollups[workloadID] = rollup
					}
					rollup.CostUSD += cost
					rollup.ActiveBuckets += activeBuckets
				}
			}
		}
	}

	workloadIDs := make([]string, 0, len(workloadRollups))
	for id := range workloadRollups {
		workloadIDs = append(workloadIDs, id)
	}
	sort.Strings(workloadIDs)
	for _, id := range workloadIDs {
		rollup := workloadRollups[id]
		workloadTable = append(workloadTable, []string{
			id, fmt.Sprintf("%.2f", rollup.CostUSD), fmt.Sprintf("%d", rollup.ActiveBuckets),
		})
	}

	return sheets()
}
