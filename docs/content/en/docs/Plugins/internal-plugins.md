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
**Version**: 0.1.0-beta

Scores and ranks Azure regions for workload migration or expansion. For each source region (where your resources currently live) × target region (candidate), the plugin computes a weighted 0–100 recommendation score across four dimensions.

| Dimension | Weight | Data source |
|-----------|-------:|-------------|
| Resource type availability | 35 % | ARM Providers API |
| SKU availability | 30 % | Per-resource-type ARM SKU APIs |
| Cost difference | 15 % | Azure Cost Management + Retail Prices API |
| Network latency | 20 % | Published Azure inter-region RTT statistics |

Availability zone loss/gain applies a multiplicative adjustment to the final score.

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

### 5. SQL Server ESU Status

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
azqr scan --plugin ai-gov --plugin carbon-emissions --plugin zone-mapping --plugin region-selection --plugin sql-eol

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
region-selection      0.1.0-beta internal   Scores and ranks Azure regions for workload migration...
sql-eol               0.6.0-beta internal   Analyzes SQL Server End-of-Life and Extended Security Update status
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

**Recommended**: Assign `Reader` and `Monitoring Reader` roles at subscription or management group scope.

## Performance Considerations

Internal plugins add processing time to scans:

- **ai-gov**: 1-3 minutes (depends on number of OpenAI accounts)
- **carbon-emissions**: 1-2 minutes (depends on subscription count)
- **zone-mapping**: <10 seconds (very fast, one API call per subscription)
- **sql-eol**: <30 seconds (single Azure Resource Graph query)

**Optimization Tips**:
- Enable only needed plugins
- Use subscription/resource group filters to reduce scope
