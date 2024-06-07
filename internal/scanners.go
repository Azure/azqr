// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/adf"
	"github.com/Azure/azqr/internal/scanners/afd"
	"github.com/Azure/azqr/internal/scanners/afw"
	"github.com/Azure/azqr/internal/scanners/agw"
	"github.com/Azure/azqr/internal/scanners/aks"
	"github.com/Azure/azqr/internal/scanners/amg"
	"github.com/Azure/azqr/internal/scanners/apim"
	"github.com/Azure/azqr/internal/scanners/appcs"
	"github.com/Azure/azqr/internal/scanners/appi"
	"github.com/Azure/azqr/internal/scanners/as"
	"github.com/Azure/azqr/internal/scanners/asp"
	"github.com/Azure/azqr/internal/scanners/ca"
	"github.com/Azure/azqr/internal/scanners/cae"
	"github.com/Azure/azqr/internal/scanners/ci"
	"github.com/Azure/azqr/internal/scanners/cog"
	"github.com/Azure/azqr/internal/scanners/cosmos"
	"github.com/Azure/azqr/internal/scanners/cr"
	"github.com/Azure/azqr/internal/scanners/dbw"
	"github.com/Azure/azqr/internal/scanners/dec"
	"github.com/Azure/azqr/internal/scanners/evgd"
	"github.com/Azure/azqr/internal/scanners/evh"
	"github.com/Azure/azqr/internal/scanners/kv"
	"github.com/Azure/azqr/internal/scanners/lb"
	"github.com/Azure/azqr/internal/scanners/logic"
	"github.com/Azure/azqr/internal/scanners/maria"
	"github.com/Azure/azqr/internal/scanners/mysql"
	"github.com/Azure/azqr/internal/scanners/psql"
	"github.com/Azure/azqr/internal/scanners/redis"
	"github.com/Azure/azqr/internal/scanners/sb"
	"github.com/Azure/azqr/internal/scanners/sigr"
	"github.com/Azure/azqr/internal/scanners/sql"
	"github.com/Azure/azqr/internal/scanners/st"
	"github.com/Azure/azqr/internal/scanners/synw"
	"github.com/Azure/azqr/internal/scanners/traf"
	"github.com/Azure/azqr/internal/scanners/vgw"
	"github.com/Azure/azqr/internal/scanners/vm"
	"github.com/Azure/azqr/internal/scanners/vmss"
	"github.com/Azure/azqr/internal/scanners/vnet"
	"github.com/Azure/azqr/internal/scanners/wps"
)

// GetScanners returns a list of all scanners
func GetScanners() []scanners.IAzureScanner {
	return []scanners.IAzureScanner{
		&dbw.DatabricksScanner{},
		&adf.DataFactoryScanner{},
		&afd.FrontDoorScanner{},
		&afw.FirewallScanner{},
		&agw.ApplicationGatewayScanner{},
		&aks.AKSScanner{},
		&amg.ManagedGrafanaScanner{},
		&apim.APIManagementScanner{},
		&appcs.AppConfigurationScanner{},
		&appi.AppInsightsScanner{},
		&as.AnalysisServicesScanner{},
		&cae.ContainerAppsEnvironmentScanner{},
		&ca.ContainerAppsScanner{},
		&ci.ContainerInstanceScanner{},
		&cog.CognitiveScanner{},
		&cosmos.CosmosDBScanner{},
		&cr.ContainerRegistryScanner{},
		&dec.DataExplorerScanner{},
		&evgd.EventGridScanner{},
		&evh.EventHubScanner{},
		&kv.KeyVaultScanner{},
		&lb.LoadBalancerScanner{},
		&logic.LogicAppScanner{},
		&maria.MariaScanner{},
		&mysql.MySQLFlexibleScanner{},
		&mysql.MySQLScanner{},
		&asp.AppServiceScanner{},
		&psql.PostgreFlexibleScanner{},
		&psql.PostgreScanner{},
		&redis.RedisScanner{},
		&sb.ServiceBusScanner{},
		&sigr.SignalRScanner{},
		&sql.SQLScanner{},
		&synw.SynapseWorkspaceScanner{},
		&traf.TrafficManagerScanner{},
		&st.StorageScanner{},
		&vm.VirtualMachineScanner{},
		&vmss.VirtualMachineScaleSetScanner{},
		&vnet.VirtualNetworkScanner{},
		&vgw.VirtualNetworkGatewayScanner{},
		&wps.WebPubSubScanner{},
	}
}
