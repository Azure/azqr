// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package registry

import (
	"github.com/Azure/azqr/internal/models"
)

type scannerSpec struct {
	name  string
	types []string
}

var registry = map[string][]scannerSpec{
	"aa":      {{"Automation Account", []string{"Microsoft.Automation/automationAccounts"}}},
	"adf":     {{"Data Factory", []string{"Microsoft.DataFactory/factories"}}},
	"afd":     {{"Front Door", []string{"Microsoft.Cdn/profiles"}}},
	"afw":     {{"Azure Firewall", []string{"Microsoft.Network/azureFirewalls", "Microsoft.Network/ipGroups"}}},
	"agw":     {{"Application Gateway", []string{"Microsoft.Network/applicationGateways"}}},
	"aif":     {{"AI Services", []string{"Microsoft.CognitiveServices/accounts"}}},
	"aks":     {{"Azure Kubernetes Service", []string{"Microsoft.ContainerService/managedClusters"}}},
	"amg":     {{"Azure Managed Grafana", []string{"Microsoft.Dashboard/grafana"}}},
	"apim":    {{"API Management", []string{"Microsoft.ApiManagement/service"}}},
	"appcs":   {{"App Configuration", []string{"Microsoft.AppConfiguration/configurationStores"}}},
	"appi":    {{"Application Insights", []string{"Microsoft.Insights/components", "Microsoft.Insights/activityLogAlerts"}}},
	"arc":     {{"Azure Arc", []string{"Microsoft.AzureArcData/sqlServerInstances"}}},
	"as":      {{"Analysis Services", []string{"Microsoft.AnalysisServices/servers"}}},
	"asa":     {{"Stream Analytics Job", []string{"Microsoft.StreamAnalytics/streamingJobs"}}},
	"asp":     {{"App Service Plan", []string{"Microsoft.Web/serverFarms", "Microsoft.Web/sites", "Microsoft.Web/connections", "Microsoft.Web/certificates"}}},
	"avail":   {{"Availability Set", []string{"Microsoft.Compute/availabilitySets"}}},
	"avd":     {{"Azure Virtual Desktop", []string{"Specialized.Workload/AVD"}}},
	"avs":     {{"Azure VMware Solution", []string{"Microsoft.AVS/privateClouds", "Specialized.Workload/AVS"}}},
	"ba":      {{"Batch Account", []string{"Microsoft.Batch/batchAccounts"}}},
	"bastion": {{"Bastion Host", []string{"Microsoft.Network/bastionHosts"}}},
	"ca":      {{"Container App", []string{"Microsoft.App/containerApps"}}},
	"cae":     {{"Container Apps Environment", []string{"Microsoft.App/managedenvironments"}}},
	"ci":      {{"Container Instance", []string{"Microsoft.ContainerInstance/containerGroups"}}},
	"con":     {{"Connection", []string{"Microsoft.Network/connections"}}},
	"cosmos":  {{"Cosmos DB", []string{"Microsoft.DocumentDB/databaseAccounts"}}},
	"cr":      {{"Container Registry", []string{"Microsoft.ContainerRegistry/registries"}}},
	"dbw":     {{"Databricks Workspace", []string{"Microsoft.Databricks/workspaces"}}},
	"ddos":    {{"DDoS Protection Plan", []string{"Microsoft.Network/ddosProtectionPlans"}}},
	"dec":     {{"Data Explorer Cluster", []string{"Microsoft.Kusto/clusters"}}},
	"disk":    {{"Disk", []string{"Microsoft.Compute/disks"}}},
	"dnsres":  {{"DNS Resolver", []string{"Microsoft.Network/dnsResolvers"}}},
	"dnsz":    {{"DNS Zone", []string{"Microsoft.Network/dnsZones"}}},
	"domain":  {{"Domain Services", []string{"Microsoft.AAD/domainServices"}}},
	"erc":     {{"ExpressRoute Circuit", []string{"Microsoft.Network/expressRouteCircuits", "Microsoft.Network/ExpressRoutePorts", "Microsoft.Network/expressRouteGateways"}}},
	"evgd":    {{"Event Grid Domain", []string{"Microsoft.EventGrid/domains"}}},
	"evgt":    {{"Event Grid Topic", []string{"Microsoft.EventGrid/topics"}}},
	"evh":     {{"Event Hub", []string{"Microsoft.EventHub/namespaces"}}},
	"fabric":  {{"Fabric", []string{"Microsoft.Fabric/capacities"}}},
	"fdfp":    {{"Front Door Firewall Policy", []string{"Microsoft.Network/frontdoorWebApplicationFirewallPolicies"}}},
	"gal":     {{"Compute Gallery", []string{"Microsoft.Compute/galleries"}}},
	"hpc":     {{"HPC", []string{"Specialized.Workload/HPC"}}},
	"hub":     {{"Machine Learning Workspace", []string{"Microsoft.MachineLearningServices/workspaces", "Microsoft.MachineLearningServices/registries"}}},
	"iot":     {{"IoT Hub", []string{"Microsoft.Devices/IotHubs"}}},
	"it":      {{"Image Template", []string{"Microsoft.VirtualMachineImages/imageTemplates"}}},
	"kv":      {{"Key Vault", []string{"Microsoft.KeyVault/vaults"}}},
	"lb":      {{"Load Balancer", []string{"Microsoft.Network/loadBalancers"}}},
	"log":     {{"Log Analytics Workspace", []string{"Microsoft.OperationalInsights/workspaces"}}},
	"logic":   {{"Logic App", []string{"Microsoft.Logic/workflows"}}},
	"mysql":   {{"MySQL Database", []string{"Microsoft.DBforMySQL/servers", "Microsoft.DBforMySQL/flexibleServers"}}},
	"netapp":  {{"NetApp Account", []string{"Microsoft.NetApp/netAppAccounts"}}},
	"ng":      {{"NAT Gateway", []string{"Microsoft.Network/natGateways"}}},
	"nic":     {{"Network Interface", []string{"Microsoft.Network/networkInterfaces"}}},
	"nsg":     {{"Network Security Group", []string{"Microsoft.Network/networkSecurityGroups"}}},
	"ntc":     {{"Azure Traffic Collector", []string{"Microsoft.NetworkFunction/azureTrafficCollectors"}}},
	"nw":      {{"Network Watcher", []string{"Microsoft.Network/networkWatchers"}}},
	"odb":     {{"Oracle Database", []string{"Oracle.Database/cloudExadataInfrastructures", "Oracle.Database/cloudVmClusters"}}},
	"p2svpng": {{"P2S VPN Gateway", []string{"Microsoft.Network/p2sVpnGateways"}}},
	"pdnsz":   {{"Private DNS Zone", []string{"Microsoft.Network/privateDnsZones"}}},
	"pep":     {{"Private Endpoint", []string{"Microsoft.Network/privateEndpoints"}}},
	"pip":     {{"Public IP Address", []string{"Microsoft.Network/publicIPAddresses"}}},
	"psql":    {{"PostgreSQL Database", []string{"Microsoft.DBforPostgreSQL/servers", "Microsoft.DBforPostgreSQL/flexibleServers"}}},
	"redis": {
		{"Redis Cache", []string{"Microsoft.Cache/Redis"}},
		{"Redis Enterprise", []string{"Microsoft.Cache/redisEnterprise"}},
	},
	"resource": {{"Resource", []string{"Microsoft.Resources"}}},
	"rg":       {{"Resource Group", []string{"Microsoft.Resources/resourceGroups"}}},
	"rsv":      {{"Recovery Services Vault", []string{"Microsoft.RecoveryServices/vaults"}}},
	"rt":       {{"Route Table", []string{"Microsoft.Network/routeTables"}}},
	"sap":      {{"SAP", []string{"Specialized.Workload/SAP"}}},
	"sb":       {{"Service Bus", []string{"Microsoft.ServiceBus/namespaces"}}},
	"sigr":     {{"SignalR", []string{"Microsoft.SignalRService/SignalR"}}},
	"sql":      {{"SQL Server", []string{"Microsoft.Sql/servers", "Microsoft.Sql/servers/databases", "Microsoft.Sql/servers/elasticPools"}}},
	"sqlmi":    {{"SQL Managed Instance", []string{"Microsoft.Sql/managedInstances"}}},
	"srch":     {{"Search Service", []string{"Microsoft.Search/searchServices"}}},
	"st":       {{"Storage Account", []string{"Microsoft.Storage/storageAccounts"}}},
	"sub":      {{"Subscription", []string{"Microsoft.Subscription/subscriptions"}}},
	"synw":     {{"Synapse Workspace", []string{"Microsoft.Synapse/workspaces", "Microsoft.Synapse/workspaces/bigDataPools", "Microsoft.Synapse/workspaces/sqlPools"}}},
	"traf":     {{"Traffic Manager", []string{"Microsoft.Network/trafficManagerProfiles"}}},
	"vdpool":   {{"Virtual Desktop Host Pool", []string{"Microsoft.DesktopVirtualization/hostPools", "Microsoft.DesktopVirtualization/scalingPlans", "Microsoft.DesktopVirtualization/workspaces"}}},
	"vgw":      {{"Virtual Network Gateway", []string{"Microsoft.Network/virtualNetworkGateways"}}},
	"vhub":     {{"Virtual Hub", []string{"Microsoft.Network/virtualHubs"}}},
	"vm":       {{"Virtual Machine", []string{"Microsoft.Compute/virtualMachines"}}},
	"vmss":     {{"Virtual Machine Scale Set", []string{"Microsoft.Compute/virtualMachineScaleSets"}}},
	"vnet":     {{"Virtual Network", []string{"Microsoft.Network/virtualNetworks", "Microsoft.Network/virtualNetworks/subnets"}}},
	"vpng":     {{"VPN Gateway", []string{"Microsoft.Network/vpnGateways"}}},
	"vpns":     {{"VPN Site", []string{"Microsoft.Network/vpnSites"}}},
	"vrouter":  {{"Virtual Router", []string{"Microsoft.Network/virtualRouters"}}},
	"vwan":     {{"Virtual WAN", []string{"Microsoft.Network/virtualWans"}}},
	"wps":      {{"Web PubSub", []string{"Microsoft.SignalRService/webPubSub"}}},
}

func init() {
	for key, specs := range registry {
		list := make([]models.IAzureScanner, 0, len(specs))
		for _, s := range specs {
			list = append(list, models.NewBaseScanner(s.name, s.types...))
		}
		models.ScannerList[key] = list
	}
}
