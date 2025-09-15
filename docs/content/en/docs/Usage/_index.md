---
title: Usage
description: Use Azure Quick Review &mdash; to analyze Azure resources and identify whether they comply with Azure's best practices and recommendations.
weight: 3
---

## Authorization

**Azure Quick Review (azqr)** requires the following permissions:

* Reader over Subscription or Management Group scope

## Authentication

**Azure Quick Review (azqr)** requires the following permissions:

* Reader over Subscription or Management Group scope

### Credential Chain Configuration

**Azure Quick Review (azqr)** uses the Azure SDK's `DefaultAzureCredential` which automatically selects the most appropriate credential based on your environment. You can customize the credential chain behavior by setting the `AZURE_TOKEN_CREDENTIALS` environment variable.

**Development environments:**
Set `AZURE_TOKEN_CREDENTIALS=dev` to use Azure CLI (`az`) or Azure Developer CLI (`azd`) credentials.

**Production environments:** 
Set `AZURE_TOKEN_CREDENTIALS=prod` to use environment variables, workload identity, or managed identity credentials.

### Service Principal Authentication

Set the following environment variables:

Powershell:

``` powershell
$env:AZURE_CLIENT_ID = '<service-principal-client-id>'
$env:AZURE_CLIENT_SECRET = '<service-principal-client-secret>'
$env:AZURE_TENANT_ID = '<tenant-id>'
```

Bash:

``` bash
export AZURE_CLIENT_ID='<service-principal-client-id>'
export AZURE_CLIENT_SECRET='<service-principal-client-secret>'
export AZURE_TENANT_ID='<tenant-id>'
```

### Authenticate with a Managed Identity

Set the following environment variables:

Powershell:

``` powershell
$env:AZURE_CLIENT_ID = '<managed-identity-client-id>'
$env:AZURE_TENANT_ID = '<tenant-id>'
```

Bash:

``` bash
export AZURE_CLIENT_ID='<managed-identity-client-id>'
export AZURE_TENANT_ID='<tenant-id>'
```

### Authenticate with Azure CLI

Authenticate to Azure:

```console
az login
```

## Scan Azure Resources

* Scan All Resources

  ```console
  azqr scan
  ```

* Scan a Management Group

  ```console
  azqr scan --management-group-id <management_group_id>
  ```

* Scan a Subscription
  
  ```console
  azqr scan --subscription-id <subscription_id>
  ```

* Scan a Resource Group

  ```console
  azqr scan --subscription-id <subscription_id> --resource-group <resource_group_name>
  ```

## Advanced Filtering

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

> Check the [overview](https://azure.github.io/azqr/docs/overview/) to get the resource type abbreviations.

## File Outputs

Currently Azure Quick Review supports 3 types of file outputs: `xlsx` (default), `csv`, `json`

### xlsx

`xlsx` is the default output format.

> Check the [overview](https://azure.github.io/azqr/docs/overview/) to get the more information.

### csv

By default `azqr` will create an xlsx document, However if you need to export to `csv` you can use the following flag: `--csv`

Example:

```bash
azqr scan --csv
```

The scan will generate 10 `csv` files:

```
<file-name>.advisor.csv
<file-name>.azurePolicy.csv
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

By default `azqr` will create an xlsx document, However if you need to export to `json` you can use the following flag: `--json`

Example:

```bash
azqr scan --json
```

The scan will generate a single consolidated `json` file:

``` 
<file-name>.json
```

The JSON file contains all data sections in a single consolidated structure:

```json
{
    "recommendations": [...],
    "impacted": [...],
    "resourceType": [...],
    "inventory": [...],
    "advisor": [...],
    "azurePolicy": [...],
    "defender": [...],
    "defenderRecommendations": [...],
    "costs": [...],
    "outOfScope": [...]
}
```

### Changing the Output File Name

You can change the output file name by using the `--output-file` or `-o` flag:

Powershell:

```powershell
$timestamp = Get-Date -Format 'yyyyMMddHHmmss'
azqr scan --output-file "azqr_action_plan_$timestamp"
```

Bash:

```bash
timestamp=$(date '+%Y%m%d%H%M%S')
azqr scan --output-file "azqr_action_plan_$timestamp"
```

> By default, the output file name is `azqr_action_plan_YYYY_MM_DD_THHMMSS`.

## Help

You can get help for `azqr` commands by running:

```console
azqr --help
```
