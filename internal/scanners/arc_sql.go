// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// ArcSQLScanner scans Arc-enabled machines with SQL Server for extension installation compliance
type ArcSQLScanner struct{}

// Scan queries Azure Resource Graph for Arc-enabled machines with SQL Server discovered but without the SQL Server extension installed
func (s *ArcSQLScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []models.ArcSQLResult {
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
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}
	result := graphClient.Query(ctx, query, subs)
	resources := []models.ArcSQLResult{}

	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			subscription := to.String(m["subscriptionId"])
			if filters.Azqr.IsSubscriptionExcluded(subscription) {
				continue
			}

			sqlInstance := to.String(m["SQLInstance"])
			if filters.Azqr.IsServiceExcluded(sqlInstance) {
				continue
			}

			subscriptionName, ok := subscriptions[subscription]
			if !ok {
				subscriptionName = ""
			}

			resources = append(resources, models.ArcSQLResult{
				SubscriptionID:   subscription,
				SubscriptionName: subscriptionName,
				Status:           to.String(m["status"]),
				AzureArcServer:   to.String(m["AzureArcServer"]),
				SQLInstance:      sqlInstance,
				ResourceGroup:    to.String(m["resourceGroup"]),
				Version:          to.String(m["version"]),
				Build:            to.String(m["Build"]),
				PatchLevel:       to.String(m["patchLevel"]),
				Edition:          to.String(m["edition"]),
				VCores:           to.String(m["vcores"]),
				License:          to.String(m["License"]),
				DPSStatus:        to.String(m["DPSStatus"]),
				TELStatus:        to.String(m["TELStatus"]),
				DefenderStatus:   to.String(m["DefenderStatus"]),
			})
		}
	}

	return resources
}
