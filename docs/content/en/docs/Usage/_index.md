---
title: Usage
description: Use Azure Quick Review &mdash; to analyze Azure resources and identify whether they comply with Azure's best practices and recommendations.
weight: 3
---

## Authentication

**Azure Quick Review (azqr)** requires the following permissions:

* Reader over Subscription or Management Group scope

### - PowerShell

> Install Azure PowerShell Modules

```powershell
Install-Module -Name 'Az' --Scope 'CurrentUser'
```

> Create Service Principal

``` powershell
$spDetails = New-AzADServicePrincipal -DisplayName 'sp-azure-quick-review'
```

From `$spDetails`, Extract the

- AppId..........: `$spDetails.appId`
- AppSecret..: `$spDetails.PasswordCredentials.SecretText`
- TenantId....: `(Get-AzContext).Tenant.Id`

> Authenticate to Azure

``` powershell
$env:AZURE_CLIENT_ID = ''
$env:AZURE_CLIENT_SECRET = ''
$env:AZURE_TENANT_ID = ''
```

> Execute Azure Quick Review Scan

``` console
azqr scan
```

### - Azure CLI

> Install Microsoft Azure CLI

```console
winget install -e 'Microsoft.AzureCLI'
```

> Authenticate to Azure

```console
az login
```

> Execute Azure Quick Review Scan

```console
azqr scan --azure-cli-credential
```

## Specific Resources or Subscriptions

Do you need to scan just one subscription or a specific resource group, AzQR can do this!

> Subscription

```console
azqr scan --subscription-id ''
```

> Resource Group

```console
azqr scan --subscription-id '' --resource-group ''
```

## File Outputs

Currently Azure Quick Review supports 3 types of file outputs: `xlsx` (default), `csv`, `json`

### - xlsx

```powershell
$date = Get-Date -Format 'yyyy-MM-dd'
$fileName = "$($date)_azqr_report"
azqr scan --output-name $fileName
```

### - csv

```powershell
$date = Get-Date -Format 'yyyy-MM-dd'
$fileName = "$($date)_azqr_report"
azqr scan --csv --output-name $fileName
```

When the scan and export is completed you will be presented with 9 csv files:

```
<file-name>.advisor.csv
<file-name>.costs.csv
<file-name>.defender.csv
<file-name>.defenderRecommendations.csv
<file-name>.impacted.csv
<file-name>.inventory.csv
<file-name>.outofscope.csv
<file-name>.recommendations.csv
<file-name>.resourceType.csv
```

### - json
By default AzQR will create an xlsx document, However if you need to export to json you can use the following command:

``` powershell
$date = Get-Date -Format 'yyyy-MM-dd'
$fileName = "$($date)_azqr_report"
azqr scan --output-name $fileName --json
```

When the scan and export is completed you will be presented with 9 json files:

```
<file-name>.advisor.json
<file-name>.costs.json
<file-name>.defender.json
<file-name>.defenderRecommendations.json
<file-name>.impacted.json
<file-name>.inventory.json
<file-name>.outofscope.json
<file-name>.recommendations.json
<file-name>.resourceType.json
```

## Help

<details>
<summary>azqr --help</summary>
<br>

Command:

```console
azqr --help
```
Help Response:

```
Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group

Usage:
  azqr [flags]
  azqr [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  pbi         Creates Power BI Desktop dashboard template
  rules       Print all recommendations
  scan        Scan Azure Resources
  types       Print all supported azure resource types

Flags:
  -h, --help      help for azqr
  -v, --version   version for azqr

Use "azqr [command] --help" for more information about a command.
```

</details>
<br>

<details>
<summary>azqr completion --help</summary>
<br>

Command:

```console
azqr completion --help
```

Help Response:

```
Generate the autocompletion script for azqr for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  azqr completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion

Use "azqr completion [command] --help" for more information about a command.
```
</details>
<br>

<details>
<summary>azqr pbi --help</summary>
<br>

Command:

```console
azqr pbi --help
```

Help Response:

```
Creates Power BI Desktop dashboard template

Usage:
  azqr pbi [flags]

Flags:
  -h, --help                   help for pbi
  -p, --template-path string   Path were the PowerBI template will be created
```

</details>
<br>

<details>
<summary>azqr rules --help</summary>
<br>

Command:

```console
azqr rules --help
```

Help Response:

```
Print all recommendations as markdown table

Usage:
  azqr rules [flags]

Flags:
  -h, --help   help for rules
```

</details>
<br>

<details>
<summary>azqr types --help</summary>
<br>

Command:

```console
azqr types --help
```

Help Response:

```
Print all supported azure resource types

Usage:
  azqr types [flags]

Flags:
  -h, --help   help for types
```

</details>
<br>

<details>
<summary>azqr scan --help</summary>
<br>

Command:

```console
azqr scan --help
```

Help Response:

```
Scan Azure Resources

Usage:
  azqr scan [flags]
  azqr scan [command]

Available Commands:
  aa          Scan Azure Automation Account
  adf         Scan Azure Data Factory
  afd         Scan Azure Front Door
  afw         Scan Azure Firewall
  agw         Scan Azure Application Gateway
  aks         Scan Azure Kubernetes Service
  amg         Scan Azure Managed Grafana
  apim        Scan Azure API Management
  appcs       Scan Azure App Configuration
  appi        Scan Azure Application Insights
  as          Scan Azure Analysis Service
  asp         Scan Azure App Service
  avail       Scan Availability Sets
  avd         Scan Azure Virtual Desktop
  avs         Scan Azure VMware Solution
  ba          Scan Azure Batch Account
  ca          Scan Azure Container Apps
  cae         Scan Azure Container Apps Environment
  ci          Scan Azure Container Instances
  cog         Scan Azure Cognitive Service Accounts
  con         Scan Connection
  cosmos      Scan Azure Cosmos DB
  cr          Scan Azure Container Registries
  dbw         Scan Azure Databricks
  dec         Scan Azure Data Explorer
  disk        Scan Disk
  erc         Scan Express Route Circuits
  evgd        Scan Azure Event Grid Domains
  evh         Scan Azure Event Hubs
  fdfp        Scan Front Door Web Application Policy
  gal         Scan Azure Galleries
  hpc         Scan HPC
  iot         Scan Azure IoT Hub
  it          Scan Image Template
  kv          Scan Azure Key Vault
  lb          Scan Azure Load Balancer
  log         Scan Log Analytics workspace
  logic       Scan Azure Logic Apps
  maria       Scan Azure Database for MariaDB
  mysql       Scan Azure Database for MySQL
  netapp      Scan NetApp
  ng          Scan Azure NAT Gateway
  nic         Scan NICs
  nsg         Scan NSG
  nw          Scan Network Watcher
  pdnsz       Scan Private DNS Zone
  pep         Scan Private Endpoint
  pip         Scan Public IP
  psql        Scan Azure Database for psql
  redis       Scan Azure Cache for Redis
  rg          Scan Resource Groups
  rsv         Scan Recovery Service
  rt          Scan Route Table
  sap         Scan SAP
  sb          Scan Azure Service Bus
  sigr        Scan Azure SignalR
  sql         Scan Azure SQL Database
  st          Scan Azure Storage
  synw        Scan Azure Synapse Workspace
  traf        Scan Azure Traffic Manager
  vdpool      Scan Azure Virtual Desktop
  vgw         Scan Virtual Network Gateway
  vm          Scan Virtual Machine
  vmss        Scan Virtual Machine Scale Set
  vnet        Scan Azure Virtual Network
  vwan        Scan Azure Virtual WAN
  wps         Scan Azure Web PubSub

Flags:
  -a, --advisor                      Scan Azure Advisor Recommendations (default) (default true)
      --azqr                         Scan Azure Quick Review Recommendations (default) (default true)
  -f, --azure-cli-credential         Force the use of Azure CLI Credential
  -c, --costs                        Scan Azure Costs (default) (default true)
      --csv                          Create csv files
      --debug                        Set log level to debug
  -d, --defender                     Scan Defender Status (default) (default true)
  -e, --filters string               Filters file (YAML format)
  -h, --help                         help for scan
      --json                         Create json file
      --management-group-id string   Azure Management Group Id
  -m, --mask                         Mask the subscription id in the report (default) (default true)
  -o, --output-name string           Output file name without extension
  -g, --resource-group string        Azure Resource Group (Use with --subscription-id)
  -s, --subscription-id string       Azure Subscription Id

Use "azqr scan [command] --help" for more information about a command.
```

</details>
<br>

> Check the [rules](https://azure.github.io/azqr/docs/recommendations/) to get the recommendation ids.
