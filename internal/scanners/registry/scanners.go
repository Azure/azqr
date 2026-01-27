// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package registry

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["aa"] = []models.IAzureScanner{
		models.NewBaseScanner("Automation Account", "Microsoft.Automation/automationAccounts"),
	}
}

func init() {
	models.ScannerList["adf"] = []models.IAzureScanner{
		models.NewBaseScanner("Data Factory", "Microsoft.DataFactory/factories"),
	}
}

func init() {
	models.ScannerList["afd"] = []models.IAzureScanner{
		models.NewBaseScanner("Front Door", "Microsoft.Cdn/profiles"),
	}
}

func init() {
	models.ScannerList["afw"] = []models.IAzureScanner{
		models.NewBaseScanner("Azure Firewall", "Microsoft.Network/azureFirewalls", "Microsoft.Network/ipGroups"),
	}
}

func init() {
	models.ScannerList["agw"] = []models.IAzureScanner{
		models.NewBaseScanner("Application Gateway", "Microsoft.Network/applicationGateways"),
	}
}

func init() {
	models.ScannerList["aif"] = []models.IAzureScanner{
		models.NewBaseScanner("AI Services", "Microsoft.CognitiveServices/accounts"),
	}
}

func init() {
	models.ScannerList["aks"] = []models.IAzureScanner{
		models.NewBaseScanner("Azure Kubernetes Service", "Microsoft.ContainerService/managedClusters"),
	}
}

func init() {
	models.ScannerList["amg"] = []models.IAzureScanner{
		models.NewBaseScanner("Azure Managed Grafana", "Microsoft.Dashboard/grafana"),
	}
}

func init() {
	models.ScannerList["apim"] = []models.IAzureScanner{
		models.NewBaseScanner("API Management", "Microsoft.ApiManagement/service"),
	}
}

func init() {
	models.ScannerList["appcs"] = []models.IAzureScanner{
		models.NewBaseScanner("App Configuration", "Microsoft.AppConfiguration/configurationStores"),
	}
}

func init() {
	models.ScannerList["appi"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Application Insights",
			"Microsoft.Insights/components",
			"Microsoft.Insights/activityLogAlerts",
		),
	}
}

func init() {
	models.ScannerList["arc"] = []models.IAzureScanner{
		models.NewBaseScanner("Azure Arc", "Microsoft.AzureArcData/sqlServerInstances"),
	}
}

func init() {
	models.ScannerList["as"] = []models.IAzureScanner{
		models.NewBaseScanner("Analysis Services", "Microsoft.AnalysisServices/servers"),
	}
}

func init() {
	models.ScannerList["asp"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"App Service Plan",
			"Microsoft.Web/serverFarms",
			"Microsoft.Web/sites",
			"Microsoft.Web/connections",
			"Microsoft.Web/certificates",
		),
	}
}

func init() {
	models.ScannerList["avail"] = []models.IAzureScanner{
		models.NewBaseScanner("Availability Set", "Microsoft.Compute/availabilitySets"),
	}
}

func init() {
	models.ScannerList["avd"] = []models.IAzureScanner{
		models.NewBaseScanner("Azure Virtual Desktop", "Specialized.Workload/AVD"),
	}
}

func init() {
	models.ScannerList["avs"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Azure VMware Solution",
			"Microsoft.AVS/privateClouds",
			"Specialized.Workload/AVS",
		),
	}
}

func init() {
	models.ScannerList["ba"] = []models.IAzureScanner{
		models.NewBaseScanner("Batch Account", "Microsoft.Batch/batchAccounts"),
	}
}

func init() {
	models.ScannerList["ca"] = []models.IAzureScanner{
		models.NewBaseScanner("Container App", "Microsoft.App/containerApps"),
	}
}

func init() {
	models.ScannerList["cae"] = []models.IAzureScanner{
		models.NewBaseScanner("Container Apps Environment", "Microsoft.App/managedenvironments"),
	}
}

func init() {
	models.ScannerList["ci"] = []models.IAzureScanner{
		models.NewBaseScanner("Container Instance", "Microsoft.ContainerInstance/containerGroups"),
	}
}

func init() {
	models.ScannerList["con"] = []models.IAzureScanner{
		models.NewBaseScanner("Connection", "Microsoft.Network/connections"),
	}
}

func init() {
	models.ScannerList["cosmos"] = []models.IAzureScanner{
		models.NewBaseScanner("Cosmos DB", "Microsoft.DocumentDB/databaseAccounts"),
	}
}

func init() {
	models.ScannerList["cr"] = []models.IAzureScanner{
		models.NewBaseScanner("Container Registry", "Microsoft.ContainerRegistry/registries"),
	}
}

func init() {
	models.ScannerList["dbw"] = []models.IAzureScanner{
		models.NewBaseScanner("Databricks Workspace", "Microsoft.Databricks/workspaces"),
	}
}

func init() {
	models.ScannerList["dec"] = []models.IAzureScanner{
		models.NewBaseScanner("Data Explorer Cluster", "Microsoft.Kusto/clusters"),
	}
}

func init() {
	models.ScannerList["disk"] = []models.IAzureScanner{
		models.NewBaseScanner("Disk", "Microsoft.Compute/disks"),
	}
}

func init() {
	models.ScannerList["domain"] = []models.IAzureScanner{
		models.NewBaseScanner("Domain Services", "Microsoft.AAD/domainServices"),
	}
}

func init() {
	models.ScannerList["erc"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"ExpressRoute Circuit",
			"Microsoft.Network/expressRouteCircuits",
			"Microsoft.Network/ExpressRoutePorts",
			"Microsoft.Network/expressRouteGateways",
		),
	}
}

func init() {
	models.ScannerList["evgd"] = []models.IAzureScanner{
		models.NewBaseScanner("Event Grid Domain", "Microsoft.EventGrid/domains"),
	}
}

func init() {
	models.ScannerList["evgt"] = []models.IAzureScanner{
		models.NewBaseScanner("Event Grid Topic", "Microsoft.EventGrid/topics"),
	}
}

func init() {
	models.ScannerList["evh"] = []models.IAzureScanner{
		models.NewBaseScanner("Event Hub", "Microsoft.EventHub/namespaces"),
	}
}

func init() {
	models.ScannerList["fabric"] = []models.IAzureScanner{
		models.NewBaseScanner("Fabric", "Microsoft.Fabric/capacities"),
	}
}

func init() {
	models.ScannerList["fdfp"] = []models.IAzureScanner{
		models.NewBaseScanner("Front Door Firewall Policy", "Microsoft.Network/frontdoorWebApplicationFirewallPolicies"),
	}
}

func init() {
	models.ScannerList["gal"] = []models.IAzureScanner{
		models.NewBaseScanner("Compute Gallery", "Microsoft.Compute/galleries"),
	}
}

func init() {
	models.ScannerList["hpc"] = []models.IAzureScanner{
		models.NewBaseScanner("HPC", "Specialized.Workload/HPC"),
	}
}

func init() {
	models.ScannerList["hub"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Machine Learning Workspace",
			"Microsoft.MachineLearningServices/workspaces",
			"Microsoft.MachineLearningServices/registries",
		),
	}
}

func init() {
	models.ScannerList["iot"] = []models.IAzureScanner{
		models.NewBaseScanner("IoT Hub", "Microsoft.Devices/IotHubs"),
	}
}

func init() {
	models.ScannerList["it"] = []models.IAzureScanner{
		models.NewBaseScanner("Image Template", "Microsoft.VirtualMachineImages/imageTemplates"),
	}
}

func init() {
	models.ScannerList["kv"] = []models.IAzureScanner{
		models.NewBaseScanner("Key Vault", "Microsoft.KeyVault/vaults"),
	}
}

func init() {
	models.ScannerList["lb"] = []models.IAzureScanner{
		models.NewBaseScanner("Load Balancer", "Microsoft.Network/loadBalancers"),
	}
}

func init() {
	models.ScannerList["log"] = []models.IAzureScanner{
		models.NewBaseScanner("Log Analytics Workspace", "Microsoft.OperationalInsights/workspaces"),
	}
}

func init() {
	models.ScannerList["logic"] = []models.IAzureScanner{
		models.NewBaseScanner("Logic App", "Microsoft.Logic/workflows"),
	}
}

func init() {
	models.ScannerList["mysql"] = []models.IAzureScanner{
		models.NewBaseScanner("MySQL Database", "Microsoft.DBforMySQL/servers", "Microsoft.DBforMySQL/flexibleServers"),
	}
}

func init() {
	models.ScannerList["netapp"] = []models.IAzureScanner{
		models.NewBaseScanner("NetApp Account", "Microsoft.NetApp/netAppAccounts"),
	}
}

func init() {
	models.ScannerList["ng"] = []models.IAzureScanner{
		models.NewBaseScanner("NAT Gateway", "Microsoft.Network/natGateways"),
	}
}

func init() {
	models.ScannerList["nic"] = []models.IAzureScanner{
		models.NewBaseScanner("Network Interface", "Microsoft.Network/networkInterfaces"),
	}
}

func init() {
	models.ScannerList["nsg"] = []models.IAzureScanner{
		models.NewBaseScanner("Network Security Group", "Microsoft.Network/networkSecurityGroups"),
	}
}

func init() {
	models.ScannerList["bastion"] = []models.IAzureScanner{
		models.NewBaseScanner("Bastion Host", "Microsoft.Network/bastionHosts"),
	}
}

func init() {
	models.ScannerList["ddos"] = []models.IAzureScanner{
		models.NewBaseScanner("DDoS Protection Plan", "Microsoft.Network/ddosProtectionPlans"),
	}
}

func init() {
	models.ScannerList["dnsres"] = []models.IAzureScanner{
		models.NewBaseScanner("DNS Resolver", "Microsoft.Network/dnsResolvers"),
	}
}

func init() {
	models.ScannerList["dnsz"] = []models.IAzureScanner{
		models.NewBaseScanner("DNS Zone", "Microsoft.Network/dnsZones"),
	}
}

func init() {
	models.ScannerList["nw"] = []models.IAzureScanner{
		models.NewBaseScanner("Network Watcher", "Microsoft.Network/networkWatchers"),
	}
}

func init() {
	models.ScannerList["ntc"] = []models.IAzureScanner{
		models.NewBaseScanner("Azure Traffic Collector", "Microsoft.NetworkFunction/azureTrafficCollectors"),
	}
}

func init() {
	models.ScannerList["odb"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Oracle Database",
			"Oracle.Database/cloudExadataInfrastructures",
			"Oracle.Database/cloudVmClusters",
		),
	}
}

func init() {
	models.ScannerList["p2svpng"] = []models.IAzureScanner{
		models.NewBaseScanner("P2S VPN Gateway", "Microsoft.Network/p2sVpnGateways"),
	}
}

func init() {
	models.ScannerList["pdnsz"] = []models.IAzureScanner{
		models.NewBaseScanner("Private DNS Zone", "Microsoft.Network/privateDnsZones"),
	}
}

func init() {
	models.ScannerList["pep"] = []models.IAzureScanner{
		models.NewBaseScanner("Private Endpoint", "Microsoft.Network/privateEndpoints"),
	}
}

func init() {
	models.ScannerList["pip"] = []models.IAzureScanner{
		models.NewBaseScanner("Public IP Address", "Microsoft.Network/publicIPAddresses"),
	}
}

func init() {
	models.ScannerList["psql"] = []models.IAzureScanner{
		models.NewBaseScanner("PostgreSQL Database", "Microsoft.DBforPostgreSQL/servers", "Microsoft.DBforPostgreSQL/flexibleServers"),
	}
}

func init() {
	models.ScannerList["redis"] = []models.IAzureScanner{
		models.NewBaseScanner("Redis Cache", "Microsoft.Cache/Redis"),
	}
}

func init() {
	models.ScannerList["resource"] = []models.IAzureScanner{
		models.NewBaseScanner("Resource", "Microsoft.Resources"),
	}
}

func init() {
	models.ScannerList["rg"] = []models.IAzureScanner{
		models.NewBaseScanner("Resource Group", "Microsoft.Resources/resourceGroups"),
	}
}

func init() {
	models.ScannerList["rsv"] = []models.IAzureScanner{
		models.NewBaseScanner("Recovery Services Vault", "Microsoft.RecoveryServices/vaults"),
	}
}

func init() {
	models.ScannerList["rt"] = []models.IAzureScanner{
		models.NewBaseScanner("Route Table", "Microsoft.Network/routeTables"),
	}
}

func init() {
	models.ScannerList["sap"] = []models.IAzureScanner{
		models.NewBaseScanner("SAP", "Specialized.Workload/SAP"),
	}
}

func init() {
	models.ScannerList["sb"] = []models.IAzureScanner{
		models.NewBaseScanner("Service Bus", "Microsoft.ServiceBus/namespaces"),
	}
}

func init() {
	models.ScannerList["sigr"] = []models.IAzureScanner{
		models.NewBaseScanner("SignalR", "Microsoft.SignalRService/SignalR"),
	}
}

func init() {
	models.ScannerList["sqlmi"] = []models.IAzureScanner{
		models.NewBaseScanner("SQL Managed Instance", "Microsoft.Sql/managedInstances"),
	}
}

func init() {
	models.ScannerList["sql"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"SQL Server",
			"Microsoft.Sql/servers",
			"Microsoft.Sql/servers/databases",
			"Microsoft.Sql/servers/elasticPools",
		)}
}

func init() {
	models.ScannerList["srch"] = []models.IAzureScanner{
		models.NewBaseScanner("Search Service", "Microsoft.Search/searchServices"),
	}
}

func init() {
	models.ScannerList["st"] = []models.IAzureScanner{
		models.NewBaseScanner("Storage Account", "Microsoft.Storage/storageAccounts"),
	}
}

func init() {
	models.ScannerList["asa"] = []models.IAzureScanner{
		models.NewBaseScanner("Stream Analytics Job", "Microsoft.StreamAnalytics/streamingJobs"),
	}
}

func init() {
	models.ScannerList["sub"] = []models.IAzureScanner{
		models.NewBaseScanner("Subscription", "Microsoft.Subscription/subscriptions"),
	}
}

func init() {
	models.ScannerList["synw"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Synapse Workspace",
			"Microsoft.Synapse/workspaces",
			"Microsoft.Synapse/workspaces/bigDataPools",
			"Microsoft.Synapse/workspaces/sqlPools",
		)}
}

func init() {
	models.ScannerList["traf"] = []models.IAzureScanner{
		models.NewBaseScanner("Traffic Manager", "Microsoft.Network/trafficManagerProfiles"),
	}
}

func init() {
	models.ScannerList["vdpool"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Virtual Desktop Host Pool",
			"Microsoft.DesktopVirtualization/hostPools",
			"Microsoft.DesktopVirtualization/scalingPlans",
			"Microsoft.DesktopVirtualization/workspaces",
		),
	}
}

func init() {
	models.ScannerList["vgw"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual Network Gateway", "Microsoft.Network/virtualNetworkGateways"),
	}
}

func init() {
	models.ScannerList["vhub"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual Hub", "Microsoft.Network/virtualHubs"),
	}
}

func init() {
	models.ScannerList["vrouter"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual Router", "Microsoft.Network/virtualRouters"),
	}
}

func init() {
	models.ScannerList["vpng"] = []models.IAzureScanner{
		models.NewBaseScanner("VPN Gateway", "Microsoft.Network/vpnGateways"),
	}
}

func init() {
	models.ScannerList["vpns"] = []models.IAzureScanner{
		models.NewBaseScanner("VPN Site", "Microsoft.Network/vpnSites"),
	}
}

func init() {
	models.ScannerList["vm"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual Machine", "Microsoft.Compute/virtualMachines"),
	}
}

func init() {
	models.ScannerList["vmss"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual Machine Scale Set", "Microsoft.Compute/virtualMachineScaleSets"),
	}
}

func init() {
	models.ScannerList["vnet"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual Network", "Microsoft.Network/virtualNetworks", "Microsoft.Network/virtualNetworks/subnets"),
	}
}

func init() {
	models.ScannerList["vwan"] = []models.IAzureScanner{
		models.NewBaseScanner("Virtual WAN", "Microsoft.Network/virtualWans"),
	}
}

func init() {
	models.ScannerList["wps"] = []models.IAzureScanner{
		models.NewBaseScanner("Web PubSub", "Microsoft.SignalRService/webPubSub"),
	}
}
