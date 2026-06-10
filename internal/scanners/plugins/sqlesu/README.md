# SQL ESU Azure Resource Graph Query

> **Version: 0.1.0-beta** — This query is in early development. Output schema and cost estimates may change without notice.

## Overview

The `sql-esu.kql` file is an Azure Resource Graph (ARG) query that analyzes SQL Server **End-of-Life (EOL)** and **Extended Security Update (ESU)** status across your Azure environment. It produces lifecycle status and cost projections for every discovered SQL Server instance.

## What the Query Does

The KQL query (`kql/sql-esu.kql`):

1. **Discovers** SQL Server instances from two resource types:
   - `microsoft.azurearcdata/sqlserverinstances` (Arc-enabled / on-premises)
   - `microsoft.sqlvirtualmachine/sqlvirtualmachines` (SQL VMs on Azure)
2. **Resolves** the SQL Server version, edition, and vCore count (joining the underlying VM for IaaS instances).
3. **Classifies** each instance into an EOL status: `Supported`, `Upcoming ESU`, `ESU Active`, or `Expired`.
4. **Estimates costs** including ESU monthly/annual/3-year costs, maintenance operations overhead, and potential Azure SQL Managed Instance migration savings.

## Assumptions

The following assumptions are built into the KQL query:

| Area | Assumption |
|------|------------|
| **ESU pricing** | Standard/Web editions: **$139/core/month**. Enterprise: **$540.50/core/month**. Developer/Express: **$0** (free). Prices reflect public list pricing and do not account for EA/CSP discounts. |
| **Minimum billable cores** | Microsoft bills a minimum of **4 cores** per instance, even if fewer physical/virtual cores are present. |
| **Maintenance operations cost** | Estimated at **$160/month** per instance (2 hours/month × $80/hour) for manual maintenance overhead. |
| **SQL MI migration cost** | General Purpose tier: **$123/vCore/month**. Business Critical tier: **$367/vCore/month**. Enterprise editions map to BC; all others to GP. |
| **VM vCore resolution** | For SQL VMs, vCore counts are derived from the VM size SKU using a lookup table for legacy series and regex extraction for modern v3/v4/v5/v6 SKUs. Unrecognized sizes default to **4 vCores**. |
| **EOL/ESU dates** | Based on Microsoft's published lifecycle dates as of June 2025. The query uses `now()` so status transitions happen automatically. |
| **Upcoming ESU window** | Versions are flagged as "Upcoming ESU" **6 months** before their ESU start date to allow planning time. |
| **Savings calculation** | When ESU status is `Expired` or `Supported` (no ESU cost), the estimated saving from migrating to SQL MI reflects only the patch operations cost avoided ($160/month). For editions with no ESU cost (Developer/Express), the saving similarly reflects only operational overhead, not licensing. |
| **Scope** | Only resources visible in Azure Resource Graph are scanned. On-premises SQL Servers without Arc enrollment are not discoverable. |

## Output Columns

The query produces the following columns:

- **Name** / **Resource Group** / **Subscription** / **Location** — Resource identity
- **Cloud Type** — `Arc-enabled (On-Prem)` or `Azure VM (SQL IaaS)`
- **SQL Version** / **Edition** / **vCores** — Instance details
- **EOL Status** — Lifecycle classification
- **Mainstream End Date** / **ESU Start Date** / **ESU End Date** — Key dates
- **ESU Monthly Cost/Core** / **Billable Cores** — Pricing inputs
- **Estimated Monthly/Annual/3-Year Cost** — Projected ESU spend
- **Patch Ops Monthly/Annual/3-Year Cost** — Operational overhead
- **Est SQL MI Monthly/Annual/3-Year Cost** — Migration target cost
- **Est SQL MI Monthly/Annual/3-Year Saving** — Potential savings from migration

## Limitations (Beta)

- Cost estimates use public list pricing and do not reflect negotiated discounts.
- SQL MI migration estimates are simplified (single GP/BC tier mapping) and do not account for reserved capacity, hybrid benefit, or workload-specific sizing.
- VM size lookup covers common SKU families but may not include every available Azure VM size.
- The query does not detect SQL Server instances running inside containers or on non-Azure VMs without Arc enrollment.

## Plugin Integration

This KQL query is used by the `sql-esu` internal plugin in [Azure Quick Review (`azqr`)](https://github.com/Azure/azqr). The plugin executes this query via Azure Resource Graph, processes the results, and outputs them as a dedicated "SQL ESU" sheet in the `azqr` Excel report.

```bash
# Run as standalone command (fast, plugin-only mode)
azqr sql-esu

# Or integrate with a full scan
azqr scan --plugin sql-esu
```

> Plugin commands (e.g., `azqr sql-esu`) run in optimized plugin-only mode for faster execution, skipping resource and APRL scanning. Use `azqr plugins list` to see all available plugins.

## License

Licensed under the [MIT License](../../../../LICENSE).
