[![build](https://github.com/Azure/azqr/actions/workflows/build.yaml/badge.svg)](https://github.com/Azure/azqr/actions/workflows/build.yaml)
[![CodeQL](https://github.com/Azure/azqr/actions/workflows/codeql.yml/badge.svg)](https://github.com/Azure/azqr/actions/workflows/codeql.yml)
[![Github All Releases](https://img.shields.io/github/downloads/Azure/azqr/total.svg)]()
[![codecov](https://codecov.io/gh/Azure/azqr/branch/main/graph/badge.svg?token=VReik9rs3l)](https://codecov.io/gh/Azure/azqr)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9896/badge)](https://www.bestpractices.dev/projects/9896)
[![Average time to resolve an issue](http://isitmaintained.com/badge/resolution/Azure/azqr.svg)](http://isitmaintained.com/project/Azure/azqr "Average time to resolve an issue")
[![Percentage of issues still open](http://isitmaintained.com/badge/open/Azure/azqr.svg)](http://isitmaintained.com/project/Azure/azqr "Percentage of issues still open")

# Azure Quick Review

[![Open in vscode.dev](https://img.shields.io/badge/Open%20in-vscode.dev-blue)](https://vscode.dev/github/Azure/azqr)

![Azure Quick Review](/docs/static/logo/azqr_readme.png)

**Azure Quick Review (azqr)** is a powerful command-line interface (CLI) tool that specializes in analyzing Azure resources to ensure compliance with Azure's best practices and recommendations. Its main objective is to offer users a comprehensive overview of their Azure resources, allowing them to easily identify any non-compliant configurations or areas for improvement.

## Azure Quick Review Recommendations

**Azure Quick Review (azqr)** scans your resources with 3 types of recommendations:

* **Azure Resource Graph (ARG)** queries provided by the [Azure Proactive Resiliency Library v2 (APRL)](https://aka.ms/aprl) project
* **Azure Resource Manager (ARM)** rules built with the Azure Golang SDK
* **Azure Orphan Resources (ARG)** queries provided by the [Azure Orphan Resources](https://github.com/dolevshor/azure-orphan-resources) project

To learn more about the recommendations used by **Azure Quick Review (azqr)**, you can refer to the documentation available [here](https://azure.github.io/azqr/docs/recommendations/).

## Scan Results

The output generated by **Azure Quick Review (azqr)** is written by default to an Excel file, which contains the following sheets:

* **Recommendations**: a list with all recommendations with the number of resources that are impacted. You can use this table as an action plan to improve the compliance of your resources.
* **ImpactedResources**: a list with all resources that are impacted. You can use this table to identify resources that have issues that need to be addressed.
* **ResourceTypes**: a list of impacted resource types.
* **Inventory**: a list of all resources scanned by the tool. Here you'll find details such as SKU, Tier, Kind or calculated SLA.
* **Advisor**: a list of recommendations provided by Azure Advisor.
* **DefenderRecommendations**: a list of recommendations provided by Microsoft Defender for Cloud.
* **OutOfScope**: a list of resources that were not scanned.
* **Defender**: a list of Microsoft Defender for Cloud plans and their tiers.
* **Costs**: a list of costs associated with the scanned subscription for the last 3 months.

> By default, Azure Quick Review (azqr) obfuscates the Subscription Ids in the output to ensure the protection of sensitive information and maintain data privacy and security. If you want to display the Subscription Ids without obfuscation, you can use the `--mask=false` flag when executing the tool.

> Azure Quick Review can also generate an csv files with the same information as the excel. To generate the csv files, you can use the `--csv` flag when running the tool.

> A Power BI template is also available to help you visualize the results generated by Azure Quick Review. You can create the template running Azure Quick Review with the `pbi` command and then loading the excel file generated by the tool.

## Supported Azure Services

**Azure Quick Review (azqr)** currently supports the following Azure services:

Abbreviation  | Resource Type
---|---
aa | Microsoft.Automation/automationAccounts
adf | Microsoft.DataFactory/factories
afd | Microsoft.Cdn/profiles
afw | Microsoft.Network/azureFirewalls
afw | Microsoft.Network/ipGroups
agw | Microsoft.Network/applicationGateways
aif | Microsoft.CognitiveServices/accounts
aks | Microsoft.ContainerService/managedClusters
amg | Microsoft.Dashboard/grafana
apim | Microsoft.ApiManagement/service
appcs | Microsoft.AppConfiguration/configurationStores
appi | Microsoft.Insights/components
appi | Microsoft.Insights/activityLogAlerts
as | Microsoft.AnalysisServices/servers
asp | Microsoft.Web/serverFarms
asp | Microsoft.Web/sites
asp | Microsoft.Web/connections
asp | Microsoft.Web/certificates
avail | Microsoft.Compute/availabilitySets
avd | Specialized.Workload/AVD
avs | Microsoft.AVS/privateClouds
avs | Specialized.Workload/AVS
ba | Microsoft.Batch/batchAccounts
ca | Microsoft.App/containerApps
cae | Microsoft.App/managedenvironments
ci | Microsoft.ContainerInstance/containerGroups
con | Microsoft.Network/connections
cosmos | Microsoft.DocumentDB/databaseAccounts
cr | Microsoft.ContainerRegistry/registries
dbw | Microsoft.Databricks/workspaces
dec | Microsoft.Kusto/clusters
disk | Microsoft.Compute/disks
erc | Microsoft.Network/expressRouteCircuits
erc | Microsoft.Network/ExpressRoutePorts
evgd | Microsoft.EventGrid/domains
evh | Microsoft.EventHub/namespaces
fdfp | Microsoft.Network/frontdoorWebApplicationFirewallPolicies
gal | Microsoft.Compute/galleries
hpc | Specialized.Workload/HPC
hub | Microsoft.MachineLearningServices/workspaces
iot | Microsoft.Devices/IotHubs
it | Microsoft.VirtualMachineImages/imageTemplates
kv | Microsoft.KeyVault/vaults
lb | Microsoft.Network/loadBalancers
log | Microsoft.OperationalInsights/workspaces
logic | Microsoft.Logic/workflows
maria | Microsoft.DBforMariaDB/servers
maria | Microsoft.DBforMariaDB/servers/databases
mysql | Microsoft.DBforMySQL/servers
mysql | Microsoft.DBforMySQL/flexibleServers
netapp | Microsoft.NetApp/netAppAccounts
ng | Microsoft.Network/natGateways
nic | Microsoft.Network/networkInterfaces
nsg | Microsoft.Network/networkSecurityGroups
nw | Microsoft.Network/networkWatchers
pdnsz | Microsoft.Network/privateDnsZones
pep | Microsoft.Network/privateEndpoints
pip | Microsoft.Network/publicIPAddresses
psql | Microsoft.DBforPostgreSQL/servers
psql | Microsoft.DBforPostgreSQL/flexibleServers
redis | Microsoft.Cache/Redis
rg | Microsoft.Resources/resourceGroups
rsv | Microsoft.RecoveryServices/vaults
rt | Microsoft.Network/routeTables
sap | Specialized.Workload/SAP
sb | Microsoft.ServiceBus/namespaces
sigr | Microsoft.SignalRService/SignalR
sql | Microsoft.Sql/servers
sql | Microsoft.Sql/servers/databases
sql | Microsoft.Sql/servers/elasticPools
srch | Microsoft.Search/searchServices
st | Microsoft.Storage/storageAccounts
synw | Microsoft.Synapse/workspaces
synw | Microsoft.Synapse workspaces/bigDataPools
synw | Microsoft.Synapse/workspaces/sqlPools
traf | Microsoft.Network/trafficManagerProfiles
vdpool | Microsoft.DesktopVirtualization/hostPools
vdpool | Microsoft.DesktopVirtualization/scalingPlans
vdpool | Microsoft.DesktopVirtualization/workspaces
vgw | Microsoft.Network/virtualNetworkGateways
vm | Microsoft.Compute/virtualMachines
vmss | Microsoft.Compute/virtualMachineScaleSets
vnet | Microsoft.Network/virtualNetworks
vnet | Microsoft.Network/virtualNetworks/subnets
vwan | Microsoft.Network/virtualWans
wps | Microsoft.SignalRService/webPubSub

## Usage

### Install on Linux or Azure Cloud Shell (Bash)

```bash
latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-ubuntu-latest-amd64 -O azqr
chmod +x azqr
```

### Install on Windows

Use `winget`:

```console
winget install azqr
```

or download the executable file:

```
$latest_azqr=$(iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | convertfrom-json | Select-Object -ExpandProperty tag_name
iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-windows-latest-amd64.exe -OutFile azqr.exe
```

### Install on Mac

Use `homebrew`:

```console
brew install azqr
```

or download the latest release from [here](https://github.com/Azure/azqr/releases).

### Authentication

**Azure Quick Review (azqr)** supports the following authentication methods:

* Service Principal. You'll need to set the following environment variables:
  * AZURE_CLIENT_ID
  * AZURE_CLIENT_SECRET
  * AZURE_TENANT_ID
* Azure Managed Identity
* Azure CLI (Using this type of authentication will make scans run slower)

### Authorization

**Azure Quick Review (azqr)** requires the following permissions:

* Reader over Subscription or Management Group scope

### Running the Scan

To scan all resources in all subscription run:

```bash
./azqr scan
```

To scan all resources in a specific management group run:

```bash
./azqr scan --management-group-id <management_group_id>
```

To scan all resources in a specific subscription run:

```bash
./azqr scan -s <subscription_id>
```

To scan a specific resource group in a specific subscription run:

```bash
./azqr scan -s <subscription_id> -g <resource_group_name>
```

For information on available commands and help run:

```bash
./azqr -h
```

## Filtering Recommendations and more

You can configure Azure Quick Review to include or exclude specific subscriptions or resource groups and also exclude services or recommendations. To do so, create a `yaml` file with the following format:

```yaml
azqr:
  include:
    subscriptions:
      - <subscription_id> # format: <subscription_id>
    resourceGroups:
      - <resource_group_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>
    resourceTypes:
      - <resource type abbreviation> # format: Abbreviation of the resource type. For example: "vm" for "Microsoft.Compute/virtualMachines"
  exclude:
    subscriptions:
      - <subscription_id> # format: <subscription_id>
    resourceGroups:
      - <resource_group_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>
    services:
      - <service_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>/providers/<service_provider>/<service_name>
    recommendations:
      - <recommendation_id> # format: <recommendation_id>
```

Then run the scan with the `--filters` flag:

```bash
./azqr scan --filters <path_to_yaml_file>
```

> Check the [rules](https://azure.github.io/azqr/docs/recommendations/) to get the recommendation ids.

## Troubleshooting

If you encounter any issue while using **Azure Quick Review (azqr)**, please set the `AZURE_SDK_GO_LOGGING` environment variable to `all`, run the tool with the `--debug` flag and then share the console output with us by filing a new [issue](https://github.com/Azure/azqr/issues).

## Building Locally

Make sure you have `Go 1.23.x` or higher installed in your environment. You can set `GOROOT=<path_to_go_libexec> folder` and `GOPATH=<path_to_go_dep_folder>` if you want to be specific about where to find Go binary and Go dependencies.

```bash
   git clone git@github.com:Azure/azqr.git
   cd azqr
   git submodule init
   git submodule update --recursive
   go build -o azqr cmd/azqr/main.go
 ```

## Support

This project uses GitHub Issues to track bugs and feature requests.
Before logging an issue please check our [troubleshooting](#troubleshooting) guide.

Please search the existing issues before filing new issues to avoid duplicates.

- For new issues, file your bug or feature request as a new [issue](https://github.com/Azure/azqr/issues).
- For help, discussion, and support questions about using this project, join or start a [discussion](https://github.com/Azure/azqr/discussions).

Support for this project / product is limited to the resources listed above.

## Contributors

Thanks to everyone who has contributed!

<a href="https://github.com/Azure/azqr/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=Azure/azqr" />
</a>

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct](CODE_OF_CONDUCT.md)

## Trademark Notice

> **Trademarks** This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft trademarks or logos is subject to and must follow Microsoft’s Trademark & Brand Guidelines. Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party’s policies.
