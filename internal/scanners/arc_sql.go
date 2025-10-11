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
		| where type == "microsoft.hybridcompute/machines"
		| where properties.detectedProperties.mssqldiscovered == "true"
		| extend machineIdHasSQLServerDiscovered = id
		| project name, machineIdHasSQLServerDiscovered, resourceGroup, subscriptionId, tags, location, properties
		| join kind= leftouter (
			resources
			| where type == "microsoft.hybridcompute/machines/extensions"
			| where properties.type in ("WindowsAgent.SqlServer","LinuxAgent.SqlServer")
			| extend machineIdHasSQLServerExtensionInstalled = iff(id contains "/extensions/WindowsAgent.SqlServer" or id contains "/extensions/LinuxAgent.SqlServer", substring(id, 0, indexof(id, "/extensions/")), "")
			| project Provisioning_State = properties.provisioningState,
			License_Type = properties.settings.LicenseType,
			ESU = iff(notnull(properties.settings.enableExtendedSecurityUpdates), "enabled", "disabled"),
			Extension_Version = properties.instanceView.typeHandlerVersion,
			Excluded_instances = properties.ExcludedSqlInstances,
			Purview = iff(notnull(properties.settings.ExternalPolicyBasedAuthorization),"enabled","disabled"),
			Entra = iff(notnull(properties.settings.AzureAD),"enabled","disabled"),
			BPA = iff(notnull(properties.settings.AssessmentSettings),"enabled","disabled"),
			machineIdHasSQLServerExtensionInstalled)
		on $left.machineIdHasSQLServerDiscovered == $right.machineIdHasSQLServerExtensionInstalled
		| extend Status = iff(isnotempty(machineIdHasSQLServerExtensionInstalled), "Extension Installed", "Extension Not Installed")
		| project subscriptionId, name, machineId=machineIdHasSQLServerDiscovered, resourceGroup, location, tags, Status, 
			Provisioning_State, License_Type, ESU, Extension_Version, Excluded_instances, Purview, Entra, BPA
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

			machineId := to.String(m["machineId"])
			if filters.Azqr.IsServiceExcluded(machineId) {
				continue
			}

			subscriptionName, ok := subscriptions[subscription]
			if !ok {
				subscriptionName = ""
			}

			resources = append(resources, models.ArcSQLResult{
				SubscriptionID:    subscription,
				SubscriptionName:  subscriptionName,
				ResourceGroup:     to.String(m["resourceGroup"]),
				Location:          to.String(m["location"]),
				MachineName:       to.String(m["name"]),
				MachineID:         machineId,
				Tags:              to.String(m["tags"]),
				Status:            to.String(m["Status"]),
				ProvisioningState: to.String(m["Provisioning_State"]),
				LicenseType:       to.String(m["License_Type"]),
				ESU:               to.String(m["ESU"]),
				ExtensionVersion:  to.String(m["Extension_Version"]),
				ExcludedInstances: to.String(m["Excluded_instances"]),
				PurviewEnabled:    to.String(m["Purview"]),
				EntraEnabled:      to.String(m["Entra"]),
				BPAEnabled:        to.String(m["BPA"]),
			})
		}
	}

	return resources
}
