---
title: Usage
description: Use Azure Quick Review &mdash; to analyze Azure resources and identify whether they comply with Azure's best practices and recommendations.
weight: 2
---

## Authentication

**Azure Quick Review (azqr)** supports the following authentication methods:

* Service Principal. You'll need to set the following environment variables:
  * AZURE_CLIENT_ID
  * AZURE_CLIENT_SECRET
  * AZURE_TENANT_ID
* Azure Managed Identity
* Azure CLI (Using this type of authentication will make scans run slower)

## Authorization

**Azure Quick Review (azqr)** requires the following permissions:

* Subscription Reader

## Running the Scan

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

## Filtering Recommendations and more

You can configure Azure Quick Review to include or exclude specific subscriptions or resource groups and also exclude services or recommendations. To do so, create a `yaml` file with the following format:

```yaml
azqr:
  include:
    subscriptions:
      - <subscription_id> # format: <subscription_id>
    resourceGroups:
      - <resource_group_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>
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