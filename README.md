[![build](https://github.com/cmendible/azqr/actions/workflows/build.yaml/badge.svg)](https://github.com/cmendible/azqr/actions/workflows/build.yaml)

# Azure Quick Review

Azure Quick Review (azqr) goal is to produce a high level assemesment of an Azure Subscription or Resource Group providing the following information for each Azure Service:

* SLA: current expected SLA
* Availability Zones: checks if the service is protected against Zone failures. 
* Private Endpoints: checks if the service uses Private Endpoints.
* Diagnostic Settings: checks if there are Diagnostic Settings configured for the service. 
* CAF Naming convention: checks if the service follows [CAF Naming convention](https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations).

## Usage

Download the latest release from [here](https://github.com/cmendible/azqr/releases).

### Authentication

**azqr** supports the following authentication methods:

* Azure CLI
* Service Principal. You'll need to set the following environment variables:
  * AZURE_CLIENT_ID
  * AZURE_CLIENT_SECRET
  * AZURE_TENANT_ID

### Running the Review

To review all resource groups in a specific subscription run:

```bash
./azqr -s <subscription_id>
```

To review a specific resource group in a specific subscription run:

```bash
./azqr -s <subscription_id> -r <resource_group_name>
```

For help run:

```bash
./azqr -h
```

## Contribution

Thanks to everyone who has contributed!

<a href="https://github.com/cmendible/azqr/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=cmendible/azqr" />
</a>

