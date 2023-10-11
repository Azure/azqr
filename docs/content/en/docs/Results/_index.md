---
title: Scan Results
description: Scan Results
weight: 4
---

Azure Quick Review (azqr) creates an excel spreadsheet with the following sections:

* [Overview](#overview)
* [Recommendations](#recommendations)
* [Services](#services)
* [Defender](#defender)
* [Advisor](#advisor)
* [Costs](#costs) (Disabled by default)

## Overview

The overview section contains the following information:

* **SubscriptionID**: This is the unique identifier for the Azure subscription under which the resource is deployed.
* **ResourceGroup**: The resource group where the resource is deployed.
* **Location**: The geographical region where the resource is deployed.
* **Type**: The specific type or category of the Azure resource.
* **Name**: The name assigned to the resource, providing a human-readable identifier for easy reference and management.
* **SKU**: The SKU represents the specific variant or configuration of the Azure resource. It defines the characteristics and capabilities of the resource.
* **SLA**: The Service Level Agreement (SLA) represents the agreed-upon performance and availability guarantees for the Azure service based on its current configuration.
* **AZ**: A Boolean value indicating whether the service is "Availability Zone aware." Availability Zones are physically separate datacenters within an Azure region, providing increased resiliency and fault tolerance for critical services.
* **PVT**: A Boolean value indicating whether the service has a private IP address. Private IP addresses are used for internal communication within Azure Virtual Networks.
* **DS**: A Boolean value indicating whether diagnostic settings are enabled for the service. Diagnostic settings allow you to collect logs, metrics, and other monitoring data for Azure resources.
* **CAF**: A Boolean value indicating whether the service is compliant with the [Cloud Adoption Framework](https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations) (CAF) naming convention. The CAF provides best practices and guidance for organizations adopting Azure.

![overview](/azqr/img/overview.png)

## Recommendations

The recommendations section contains a summary of the recommendations for the scanned services:

* **Id**: The unique identifier for the rule.
* **Category**: The category of the rule.
* **Subcategory**: The subcategory of the rule.
* **Description**: The description of the rule.
* **Severity**: The severity of the rule (High, Medium, Low).
* **Learn**: Link to relevant documentation.

![recommendations](/azqr/img/recommendations.png)

## Services

The services section contains the following information:

* **Subscription**: This is the unique identifier for the Azure subscription under which the resource is deployed.
* **Resource Group**: The resource group where the resource is deployed. 
* **Location**: The geographical region where the resource is deployed.
* **Type**: The specific type or category of the Azure resource.
* **Service Name**: The name assigned to the resource.
* **Category**: The category of the rule.
* **Subcategory**: The subcategory of the rule.
* **Severity**: The severity of the rule (High, Medium, Low).
* **Description**: The description of the rule.
* **Result**: The result of the rule evaluation.
* **Broken**: True if the rule is broken.
* **Learn**: Link to relevant documentation.

![services](/azqr/img/services.png)

## Defender

The defender section contains the following information:

* **Name**: Microsoft Defender for Cloud plan name.
* **Tier**: The tier of the plan.
* **Deprecated**: True if the plan is deprecated.

![defender](/azqr/img/defender.png)

## Advisor

This section shows the Azure Advisor Recommendations with the following information:

* **Subscription Id**: This is the unique identifier for the Azure subscription under which the resource is deployed.
* **Name**: The name of the resource identified by Advisor.
* **Type**: The resource type of the resource identified by Advisor.
* **Category**: The category of the recommendation.
* **Description**: The description of the recommendation.
* **PotentialBenefits**: The potential benefits of the recommendation.
* **Risk**: Risk related to the recommendation.
* **LearnMoreLink** Link to relevant documentation.

## Costs

Displays the Azure Actual Costs for the period from the first day of the current month until the day Azure Quick Review (azqr) is used.
