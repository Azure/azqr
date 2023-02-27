[![build](https://github.com/cmendible/azqr/actions/workflows/build.yaml/badge.svg)](https://github.com/cmendible/azqr/actions/workflows/build.yaml)
[![codecov](https://codecov.io/gh/cmendible/azqr/branch/main/graph/badge.svg?token=VReik9rs3l)](https://codecov.io/gh/cmendible/azqr)

# Azure Quick Review

Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group providing the following information for each Azure Service:

* SLA: current expected SLA
* Availability Zones: checks if the service is protected against Zone failures. 
* Private Endpoints: checks if the service uses Private Endpoints.
* Diagnostic Settings: checks if there are Diagnostic Settings configured for the service. 
* CAF Naming convention: checks if the service follows [CAF Naming convention](https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations).

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
* Azure SQL Database
* Azure Key Vault
* Azure App Configuration
* Azure Application Gateway
* Azure Front Door
* Azure Storage Account

## Microsoft Defender Status

Azure Quick Review (azqr) also reports on the status of Microsoft Defender for Cloud plans.

## Usage

### Install on Linux or Azure Cloud Shell

```bash
latest_azqr=$(curl -sL https://api.github.com/repos/cmendible/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/cmendible/azqr/releases/download/$latest_azqr/azqr-ubuntu-latest-amd64 -O azqr
chmod +x azqr
```

### Install on Windows or Mac

Download the latest release from [here](https://github.com/cmendible/azqr/releases).

### Authentication

**azqr** supports the following authentication methods:

* Azure CLI
* Service Principal. You'll need to set the following environment variables:
  * AZURE_CLIENT_ID
  * AZURE_CLIENT_SECRET
  * AZURE_TENANT_ID

### Running the Scan

To scan all resource groups in a specific subscription run:

```bash
./azqr -s <subscription_id>
```

To scan a specific resource group in a specific subscription run:

```bash
./azqr -s <subscription_id> -g <resource_group_name>
```

For help run:

```bash
./azqr -h
```

### Scan Results

Azure Quick Review (azqr) creates an excel spreadsheet with the results of the scan.

Check the [Azure Quick Review Scan Results](docs/scan_results/README.md) documentation for more information.

## Troubleshooting

### Error: "RESPONSE 429: 429 Too Many Requests"

If the output of `azqr` shows an error similar to the following:

```bash
--------------------------------------------------------------------------------
RESPONSE 429: 429 Too Many Requests
ERROR CODE: ResourceRequestsThrottled
--------------------------------------------------------------------------------
{
  "error": {
    "code": "ResourceRequestsThrottled",
    "message": "Number of requests for action 'Microsoft.Cdn/profiles/read' exceeded the limit of '50' for time interval '00:05:00'. Please try again after '372' seconds."
  }
}
```

Reduce the number of parallel requests that `azqr` is making. You can do this by setting the value of the `-p` parameter to a lower value (default is 4) as follows:

```bash
./azqr -s <subscription_id> -p 2
```

## Contributors

Thanks to everyone who has contributed!

<a href="https://github.com/cmendible/azqr/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=cmendible/azqr" />
</a>

## Code of Conduct

This project has adopted the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md)
