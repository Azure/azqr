---
title: Usage
description: Use Azure Quick Review &mdash; to analyze Azure resources and identify whether they comply with Azure's best practices and recommendations.
weight: 1
---

## Authentication

**Azure Quick Review (azqr)** supports the following authentication methods:

* Azure CLI
* Service Principal. You'll need to set the following environment variables:
  * AZURE_CLIENT_ID
  * AZURE_CLIENT_SECRET
  * AZURE_TENANT_ID

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
