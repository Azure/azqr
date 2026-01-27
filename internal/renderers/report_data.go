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
		Graph                   []*models.GraphResult
		Defender                []*models.DefenderResult
		DefenderRecommendations []*models.DefenderRecommendation
		Advisor                 []*models.AdvisorResult
		AzurePolicy             []*models.AzurePolicyResult
		ArcSQL                  []*models.ArcSQLResult
		Cost                    *models.CostResult
		Recommendations         map[string]map[string]models.GraphRecommendation
		Resources               []*models.Resource
		ExludedResources        []*models.Resource
		ResourceTypeCount       []models.ResourceTypeCount
		PluginResults           []PluginResult // Results from external plugins
		Stages                  *models.StageConfigs
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
	return rd.resourcesTable(rd.Resources)
}

func (rd *ReportData) ExcludedResourcesTable() [][]string {
	return rd.resourcesTable(rd.ExludedResources)
}

func (rd *ReportData) ImpactedTable() [][]string {
	headers := []string{"Validated Using", "Source", "Category", "Impact", "Resource Type", "Recommendation", "Recommendation Id", "Subscription Id", "Subscription Name", "Resource Group", "Resource Name", "Resource Id", "Param1", "Param2", "Param3", "Param4", "Param5", "Learn"}

	// Use a map to track unique entries by ResourceID + RecommendationID
	seen := make(map[string]bool)
	rows := [][]string{}

	for _, r := range rd.Graph {

		if skipCategory(string(r.Category)) {
			continue
		}

		// Create composite key for deduplication
		key := r.ResourceID + "|" + r.RecommendationID
		if seen[key] {
			continue // Skip duplicate
		}
		seen[key] = true

		row := []string{
			"Azure Resource Graph",
			r.Source,
			string(r.Category),
			string(r.Impact),
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

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) CostTable() [][]string {
	headers := []string{"From", "To", "Subscription Id", "Subscription Name", "Service Name", "Value", "Currency"}

	rows := [][]string{}
	for _, r := range rd.Cost.Items {
		row := []string{
			rd.Cost.From.Format("2006-01-02"),
			rd.Cost.To.Format("2006-01-02"),
			MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.SubscriptionName,
			r.ServiceName,
			r.Value,
			r.Currency,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) DefenderTable() [][]string {
	headers := []string{"Subscription Id", "Subscription Name", "Name", "Tier"}
	rows := [][]string{}
	for _, d := range rd.Defender {
		row := []string{
			MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.SubscriptionName,
			d.Name,
			d.Tier,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) AdvisorTable() [][]string {
	headers := []string{"Subscription Id", "Subscription Name", "Resource Type", "Resource Name", "Category", "Impact", "Description", "Resource Id", "Recommendation Id"}
	rows := [][]string{}
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

	rows = append([][]string{headers}, rows...)
	return rows
}

// AzurePolicyTable returns Azure Policy data formatted as a table with headers and rows for reporting.
func (rd *ReportData) AzurePolicyTable() [][]string {
	headers := []string{"Subscription Id", "Subscription Name", "Resource Group", "Resource Type", "Resource Name", "Policy Display Name", "Policy Description", "Resource Id", "Time Stamp", "Policy Definition Name", "Policy Definition Id", "Policy Assignment Name", "Policy Assignment Id", "Compliance State"}
	rows := [][]string{}
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

	rows = append([][]string{headers}, rows...)
	return rows
}

// ArcSQLTable returns Arc-enabled SQL Server data formatted as a table with headers and rows for reporting.
func (rd *ReportData) ArcSQLTable() [][]string {
	headers := []string{"Subscription Id", "Subscription Name", "Azure Arc Server", "SQL Instance", "Resource Group", "Version", "Build", "Patch Level", "Edition", "VCores", "License", "DPS Status", "TEL Status", "Defender Status"}
	rows := [][]string{}
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

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) RecommendationsTable() [][]string {
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
	rows := [][]string{}
	for _, rt := range rd.Recommendations {
		for _, r := range rt {
			if skipCategory(r.Category) {
				continue
			}
			typeIsDeployed := false
			for _, resType := range rd.ResourceTypeCount {
				if strings.EqualFold(resType.ResourceType, r.ResourceType) ||
					strings.EqualFold(r.ResourceType, "Microsoft.Resources") {
					typeIsDeployed = true
					break
				}
			}

			implemented := "N/A"
			switch {
			case typeIsDeployed && counter[r.RecommendationID] == 0:
				implemented = "true"
			case typeIsDeployed && counter[r.RecommendationID] > 0:
				implemented = "false"
			}

			categoryPart := ""
			servicePart := ""
			typeParts := strings.Split(r.ResourceType, "/")
			categoryPart = typeParts[0]
			if len(typeParts) > 1 {
				servicePart = typeParts[1]
			}

			row := []string{
				implemented,
				fmt.Sprint(counter[r.RecommendationID]),
				"Azure Service",
				r.Source,
				categoryPart,
				servicePart,
				string(r.Category),
				r.Recommendation,
				string(r.Impact),
				r.LongDescription,
				r.LearnMoreLink[0].Url,
				r.RecommendationID,
			}
			rows = append(rows, row)
		}
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) ResourceTypesTable() [][]string {
	headers := []string{"Subscription Name", "Resource Type", "Number of Resources"}
	rows := [][]string{}
	for _, r := range rd.ResourceTypeCount {
		row := []string{
			r.Subscription,
			r.ResourceType,
			fmt.Sprint(r.Count),
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) DefenderRecommendationsTable() [][]string {
	headers := []string{"Subscription Id", "Subscription Name", "Resource Group", "Resource Type", "Resource Name", "Category", "Recommendation Severity", "Recommendation Name", "Action Description", "Remediation Description", "AzPortal Link", "Resource Id"}
	rows := [][]string{}
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

	rows = append([][]string{headers}, rows...)
	return rows
}

func NewReportData(outputFile string, mask bool, stages *models.StageConfigs) ReportData {
	return ReportData{
		OutputFileName:          outputFile,
		Mask:                    mask,
		Recommendations:         map[string]map[string]models.GraphRecommendation{},
		Graph:                   []*models.GraphResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost: &models.CostResult{
			Items: []*models.CostResultItem{},
		},
		ResourceTypeCount: []models.ResourceTypeCount{},
		Stages:            stages,
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

	rows := [][]string{}
	for _, r := range resources {
		sla := ""

		for _, a := range rd.Graph {
			if strings.EqualFold(strings.ToLower(a.ResourceID), strings.ToLower(r.ID)) {
				if a.Category == models.CategorySLA {
					sla = a.Param1
				}
				if sla != "" {
					break
				}
			}
		}

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

	rows = append([][]string{headers}, rows...)
	return rows
}

func skipCategory(category string) bool {
	return category == string(models.CategorySLA)
}
