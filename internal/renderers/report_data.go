// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/models"
)

type (
	ReportData struct {
		OutputFileName          string
		Mask                    bool
		Graph                   []*models.GraphResult                             `json:"graph,omitempty"`
		Defender                []*models.DefenderResult                          `json:"defender,omitempty"`
		DefenderRecommendations []*models.DefenderRecommendation                  `json:"defenderRecommendations,omitempty"`
		Advisor                 []*models.AdvisorResult                           `json:"advisor,omitempty"`
		AzurePolicy             []*models.AzurePolicyResult                       `json:"azurePolicy,omitempty"`
		ArcSQL                  []*models.ArcSQLResult                            `json:"arcSQL,omitempty"`
		Cost                    []*models.CostResult                              `json:"cost,omitempty"`
		Recommendations         map[string]map[string]*models.GraphRecommendation `json:"-"`
		Resources               []*models.Resource                                `json:"resources,omitempty"`
		ExludedResources        []*models.Resource                                `json:"-"`
		ResourceTypeCount       []*models.ResourceTypeCount                       `json:"resourceTypeCount,omitempty"`
		PluginResults           []*PluginResult                                   `json:"pluginResults,omitempty"`
		Stages                  *models.StageConfigs                              `json:"-"`

		// Table caches - populated on first call, reused thereafter
		cachedImpactedTable                [][]string `json:"-"`
		cachedCostTable                    [][]string `json:"-"`
		cachedDefenderTable                [][]string `json:"-"`
		cachedAdvisorTable                 [][]string `json:"-"`
		cachedAzurePolicyTable             [][]string `json:"-"`
		cachedArcSQLTable                  [][]string `json:"-"`
		cachedRecommendationsTable         [][]string `json:"-"`
		cachedResourceTypesTable           [][]string `json:"-"`
		cachedDefenderRecommendationsTable [][]string `json:"-"`
		cachedResourcesTable               [][]string `json:"-"`
		cachedExcludedResourcesTable       [][]string `json:"-"`
	}

	// PluginResult represents data from an external plugin
	PluginResult struct {
		PluginName  string     // Name of the plugin
		SheetName   string     // Name for Excel sheet
		Description string     // Description of the data
		Table       [][]string // Table data (first row is headers)
	}

	ResourceTypeCountResults struct {
		ResourceType []models.ResourceTypeCount `json:"ResourceType"`
	}
)

func (rd *ReportData) ResourcesTable() [][]string {
	if rd.cachedResourcesTable != nil {
		return rd.cachedResourcesTable
	}
	rd.cachedResourcesTable = rd.resourcesTable(rd.Resources)
	return rd.cachedResourcesTable
}

func (rd *ReportData) ExcludedResourcesTable() [][]string {
	if rd.cachedExcludedResourcesTable != nil {
		return rd.cachedExcludedResourcesTable
	}
	rd.cachedExcludedResourcesTable = rd.resourcesTable(rd.ExludedResources)
	return rd.cachedExcludedResourcesTable
}

func (rd *ReportData) ImpactedTable() [][]string {
	if rd.cachedImpactedTable != nil {
		return rd.cachedImpactedTable
	}

	headers := []string{"Validated Using", "Source", "Category", "Impact", "Resource Type", "Recommendation", "Recommendation Id", "Subscription Id", "Subscription Name", "Resource Group", "Resource Name", "Resource Id", "Param1", "Param2", "Param3", "Param4", "Param5", "Learn"}

	// Composite key type for deduplication - avoids string concatenation allocations
	type impactedKey struct {
		resourceID       string
		recommendationID string
	}

	// Pre-allocate with estimated capacity (graph length + 1 for headers)
	// This avoids multiple slice reallocations
	rows := make([][]string, 1, len(rd.Graph)+1)
	rows[0] = headers

	// Use struct{} instead of bool to save memory
	seen := make(map[impactedKey]struct{}, len(rd.Graph))

	for _, r := range rd.Graph {
		// Cache string conversions once to avoid repeated type conversions
		category := string(r.Category)
		if skipCategory(category) {
			continue
		}

		// Create composite key - avoids string concatenation
		key := impactedKey{
			resourceID:       r.ResourceID,
			recommendationID: r.RecommendationID,
		}
		if _, exists := seen[key]; exists {
			continue // Skip duplicate
		}
		seen[key] = struct{}{}

		impact := string(r.Impact)

		row := []string{
			"Azure Resource Graph",
			r.Source,
			category,
			impact,
			r.ResourceType,
			r.Recommendation,
			r.RecommendationID,
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.SubscriptionName,
			r.ResourceGroup,
			r.Name,
			MaskSubscriptionIDInResourceID(r.ResourceID, rd.Mask),
			r.Param1,
			r.Param2,
			r.Param3,
			r.Param4,
			r.Param5,
			r.Learn,
		}
		rows = append(rows, row)
	}

	rd.cachedImpactedTable = rows
	return rows
}

func (rd *ReportData) CostTable() [][]string {
	if rd.cachedCostTable != nil {
		return rd.cachedCostTable
	}

	headers := []string{"From", "To", "Subscription Id", "Subscription Name", "Service Name", "Value", "Currency"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.Cost)+1)
	rows[0] = headers

	for _, r := range rd.Cost {
		row := []string{
			r.From.Format("2006-01-02"),
			r.To.Format("2006-01-02"),
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.SubscriptionName,
			r.ServiceName,
			r.Value,
			r.Currency,
		}
		rows = append(rows, row)
	}

	rd.cachedCostTable = rows
	return rows
}

func (rd *ReportData) DefenderTable() [][]string {
	if rd.cachedDefenderTable != nil {
		return rd.cachedDefenderTable
	}

	headers := []string{"Subscription Id", "Subscription Name", "Name", "Tier"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.Defender)+1)
	rows[0] = headers

	for _, d := range rd.Defender {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			d.Name,
			d.Tier,
		}
		rows = append(rows, row)
	}

	rd.cachedDefenderTable = rows
	return rows
}

func (rd *ReportData) AdvisorTable() [][]string {
	if rd.cachedAdvisorTable != nil {
		return rd.cachedAdvisorTable
	}

	headers := []string{"Subscription Id", "Subscription Name", "Resource Type", "Resource Name", "Category", "Impact", "Description", "Resource Id", "Recommendation Id"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.Advisor)+1)
	rows[0] = headers

	for _, d := range rd.Advisor {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			d.Type,
			d.Name,
			d.Category,
			d.Impact,
			d.Description,
			MaskSubscriptionIDInResourceID(d.ResourceID, rd.Mask),
			d.RecommendationID,
		}
		rows = append(rows, row)
	}

	rd.cachedAdvisorTable = rows
	return rows
}

// AzurePolicyTable returns Azure Policy data formatted as a table with headers and rows for reporting.
func (rd *ReportData) AzurePolicyTable() [][]string {
	if rd.cachedAzurePolicyTable != nil {
		return rd.cachedAzurePolicyTable
	}

	headers := []string{"Subscription Id", "Subscription Name", "Resource Group", "Resource Type", "Resource Name", "Policy Display Name", "Policy Description", "Resource Id", "Time Stamp", "Policy Definition Name", "Policy Definition Id", "Policy Assignment Name", "Policy Assignment Id", "Compliance State"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.AzurePolicy)+1)
	rows[0] = headers

	for _, d := range rd.AzurePolicy {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			d.ResourceGroupName,
			d.Type,
			d.Name,
			d.PolicyDisplayName,
			d.PolicyDescription,
			MaskSubscriptionIDInResourceID(d.ResourceID, rd.Mask),
			d.TimeStamp,
			d.PolicyDefinitionName,
			d.PolicyDefinitionID,
			d.PolicyAssignmentName,
			d.PolicyAssignmentID,
			d.ComplianceState,
		}
		rows = append(rows, row)
	}

	rd.cachedAzurePolicyTable = rows
	return rows
}

// ArcSQLTable returns Arc-enabled SQL Server data formatted as a table with headers and rows for reporting.
func (rd *ReportData) ArcSQLTable() [][]string {
	if rd.cachedArcSQLTable != nil {
		return rd.cachedArcSQLTable
	}

	headers := []string{"Subscription Id", "Subscription Name", "Azure Arc Server", "SQL Instance", "Resource Group", "Version", "Build", "Patch Level", "Edition", "VCores", "License", "DPS Status", "TEL Status", "Defender Status"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.ArcSQL)+1)
	rows[0] = headers

	for _, d := range rd.ArcSQL {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			models.GetResourceNameFromResourceID(d.AzureArcServer),
			models.GetResourceNameFromResourceID(d.SQLInstance),
			d.ResourceGroup,
			d.Version,
			d.Build,
			d.PatchLevel,
			d.Edition,
			d.VCores,
			d.License,
			d.DPSStatus,
			d.TELStatus,
			d.DefenderStatus,
		}
		rows = append(rows, row)
	}

	rd.cachedArcSQLTable = rows
	return rows
}

func (rd *ReportData) RecommendationsTable() [][]string {
	if rd.cachedRecommendationsTable != nil {
		return rd.cachedRecommendationsTable
	}

	counter := map[string]int{}
	for _, rt := range rd.Recommendations {
		for _, r := range rt {
			counter[r.RecommendationID] = 0
		}
	}

	for _, r := range rd.Graph {
		counter[r.RecommendationID]++
	}

	headers := []string{"Implemented", "Number of Impacted Resources", "Azure Service / Well-Architected", "Recommendation Source",
		"Azure Service Category / Well-Architected Area", "Azure Service / Well-Architected Topic", "Category", "Recommendation",
		"Impact", "Best Practices Guidance", "Read More", "Recommendation Id"}

	// Estimate capacity based on recommendations count
	estimatedCap := len(counter) + 1
	rows := make([][]string, 1, estimatedCap)
	rows[0] = headers

	deployedTypes := make(map[string]bool, len(rd.ResourceTypeCount))
	for _, resType := range rd.ResourceTypeCount {
		deployedTypes[strings.ToLower(resType.ResourceType)] = true
	}
	// Always consider Microsoft.Resources as deployed
	deployedTypes["microsoft.resources"] = true

	for _, rt := range rd.Recommendations {
		for _, r := range rt {
			if skipCategory(r.Category) {
				continue
			}
			typeIsDeployed := deployedTypes[strings.ToLower(r.ResourceType)]

			var implemented string
			switch {
			case !typeIsDeployed:
				implemented = "N/A"
			case counter[r.RecommendationID] == 0:
				implemented = "true"
			default:
				implemented = "false"
			}

			categoryPart := ""
			servicePart := ""
			typeParts := strings.Split(r.ResourceType, "/")
			categoryPart = typeParts[0]
			if len(typeParts) > 1 {
				servicePart = typeParts[1]
			}

			// Cache string conversions
			category := string(r.Category)
			impact := string(r.Impact)

			row := []string{
				implemented,
				fmt.Sprint(counter[r.RecommendationID]),
				"Azure Service",
				r.Source,
				categoryPart,
				servicePart,
				category,
				r.Recommendation,
				impact,
				r.LongDescription,
				r.LearnMoreLink[0].Url,
				r.RecommendationID,
			}
			rows = append(rows, row)
		}
	}

	rd.cachedRecommendationsTable = rows
	return rows
}

func (rd *ReportData) ResourceTypesTable() [][]string {
	if rd.cachedResourceTypesTable != nil {
		return rd.cachedResourceTypesTable
	}

	headers := []string{"Subscription Name", "Resource Type", "Number of Resources"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.ResourceTypeCount)+1)
	rows[0] = headers

	for _, r := range rd.ResourceTypeCount {
		row := []string{
			r.Subscription,
			r.ResourceType,
			fmt.Sprint(r.Count),
		}
		rows = append(rows, row)
	}

	rd.cachedResourceTypesTable = rows
	return rows
}

func (rd *ReportData) DefenderRecommendationsTable() [][]string {
	if rd.cachedDefenderRecommendationsTable != nil {
		return rd.cachedDefenderRecommendationsTable
	}

	headers := []string{"Subscription Id", "Subscription Name", "Resource Group", "Resource Type", "Resource Name", "Category", "Recommendation Severity", "Recommendation Name", "Action Description", "Remediation Description", "AzPortal Link", "Resource Id"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(rd.DefenderRecommendations)+1)
	rows[0] = headers

	for _, d := range rd.DefenderRecommendations {
		row := []string{
			MaskSubscriptionID(d.SubscriptionId, rd.Mask),
			d.SubscriptionName,
			d.ResourceGroupName,
			d.ResourceType,
			d.ResourceName,
			d.Category,
			d.RecommendationSeverity,
			d.RecommendationName,
			d.ActionDescription,
			d.RemediationDescription,
			d.AzPortalLink,
			MaskSubscriptionIDInResourceID(d.ResourceId, rd.Mask),
		}
		rows = append(rows, row)
	}

	rd.cachedDefenderRecommendationsTable = rows
	return rows
}

// ClearTableCache clears all cached table data, forcing regeneration on next call.
// Useful if underlying data changes after initial table generation.
func (rd *ReportData) ClearTableCache() {
	rd.cachedImpactedTable = nil
	rd.cachedCostTable = nil
	rd.cachedDefenderTable = nil
	rd.cachedAdvisorTable = nil
	rd.cachedAzurePolicyTable = nil
	rd.cachedArcSQLTable = nil
	rd.cachedRecommendationsTable = nil
	rd.cachedResourceTypesTable = nil
	rd.cachedDefenderRecommendationsTable = nil
	rd.cachedResourcesTable = nil
	rd.cachedExcludedResourcesTable = nil
}

func NewReportData(outputFile string, mask bool, stages *models.StageConfigs) ReportData {
	return ReportData{
		OutputFileName:          outputFile,
		Mask:                    mask,
		Recommendations:         map[string]map[string]*models.GraphRecommendation{},
		Graph:                   []*models.GraphResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost:                    []*models.CostResult{},
		ResourceTypeCount:       []*models.ResourceTypeCount{},
		Stages:                  stages,
	}
}

func MaskSubscriptionID(subscriptionID string, mask bool) string {
	if len(subscriptionID) < 36 {
		return ""
	}

	if !mask {
		return subscriptionID
	}

	// Show only last 7 chars of the subscription ID
	return fmt.Sprintf("xxxxxxxx-xxxx-xxxx-xxxx-xxxxx%s", subscriptionID[29:])
}

func MaskSubscriptionIDInResourceID(resourceID string, mask bool) string {
	if !strings.HasPrefix(resourceID, "/subscriptions/") {
		return ""
	}

	if !mask {
		return resourceID
	}

	parts := strings.Split(resourceID, "/")
	parts[2] = MaskSubscriptionID(parts[2], mask)

	return strings.Join(parts, "/")
}

func (rd *ReportData) resourcesTable(resources []*models.Resource) [][]string {
	headers := []string{"Subscription Id", "Resource Group", "Location", "Resource Type", "Resource Name", "Sku Name", "Sku Tier", "Kind", "SLA", "Resource Id"}

	// Pre-allocate with capacity to avoid reallocations
	rows := make([][]string, 1, len(resources)+1)
	rows[0] = headers

	slaMap := make(map[string]string, len(rd.Graph))
	for _, a := range rd.Graph {
		if a.Category == models.CategorySLA {
			// Use lowercase for case-insensitive lookup
			slaMap[strings.ToLower(a.ResourceID)] = a.Param1
		}
	}

	for _, r := range resources {
		sla := slaMap[strings.ToLower(r.ID)]

		row := []string{
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.ResourceGroup,
			r.Location,
			r.Type,
			r.Name,
			r.SkuName,
			r.SkuTier,
			r.Kind,
			sla,
			MaskSubscriptionIDInResourceID(r.ID, rd.Mask),
		}
		rows = append(rows, row)
	}

	return rows
}

func skipCategory(category string) bool {
	return category == string(models.CategorySLA)
}
