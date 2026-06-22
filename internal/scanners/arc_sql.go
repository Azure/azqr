// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// ArcSQLScanner scans Arc-enabled machines with SQL Server for extension installation compliance
type ArcSQLScanner struct{}

// Scan queries Azure Resource Graph for Arc-enabled machines with SQL Server discovered but without the SQL Server extension installed
func (s *ArcSQLScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.ArcSQLResult {
	models.LogResourceTypeScan("Azure Arc-enabled SQL Server")

	graphClient := graph.NewGraphQuery(cred)
	query := `
		resources
		| where type =~ "Microsoft.AzureArcData/sqlServerInstances"
		| extend SQLInstance = id, AzureArcServer = tolower(tostring(properties.containerResourceId))
		| extend version = tostring(properties.version)
		| extend edition = tostring(properties.edition)
		| extend Build = tostring(properties.currentVersion)
		| extend DefenderStatus = tostring(properties.azureDefenderStatus)
		| extend patchLevel = tostring(properties.patchLevel)
		| extend vcores = toint(properties.vCore)
		| join kind=inner (resources
		| where type == 'microsoft.hybridcompute/machines/extensions' 
		| where properties.type == "WindowsAgent.SqlServer"
		| order by ['id'] asc
		| extend License = case(properties.settings.LicenseType == "Paid","SA",properties.settings.LicenseType == "PAYG","PAYG","unset")
		| extend Serverid = tolower(tostring(split(id,'/extensions/WindowsAgent.SqlServer')[0]))
		| parse properties with * 'uploadStatus : ' DPSStatus ';' *
		| parse properties with * 'telemetryUploadStatus : ' TELStatusRaw ';' *
		| extend DPSStatus = iff(DPSStatus == "", "No Data",DPSStatus)
		| extend TELStatuslogs = (parse_json(replace('.\"','\"',TELStatusRaw))).logs
		| extend TELStatus = iff(TELStatuslogs.status == "OK","__",iff(TELStatuslogs.message == "","No Data",TELStatuslogs.message))
		) on $left.AzureArcServer == $right.Serverid
		| join kind=inner (resources
		| where type == "microsoft.hybridcompute/machines"
		| extend status = tostring(properties.status)
		| project id = tolower(id),status) on $left.AzureArcServer == $right.id
		| project subscriptionId,status,AzureArcServer,SQLInstance,resourceGroup,version,Build,patchLevel,edition,vcores,License,DPSStatus,TELStatus,DefenderStatus
		`

	log.Debug().Msg(query)
	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for Arc SQL resources")
		return nil
	}
	resources := []*models.ArcSQLResult{}

	type arcSQLRow struct {
		SubscriptionID string `json:"subscriptionId"`
		Status         string `json:"status"`
		AzureArcServer string `json:"AzureArcServer"`
		SQLInstance    string `json:"SQLInstance"`
		ResourceGroup  string `json:"resourceGroup"`
		Version        string `json:"version"`
		Build          string `json:"Build"`
		PatchLevel     string `json:"patchLevel"`
		Edition        string `json:"edition"`
		VCores         string `json:"vcores"`
		License        string `json:"License"`
		DPSStatus      string `json:"DPSStatus"`
		TELStatus      string `json:"TELStatus"`
		DefenderStatus string `json:"DefenderStatus"`
	}
	for _, r := range graph.UnmarshalRows[arcSQLRow](result.Data, "Arc SQL") {
		if filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
			continue
		}

		if filters.Azqr.IsServiceExcluded(r.SQLInstance) {
			continue
		}

		subscriptionName, ok := subscriptions[r.SubscriptionID]
		if !ok {
			subscriptionName = ""
		}

		resources = append(resources, &models.ArcSQLResult{
			SubscriptionID:   r.SubscriptionID,
			SubscriptionName: subscriptionName,
			Status:           r.Status,
			AzureArcServer:   r.AzureArcServer,
			SQLInstance:      r.SQLInstance,
			ResourceGroup:    r.ResourceGroup,
			Version:          r.Version,
			Build:            r.Build,
			PatchLevel:       r.PatchLevel,
			Edition:          r.Edition,
			VCores:           r.VCores,
			License:          r.License,
			DPSStatus:        r.DPSStatus,
			TELStatus:        r.TELStatus,
			DefenderStatus:   r.DefenderStatus,
		})
	}

	return resources
}
