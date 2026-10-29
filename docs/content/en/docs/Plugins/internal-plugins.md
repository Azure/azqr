---
title: Internal Plugins
description: Built-in analysis plugins for advanced Azure resource insights
weight: 2
---

## Overview

Azure Quick Review (azqr) includes **internal plugins** - specialized built-in scanners that provide advanced analytics beyond standard best practice recommendations. Unlike YAML plugins (which add custom Resource Graph queries), internal plugins perform complex data analysis, API integrations, and multi-source data correlation.

Internal plugins are disabled by default and must be explicitly enabled using command-line flags.

## Available Internal Plugins

### 1. AI Governance

**Plugin Name**: `ai-gov`  
**Command**: `azqr ai-gov`  
**Flag**: `--plugin ai-gov`  
**Version**: 1.0.0

Monitors Azure OpenAI and Cognitive Services accounts for throttling (429 errors) to identify capacity constraints.

**Key Features**:
- Tracks 429 throttling errors by hour, model, and deployment
- Analyzes spillover configuration effectiveness
- Reports request counts by status code
- Identifies peak throttling periods

**Use Cases**:
- Capacity planning for OpenAI deployments
- Troubleshooting throttling issues
- Optimizing deployment spillover configuration
- Monitoring API usage patterns

**Output Columns**:
- Subscription, Resource Group, Account Name
- Kind (OpenAI, Cognitive Services)
- SKU and deployment details
- Model name and spillover settings
- Hourly throttling statistics (status code, request count)

**Data Source**: Azure Monitor Metrics API (last 24-48 hours)

---

### 2. Carbon Emissions

**Plugin Name**: `carbon-emissions`  
**Command**: `azqr carbon-emissions`  
**Flag**: `--plugin carbon-emissions`  
**Version**: 1.0.0

Analyzes carbon emissions by Azure resource type to support sustainability reporting and optimization.

**Key Features**:
- Tracks emissions by resource type across subscriptions
- Calculates month-over-month change ratios
- Aggregates emissions from multiple subscriptions
- Supports sustainability compliance reporting

**Use Cases**:
- Sustainability reporting and compliance
- Identifying high-emission resource types
- Tracking carbon reduction progress
- Environmental impact analysis

**Output Columns**:
- Period From/To (reporting period)
- Resource Type
- Latest Month Emissions
- Previous Month Emissions
- Month-over-Month Change Ratio
- Monthly Change Value
- Unit (metric tons CO2 equivalent)

**Data Source**: Azure Carbon Optimization API

---

### 3. Zone Mapping

**Plugin Name**: `zone-mapping`  
**Command**: `azqr zone-mapping`  
**Flag**: `--plugin zone-mapping`  
**Version**: 1.0.0

Retrieves logical-to-physical availability zone mappings for all Azure regions in each subscription.

**Key Features**:
- Maps logical zones (1, 2, 3) to physical zone identifiers
- Reveals subscription-specific zone mappings
- Helps ensure proper zone alignment across subscriptions
- Supports multi-subscription architecture planning

**Use Cases**:
- Multi-subscription architecture design
- DR planning with zone awareness
- Zone alignment for latency optimization
- Compliance and audit documentation

**Output Columns**:
- Subscription, Location, Display Name
- Logical Zone (1, 2, or 3)
- Physical Zone (e.g., `eastus-az1`, `westeurope-az2`)

**Data Source**: Azure Resource Manager Subscriptions API

[📖 Full Documentation](./zone-mapping)

---

### 4. Region Selection

**Plugin Name**: `region-selection`  
**Command**: `azqr region-selection`  
**Flag**: `--plugin region-selection`  
**Version**: 0.2.1

Scores and ranks Azure regions for workload migration or expansion. For each source region (where your resources currently live) × target region (candidate), the plugin computes a weighted 0–100 recommendation score across four dimensions.

| Dimension | Weight | Data source |
|-----------|-------:|-------------|
| Resource type availability | 35 % | ARM Providers API |
| SKU availability | 30 % | Per-resource-type ARM SKU APIs |
| Cost difference | 15 % | Azure Cost Management + Retail Prices API |
| Network latency | 20 % | Published Azure inter-region RTT statistics |

Availability zone loss/gain applies a multiplicative adjustment to the final score.

**Network Latency Details**:
- Values are **round-trip time (RTT)** sourced from Microsoft's published [Azure Network Round-trip Latency Statistics](https://learn.microsoft.com/en-us/azure/networking/azure-network-latency) (P50/median RTT in ms)
- Lookup order: exact measured region pair → reverse pair (RTT is treated as symmetric) → geographic cluster-pair average as a fallback estimate
- Estimated values (cluster-based, no direct measurement available) are shown as "X.X (est.)" in the Avg Latency (ms) column
- Pairs with no data or estimate available show "N/A" and are scored neutrally

**Key Features**:
- Qualitative **Recommended** (≥ 80), **Neutral** (60–79), **Not Recommended** (< 60) bands
- **Score Quality** flag notes when cost or latency data was unavailable
- Per-target-region **Svc Avail** Excel sheets with service and SKU availability per resource type
- **CostComparison** Excel sheet with per-meter retail pricing across all analysed regions
- Optional `--target-regions` flag to narrow analysis to specific candidates

**Use Cases**:
- Migration planning and DR site selection
- Regional expansion decisions
- Compliance with data-residency requirements

**Output Columns** (main sheet):
- Subscription, Source Region, Target Region
- Source Resource Type Count, Available/Unavailable Resource Types, Availability %
- Total SKUs Checked, Available/Unavailable/Restricted/Unknown SKUs, SKU Availability %
- Availability Zones, Avg Latency (ms), Avg Cost Difference %
- Recommendation Score, Score Quality, Recommendation
- Missing Resource Types, Unavailable SKUs (detail), Restricted SKUs (detail)

### 5. SQL Server EOL

**Plugin Name**: `sql-eol`  
**Command**: `azqr sql-eol`  
**Flag**: `--plugin sql-eol`  
**Version**: 0.6.0-beta

Analyzes SQL Server End-of-Life (EOL) and Extended Security Update (ESU) status across Arc-enabled SQL Server instances and SQL Virtual Machines on Azure.

**Key Features**:
- Detects EOL status dynamically using current date (Expired, ESU Active, Upcoming ESU, Supported)
- Models ESU billing the way Microsoft meters it: **once per host (OSE), per SQL Server version, at the highest edition present** — multiple same-version components (Engine, SSIS, SSAS, SSRS, PBIRS) on one host share a single ESU charge
- Enforces Standard edition **24-core cap** on ESU billing (per Microsoft ESU pricing docs; Enterprise and Web are uncapped)
- Passive **HA/DR replicas excluded** from ESU cost ($0 — covered under primary's SA failover rights)
- Reports **ESU subscription status** (`ESUEnabled`) from the Arc machine's `WindowsAgent.SqlServer` extension
- Estimates SQL Managed Instance migration savings with 2:1 consolidation model
- Covers both Arc-enabled SQL (on-prem, AWS, GCP) and Azure VM (SQL IaaS)

**ESU Billing Model**: ESU is metered once per OSE (Operating System Environment — a physical or virtual machine), per SQL Server version, at the highest edition installed for that version. All SQL Server service types (Engine, SSIS, SSAS, SSRS, Power BI Report Server) are ESU-eligible and roll into that single per-host+version charge. Free editions (Developer, Express, Evaluation) are never billed; passive HA/DR replicas are covered at $0 under the primary server's Software Assurance.

**Use Cases**:
- ESU cost forecasting and budgeting (accurate host-level model, not overcounted per-instance)
- Identifying which Arc hosts already have ESU subscribed vs. unsubscribed
- Migration planning to Azure SQL Managed Instance
- Compliance reporting for end-of-support software
- License optimization across SQL estates

**Output Columns**:
- Subscription, Resource Group, Name, Location
- Arc Server Name (underlying Arc machine name; `N/A (SQL VM)` for Azure VMs)
- Cloud Type (Arc-enabled (On-Prem) / Arc-enabled (AWS) / Arc-enabled (GCP) / Azure VM (SQL IaaS))
- Service Type (Engine / SSIS / SSAS / SSRS / PBIRS for Arc; Engine for SQL VMs)
- SQL Version, Edition
- EOL Status (Expired / ESU Active / Upcoming ESU / Supported)
- ESU Applicable (`Yes` / `No - free edition` / `No - passive HA/DR replica`)
- ESU Enabled (`Enabled` / `Not Enabled` for Arc; `Unknown (SQL VM)` for Azure VMs)
- ESU Start Date, ESU End Date
- Migration Target Tier (General Purpose / N/A)
- Migration Recommendation (actionable text)
- vCores, Billable Cores
- ESU Monthly Cost/Core
- SQL License Type, SQL License Cost/Core/Month, SQL License Monthly Cost
- VM Cost/Core/Month, Est VM Compute Monthly Cost
- Est ESU Monthly Cost (charged once per host+version; `$0` on secondary components)
- ESU Cost Basis (explains attribution: primary row, included, or N/A)
- Patch Ops Monthly Cost
- Current Monthly Cost (total current spend)
- Consolidation Ratio (2:1 default)
- Est SQL MI Monthly Cost, Est SQL MI Monthly Saving
- SQL MI Migration Verdict (Cost Savings / Break Even / Cost Increase)

### 6. VM v6/v7 Modernization Readiness

**Plugin Name**: `vm-modernization`  
**Command**: `azqr vm-modernization`  
**Flag**: `--plugin vm-modernization`  
**Version**: 0.3.0

Reports Azure VM v6/v7 series modernization readiness, following the Discover → Assess → Plan phases from the [VM v6/v7 modernization guidance](https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/lifecycle/sizes-v6-v7-modernization-discover): Discover-phase workload pattern, Assess-phase readiness signals, and a Plan-phase execution-method recommendation with a suggested target v6/v7 SKU.

**Key Features**:
- Classifies each VM's workload pattern (compute pool, AVD pooled host pool, service-managed, certified ISV appliance, or customer-managed) from Azure Resource Graph signals
- Reports Assess-phase readiness signals: Hyper-V generation, disk controller type (SCSI vs. NVMe), series generation, and Azure Disk Encryption status
- Adds heuristic platform-readiness verdicts: Generation 2 readiness, remote NVMe and MANA operating system support, a combined readiness outcome and a complexity rating. The operating system verdicts are derived from the image and OS version, so treat them as triage hints and confirm them against the image vendor's documentation
- Maps each source series to the replacement series documented on Microsoft Learn, with the specification differences, and reports the Gen1 backup policy subtype
- Adds a second sheet, `VM Modernization - SKU Avail`, that checks whether each suggested target SKU is offered in a region (including restrictions and zone limits) and whether the VM family quota covers the VMs that would move
- Recommends a Plan-phase execution method (redeploy / no action / manual review) and flags the Temp Disk Dependency and SAP/ISV Certification Gate hard gates where applicable
- Suggests a hardware-similar v6/v7 target SKU using the `internal/skus` compatibility advisor. The score reflects vCPU, memory and feature similarity only. It does not check CPU architecture, temp-disk needs, attached disk/NIC counts, vendor certification, or capacity. Regional availability and VM family quota are reported separately on the `VM Modernization - SKU Avail` sheet, which is not a capacity guarantee, so validate every suggestion
- Surfaces upcoming retirements and price changes from an embedded lifecycle/pricing dataset covering the D/Ds/Dv2/Dsv2/Ls, Av2/Amv2/B/F/Fs/Fsv2/G/Gs/Lsv2, NP/HC/HBv2 and Dv3/Dsv3/Ev3/Esv3 groups. Other series are not encoded, so `N/A` means no published lifecycle risk
- Computes per-VM current, projected, and monthly cost impact from the Azure Retail Prices API, using each VM's own current SKU (including pool-managed VMs)
- Covers standalone VMs, Availability Set members, VMSS Flex/Uniform members (including AKS node pools), and Databricks, AVD pooled and standard ARO nodes, with pool-level signals (`Orchestration Mode`, `Pool Model VM Size`, `Pool Model Series Generation`) on VMSS-backed rows
- Includes the raw inventory signals from the Assess article's starter query (placement, OS/image, storage, licensing, networking, security) next to the derived columns

**Scope**:
- The Validate phase from the [VM v6/v7 modernization guidance](https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/lifecycle/sizes-v6-v7-modernization-discover) (post-migration scoring/verification) is out of scope for this plugin; it covers only the Discover, Assess, and Plan phases

**Use Cases**:
- Discover-phase inventory: which VMs are customer-managed vs. pool-managed, and ready for per-VM modernization planning
- Assess-phase readiness triage: identifying Gen1-only VMs, SCSI-controller VMs needing the [SCSI-to-NVMe migration](https://learn.microsoft.com/en-us/azure/virtual-machines/migration/scsi-to-nvme-migration) path, and VMs already on v6/v7
- Plan-phase SKU selection: a starting v6/v7 target candidate per VM, ranked by hardware similarity (not migration eligibility)
- Lifecycle/pricing awareness: surfacing upcoming retirements and price changes relevant to each VM's current series
- Full-inventory export: every signal the Assess article's checklist references, not only the plugin's own verdicts

**Output Columns**:

Columns are grouped so an engineer reads left to right: identify, decide, target, blockers, cost, pool context, then raw inventory.

*Identify:*
- Subscription, Resource Group, Name, Location
- Current VM Size, Series Generation
- Workload Pattern (`A: Compute pool` / `B: AVD pooled host pool` / `C: Service-managed compute` / `E: Customer-managed VM (default)` / `G: Certified ISV appliance`; a generic VMSS member reads `A` with a Manual Review Required execution method until it is confirmed to be a stateless pool, and AVD-pooled and appliance signals take priority over scale-set membership), Pool Membership (`AKS Node Pool` / `ARO Cluster` / `VMSS` / `AVD Pooled Host Pool` / `Azure Databricks` / `Standalone` / appliance vendor name)

*Decide:*
- Platform Readiness Outcome (`Platform checks passed` / `OS image refresh needed` / `OS support unconfirmed` / `Gen2 migration needed` / `N/A (already on target generation)`): a combined, heuristic verdict on the blocker columns that says what is wrong, not what to do. Recommended Execution Method is the only column that says how to proceed, so pool-managed VMs still show their own platform verdict next to a "evaluate at the pool level" route
- Recommended Execution Method (one of `Redeploy`, `Manual Review Required`, `No action needed`, `Managed by pool/cluster`): a route suggestion, not a gate-cleared verdict. `Redeploy` applies to Gen2 customer-managed VMs and appliances, and is conditional on the Temp Disk Dependency and SAP/ISV Certification Gate reviews (appliances need vendor certification first). `Manual Review Required` covers Gen1 or unreported Hyper-V generation and generic VMSS members, which must be validated as stateless pools. `Managed by pool/cluster` means evaluate at the pool level; AVD pooled hosts also need golden-image checks. The reason is visible in the Hyper-V Generation, Generation 2 Ready, gate, Workload Pattern and Pool Membership columns
- Modernization Complexity (`LOW` / `MEDIUM` / `HIGH` / `REVIEW`, or `N/A` when already on a target generation) and Complexity Reason

*Target:*
- Suggested Target SKU, Target SKU Compatibility Score (0.0–1.0 hardware similarity, not migration readiness)
- Documented Replacement Series and Series Specification Differences: the replacement documented on Microsoft Learn for the VM's source series
- Lifecycle/Pricing Notes

*Blockers:*
- Generation 2 Ready, Remote NVMe OS Supported, MANA OS Supported (`TRUE` / `FALSE` / `UNKNOWN`, heuristic verdicts from the image and OS version)
- Hyper-V Generation, Disk Controller Type
- Temp Disk Dependency and SAP/ISV Certification Gate: `Manual Review Required` when a customer should check the gate, otherwise `N/A (pool/platform-managed)`. Temp Disk Dependency requires review for B, E, G; SAP/ISV for every pattern except A and C. A generic VMSS member requires review for both
- Azure Disk Encryption (`Enabled` / `Disabled`, aggregated across every attached disk)
- Hibernation Enabled, Backup Policy Subtype (the Azure Backup policy subtype protecting the VM, `Unknown` when the lookup fails), VM Power State, Instance View Captured

*Cost:*
- Current Monthly Cost (USD), Projected Monthly Cost After Increase (USD), Monthly Cost Impact (USD): existing-SKU retail exposure (hourly retail price x 730). Current cost uses the OS-matched rate (the Linux compute rate for Windows VMs with Azure Hybrid Benefit). The price increase percentage is applied only to the compute (Linux) rate, never to the Windows license portion. Only price changes not yet effective (UTC) are applied, because the current retail rate already includes earlier ones. Stopped (allocated) VMs are still billed; deallocated VMs read `N/A (VM deallocated)`
- Target Monthly Cost (USD), Target Projected Monthly Cost After Increase (USD), Target Monthly Cost Impact (USD): the same calculation for the Suggested Target SKU in the VM's region and OS. The projected cost equals the target cost when no price increase applies to the target SKU. Rows without a concrete target SKU read `N/A`; a missing retail price reads `Manual Review Required`

*Pool context:*
- Pool Membership, Membership Model (`Standalone` / `Availability Set` / `VMSS Flex`), Orchestration Mode
- VMSS Name, Pool Model VM Size, Pool Model Series Generation, Availability Set Name

*Full raw inventory columns (mirrors the Assess article's starter query):*
- Zone Placement (`Zonal` / `Regional`), Availability Zone, Platform Fault Domain
- OS Type, OS Version, Computer Name
- Is Marketplace VM, Image Publisher, Image Offer, Image SKU, Image Version
- OS Disk Type, OS Disk Size (GB), Ephemeral OS Disk, Data Disk Count
- Ultra SSD Enabled, License Type
- NIC Count, Accelerated Networking, IP Forwarding
- Security Type, Secure Boot Enabled, vTPM Enabled

**SKU Availability Sheet** (`VM Modernization - SKU Avail`):

One row per (subscription, region, suggested target SKU), with these columns:
- Subscription, Region, Suggested Target SKU, Source VM Sizes, VM Count
- Availability: `Available`, `Restricted` (with Restriction Reasons), `Zone restricted` (with Restricted Zones), `Not offered`, or `Unknown` when the Compute SKUs call failed. This is not a capacity guarantee
- Quota (right after Availability): `Sufficient`, `Near limit` (less than 15% of the limit would remain), `Insufficient`, `N/A` when the SKU is not offered in the region, or `Unknown` when the quota entry could not be found or a call failed. Supporting columns: Quota Family, vCPUs Needed, Family Quota Available. A redeploy needs the full vCPU count; an in-place resize needs only the change in the new family. The regional total vCPU limit is not checked

Only VMs with a concrete suggested target SKU are evaluated; pool-managed, already-on-target and manual-review VMs are skipped.

**Target Regions**:

The plugin accepts the same `target-regions` option as the [Region Selection](#4-region-selection) plugin, so one set of regions drives both:

```bash
# Standalone command
azqr vm-modernization --target-regions eastus,westeurope

# Combined scan
azqr scan --plugin region-selection --plugin vm-modernization --stage-param plugin.target-regions=eastus,westeurope
```

When set, every VM with a suggested target SKU is checked in each listed region, and the `Region` column shows the region checked. When unset, each VM is checked in its own region (Region Selection defaults to Sweden Central instead). Target regions affect only availability and quota; cost columns stay on each VM's current region.

---

### 7. Service Health Availability

**Plugin Name**: `service-health`  
**Command**: `azqr service-health`  
**Flag**: `--plugin service-health`  
**Version**: 0.1.0-beta

Analyzes Azure service health events to determine the percentage of time resources were unaffected by service issues over the last 90 days.

**Key Features**:
- Correlates resolved `ServiceIssue` events (and their impacted-resource associations) with the full resource inventory via Azure Resource Graph
- Computes an event-free percentage per resource, capped to the 90-day (2,160-hour) window to account for overlapping events, then averages within each (subscription, region, resource type) bucket
- Reports total distinct event counts and the number of affected resources per bucket
- Results are sorted worst-first (lowest event-free percentage first)
- Honors resource-type exclusion filters (`--filters`)

**Use Cases**:
- Reliability and SLA/availability reporting
- Identifying regions or resource types prone to service health issues
- Prioritizing resiliency investments based on historical impact

**Output Columns**:
- Subscription ID, Target Region, Target Resource Type
- Percentage Without Events, Events Count, Affected Resources

**Data Source**: Azure Resource Graph (`servicehealthresources` table, last 90 days)

---

## Usage

### Running Internal Plugins

Internal plugins can be executed in two ways:

#### 1. Standalone Plugin Commands (Recommended for Fast Execution)

Run plugins as top-level commands for optimized execution. This mode skips resource and APRL scanning, executing only the specified plugin:

```bash
# Run OpenAI throttling plugin
azqr ai-gov

# Run carbon emissions plugin
azqr carbon-emissions

# Run zone mapping plugin
azqr zone-mapping

# Run region selection plugin
azqr region-selection

# Narrow region selection to specific target regions
azqr region-selection --target-regions=swedencentral,germanywestcentral

# Run SQL EOL plugin
azqr sql-eol

# Run VM v6/v7 modernization readiness plugin
azqr vm-modernization

# Run service health availability plugin
azqr service-health

# Run with specific subscriptions
azqr zone-mapping --subscription-id <sub-id>

# Run with custom output name
azqr ai-gov --output-name throttling-report
```

**Benefits of Standalone Mode:**
- ⚡ **Faster execution** - Skips resource scanning
- 📊 **Cleaner reports** - Contains only plugin results
- 🎯 **Focused analysis** - Dedicated to specific plugin output

#### 2. Integrated with Full Scan

Run plugins alongside standard compliance scanning using the `--plugin` flag:

```bash
# Enable single plugin during scan
azqr scan --plugin ai-gov

# Enable multiple plugins during scan
azqr scan --plugin ai-gov --plugin carbon-emissions --plugin zone-mapping --plugin region-selection --plugin sql-eol --plugin vm-modernization --plugin service-health

# Combine with other scan options
azqr scan --subscription-id <sub-id> --plugin zone-mapping --output-name analysis
```

**When to Use Scan Integration:**
- Need both compliance recommendations and plugin analysis
- Want consolidated report with all data
- Running comprehensive assessments

### Listing Available Plugins

View all registered plugins (internal and YAML):

```bash
azqr plugins list
```

**Sample Output**:
```
NAME                  VERSION    TYPE       DESCRIPTION
ai-gov     1.0.0      internal   Checks OpenAI/Cognitive Services accounts for...
carbon-emissions      1.0.0      internal   Analyzes carbon emissions by Azure resource type
zone-mapping          1.0.0      internal   Retrieves logical-to-physical availability zone mappings...
region-selection      0.2.1 internal   Analyzes optimal Azure region selection based on service availability...
sql-eol               0.6.0-beta internal   Analyzes SQL Server End-of-Life and Extended Security Update status
vm-modernization      0.3.0      internal   Reports Azure VM v6/v7 series modernization readiness
service-health        0.1.0-beta internal   Analyzes Azure service health events to determine...
```

### Plugin Details

Get detailed information about a specific plugin:

```bash
azqr plugins info zone-mapping
```

## Output Formats

Internal plugin results are included in all output formats:

### Excel (Default)

Each internal plugin creates a dedicated worksheet in the Excel workbook:
- **Zone Mapping** sheet
- **AI Gov** sheet  
- **Carbon Emissions** sheet
- **Region Selection** sheet (main scored table)
  - **Svc Avail `<region>`** sheets — one per target region with per-resource-type availability
  - **CostComparison** sheet — per-meter retail pricing across all analysed regions
- **SQL EOL** sheet
- **VM Modernization** sheet
  - **VM Modernization - Notices** sheet — account-wide lifecycle/pricing notices that apply regardless of VM family or region, omitted on Azure Government/Azure China
- **Service Issues** sheet

```bash
# Run plugins as standalone commands (fastest)
azqr ai-gov
azqr carbon-emissions
azqr zone-mapping
azqr region-selection

# Or run with full scan
azqr scan --plugin ai-gov --plugin carbon-emissions --plugin zone-mapping --plugin region-selection
# Generates: azqr_action_plan_YYYY_MM_DD_THHMMSS.xlsx
```

### JSON

Plugin results are included in the `pluginResults` array:

```bash
# Run as standalone command
azqr zone-mapping --json

# Or run with full scan
azqr scan --plugin zone-mapping --json
```

**JSON Structure**:
```json
{
  "recommendations": [...],
  "resources": [...],
  "pluginResults": [
    {
      "pluginName": "zone-mapping",
      "sheetName": "Zone Mapping",
      "description": "Retrieves logical-to-physical availability zone mappings for all Azure regions in each subscription",
      "table": [
        ["Subscription", "Location", "Display Name", "Logical Zone", "Physical Zone"],
        ["Production", "East US", "East US", "1", "eastus-az1"]
      ]
    }
  ]
}
```

### CSV

Plugin results are exported to separate CSV files:

```bash
# Run as standalone command
azqr zone-mapping --csv

# Or run with full scan
azqr scan --plugin zone-mapping --csv
# Generates: 
#   <filename>.zone-mapping.csv
#   <filename>.recommendations.csv
#   <filename>.inventory.csv
#   ...
```

## Permissions

Internal plugins may require additional permissions beyond standard `Reader` access:

| Plugin | Required Permissions | API Dependencies |
|--------|---------------------|------------------|
| **zone-mapping** | Reader | Subscriptions API (locations endpoint) |
| **ai-gov** | Reader + Monitoring Reader | Cognitive Services, Monitor Metrics |
| **carbon-emissions** | Reader | Carbon Optimization API |
| **sql-eol** | Reader | Azure Resource Graph |
| **vm-modernization** | Reader (includes `Microsoft.Compute/skus/read` and `Microsoft.Compute/locations/usages/read`) | Azure Resource Graph (including Recovery Services backup data), Azure Retail Prices API, Compute Resource SKUs API, Compute usages API |
| **service-health** | Reader | Azure Resource Graph (service health events) |

**Recommended**: Assign `Reader` and `Monitoring Reader` roles at subscription or management group scope.

## Performance Considerations

Internal plugins add processing time to scans:

- **ai-gov**: 1-3 minutes (depends on number of OpenAI accounts)
- **carbon-emissions**: 1-2 minutes (depends on subscription count)
- **zone-mapping**: <10 seconds (very fast, one API call per subscription)
- **sql-eol**: <30 seconds (single Azure Resource Graph query)
- **vm-modernization**: one main Azure Resource Graph query plus separate AVD, ARO ownership and scale set lookups, followed by batched/concurrent Azure Retail Prices API lookups across the estate's unique VM SKU/region pairs, plus one Compute SKUs and one usages call per unique subscription/region pair for the availability sheet. Latency depends on inventory size, ARG batching/pagination, and the number of distinct SKU/region pairs; each price batch request covers only 8 pairs.
- **service-health**: <30 seconds (single Azure Resource Graph query against 90 days of service health events)

**Optimization Tips**:
- Enable only needed plugins
- Use subscription/resource group filters to reduce scope
