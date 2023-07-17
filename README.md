[![build](https://github.com/Azure/azqr/actions/workflows/build.yaml/badge.svg)](https://github.com/Azure/azqr/actions/workflows/build.yaml)
[![CodeQL](https://github.com/Azure/azqr/actions/workflows/codeql.yml/badge.svg)](https://github.com/Azure/azqr/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/Azure/azqr/branch/main/graph/badge.svg?token=VReik9rs3l)](https://codecov.io/gh/Azure/azqr)
[![Average time to resolve an issue](http://isitmaintained.com/badge/resolution/Azure/azqr.svg)](http://isitmaintained.com/project/Azure/azqr "Average time to resolve an issue")
[![Percentage of issues still open](http://isitmaintained.com/badge/open/Azure/azqr.svg)](http://isitmaintained.com/project/Azure/azqr "Percentage of issues still open")

# Azure Quick Review

[![Open in vscode.dev](https://img.shields.io/badge/Open%20in-vscode.dev-blue)](https://vscode.dev/github/Azure/azqr)

Azure Quick Review (azqr) is a CLI tool used to identify Azure resources that may be or not compliant with Azure best practices and recommendations. 

Azure Quick Review (azqr) outputs includes an overview section with the following information:

* Subscription Id
* Resource Group name
* Location
* Resource Type
* Resource Name
* SKU
* SLA: SLA for the service, given the current configuration.
* AZ: True if the service is Availability Zone aware.
* PVT: True if the service has a private IP address. 
* DS: True if the service has diagnotics settings enabled.
* CAF: True if the service is compliant with the [Cloud Adoption Framework](https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations) naming covention.

Check the [Scan Results](#scan-results) documentation for more information on Azure Quick Review (azqr) outputs.

## Azure Quick Review Rules

Azure Quick Review (azqr) recommendations are based on a set of rules. To learn more check the [rules](docs/rules/README.md) documentation.

## Supported Azure Services

* Azure App Services
* Azure Functions
* Azure Container Apps
* Azure Kubernetes Service
* Azure Container Instances
* Azure Container Registry
* Azure API Management
* Azure Event Hub
* Azure Service Bus
* Azure Event Grid
* Azure SignalR Service
* Azure Web PubSub
* Azure Cache for Redis
* Azure Cosmos DB
* Azure Database for PostgreSQL Single Server
* Azure Database for PostgreSQL Flexible Server
* Azure Database for MySQL Single Server
* Azure Database for MySQL Flexible Server
* Azure SQL Database
* Azure Key Vault
* Azure App Configuration
* Azure Application Gateway
* Azure Front Door
* Azure Storage Account
* Azure Firewall
* Azure Application Insights
* Azure Virtual WAN
* Azure Load Balancer
* Azure Virtual Network

## Microsoft Defender Status

Azure Quick Review (azqr) also reports on the status of Microsoft Defender for Cloud plans.

## Usage

### Install on Linux or Azure Cloud Shell

```bash
latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-ubuntu-latest-amd64 -O azqr
chmod +x azqr
```

### Install on Mac

Download the latest release from [here](https://github.com/Azure/azqr/releases).

### Install on Windows

```console
winget install azqr
```

### 

### Authentication

**azqr** supports the following authentication methods:

* Azure CLI
* Service Principal. You'll need to set the following environment variables:
  * AZURE_CLIENT_ID
  * AZURE_CLIENT_SECRET
  * AZURE_TENANT_ID

### Authorization

**azqr** requires the following permissions:

* Subscription Reader

### Running the Scan

To scan all resource groups in all subscription run:

```bash
./azqr scan
```

To scan all resource groups in a specific subscription run:

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

### Scan Results

Azure Quick Review (azqr) creates an excel spreadsheet with the results of the scan.

> By default the Subscription Ids are masked in the spreadsheet.

Check the [Azure Quick Review Scan Results](docs/scan_results/README.md) documentation for more information.

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
