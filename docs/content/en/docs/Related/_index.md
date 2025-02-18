---
title: Related Projects
description: Azure Quick Review compared to APRL, Azure Review Checklists and PSRule.Rules.Azure
weight: 6
---

## AZQR and APRL

As of version 2.0.0-preview, **Azure Quick Review (azqr)** includes all [Azure Resource Graph](https://learn.microsoft.com/azure/governance/resource-graph/overview) queries provided by the the [Azure Proactive Resiliency Library (APRL)](https://aka.ms/aprl), which are used to identify non-compliant resources.

**Azure Quick Review (azqr)** extends [APRL](https://aka.ms/aprl) by providing per service instance SLAs, Diagnostic Settings detection and more. Therefore, scan results display `AZQR` or `APRL`, to indicate the source of the recommendation.

> **APRL** provides a curated catalog of resiliency recommendations for workloads running in Azure. Many of the recommendations contain supporting Azure Resource Graph (ARG) queries

## AZQR and Azure Orphan Resources

As of version 2.4.0 **Azure Quick Review (azqr)** includes all [Azure Resource Graph](https://learn.microsoft.com/azure/governance/resource-graph/overview) queries provided by the the [Azure Orphan Resources](https://github.com/dolevshor/azure-orphan-resources) project

## AZQR compared to Azure Review Checklists and PSRule.Rules.Azure

**Azure Quick Review (azqr)** was created to address a very specific need we had back in 2022. Initially, we had to run three assessments to get a clear picture of various solutions in terms of SLAs, use of Availability Zones, and Diagnostic Settings. At the time, we were not aware of the existence of the [`review-checklist`](https://github.com/Azure/review-checklists) or [`PSRule.Rules.Azure`](https://github.com/Azure/PSRule.Rules.Azure).

When some of our peers saw the assessments we were able to deliver with the early bits of **Azure Quick Review (azqr)**, they asked us to add more checks (recommendations) and change the output format from markdown to Excel.

As many of our customers work in restrictive environments, the ability to run a self-contained, cross-platform binary while using read-only permissions became a key feature.

Moving forward to 2023, based on great feedback from both peers and customers, we moved the original repo to the [Azure](https://aka.ms/azqr) organization, added support for more services, fixed some issues and even added a Power BI template.

In August 2024, we added all [APRL](https://aka.ms/aprl) recommendations to **Azure Quick Review (azqr)** and removed duplicates in favor of the ones already available as [Azure Resource Graph](https://learn.microsoft.com/azure/governance/resource-graph/overview) queries.

When compared with [`PSRule.Rules.Azure`](https://github.com/Azure/PSRule.Rules.Azure), **Azure Quick Review (azqr)** only scans deployed Azure resources and provides recommendations based on the current state. **Azure Quick Review (azqr)** does not scan ARM templates or Bicep files.

When compared to the [`review-checklist`](https://github.com/Azure/review-checklists), **Azure Quick Review (azqr)** also provides an actionable list of more than 400 recommendations (70+ Azure resource types), that can be used to improve the resiliency of your Azure solutions.
