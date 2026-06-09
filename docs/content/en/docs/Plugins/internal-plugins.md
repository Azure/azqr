---
title: Internal Plugins
description: Built-in analysis plugins for advanced Azure resource insights
weight: 2
---

## Overview

Azure Quick Review (azqr) includes **internal plugins** - specialized built-in scanners that provide advanced analytics beyond standard best practice recommendations. Unlike YAML plugins (which add custom Resource Graph queries), internal plugins perform complex data analysis, API integrations, and multi-source data correlation.

Internal plugins are disabled by default and must be explicitly enabled using command-line flags.

## Available Internal Plugins

### 1. OpenAI Throttling

**Plugin Name**: `openai-throttling`  
**Command**: `azqr openai-throttling`  
**Flag**: `--plugin openai-throttling`  
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

**Plugin Name**: `sql-esu`  
**Command**: `azqr sql-esu`  
**Flag**: `--plugin sql-esu`  
**Version**: 0.1.0-beta

Analyzes SQL Server End-of-Life (EOL) and Extended Security Update (ESU) status across Arc-enabled SQL Server instances and SQL Virtual Machines on Azure.

**Key Features**:
- Detects EOL status dynamically using current date (Expired, ESU Active, Upcoming ESU, Supported)
- Calculates ESU licensing costs per instance based on edition and vCore count
- Estimates SQL Managed Instance migration savings
- Covers both Arc-enabled SQL (on-prem) and Azure VM (SQL IaaS)

**Use Cases**:
- ESU cost forecasting and budgeting
- Migration planning to Azure SQL Managed Instance
- Compliance reporting for end-of-support software
- License optimization across SQL estates

**Output Columns**:
- Name, Resource Group, Subscription, Location
- Cloud Type (Arc-enabled or Azure VM)
- SQL Version, Edition, vCores
- EOL Status (Expired / ESU Active / Upcoming ESU / Supported)
- Mainstream End Date, ESU Start Date, ESU End Date
- ESU Monthly Cost/Core, Billable Cores
- Estimated Monthly/Annual/3-Year Cost
- Patch Ops Monthly/Annual/3-Year Cost
- Est SQL MI Monthly/Annual/3-Year Cost
- Est SQL MI Monthly/Annual/3-Year Saving

---

## Usage

### Running Internal Plugins

Internal plugins can be executed in two ways:

#### 1. Standalone Plugin Commands (Recommended for Fast Execution)

Run plugins as top-level commands for optimized execution. This mode skips resource and APRL scanning, executing only the specified plugin:

```bash
# Run OpenAI throttling plugin
azqr openai-throttling

# Run carbon emissions plugin
azqr carbon-emissions

# Run zone mapping plugin
azqr zone-mapping

# Run region selection plugin
azqr region-selection

# Narrow region selection to specific target regions
azqr region-selection --target-regions=swedencentral,germanywestcentral

# Run SQL ESU plugin
azqr sql-esu

# Run with specific subscriptions
azqr zone-mapping --subscription-id <sub-id>

# Run with custom output name
azqr openai-throttling --output-name throttling-report
```

**Benefits of Standalone Mode:**
- ⚡ **Faster execution** - Skips resource scanning
- 📊 **Cleaner reports** - Contains only plugin results
- 🎯 **Focused analysis** - Dedicated to specific plugin output

#### 2. Integrated with Full Scan

Run plugins alongside standard compliance scanning using the `--plugin` flag:

```bash
# Enable single plugin during scan
azqr scan --plugin openai-throttling

# Enable multiple plugins during scan
azqr scan --plugin openai-throttling --plugin carbon-emissions --plugin zone-mapping --plugin region-selection --plugin sql-esu

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
openai-throttling     1.0.0      internal   Checks OpenAI/Cognitive Services accounts for...
carbon-emissions      1.0.0      internal   Analyzes carbon emissions by Azure resource type
zone-mapping          1.0.0      internal   Retrieves logical-to-physical availability zone mappings...
region-selection      0.1.0-beta internal   Scores and ranks Azure regions for workload migration...
sql-esu               0.1.0-beta internal   Analyzes SQL Server End-of-Life and Extended Security Update status
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
- **OpenAI Throttling** sheet  
- **Carbon Emissions** sheet
- **Region Selection** sheet (main scored table)
  - **Svc Avail `<region>`** sheets — one per target region with per-resource-type availability
  - **CostComparison** sheet — per-meter retail pricing across all analysed regions
- **SQL ESU** sheet

```bash
# Run plugins as standalone commands (fastest)
azqr openai-throttling
azqr carbon-emissions
azqr zone-mapping
azqr region-selection

# Or run with full scan
azqr scan --plugin openai-throttling --plugin carbon-emissions --plugin zone-mapping --plugin region-selection
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

## Interactive Dashboard

View plugin results interactively using the `show` command:

```bash
# Generate report with plugins (standalone commands)
azqr openai-throttling --output-name analysis
azqr carbon-emissions --output-name analysis
azqr zone-mapping --output-name analysis

# Or generate with full scan
azqr scan --plugin openai-throttling --plugin carbon-emissions --plugin zone-mapping --output-name analysis

# Launch interactive viewer
azqr show -f analysis.xlsx --open
```

The dashboard provides:
- Filterable columns (dropdowns, search)
- Sortable data tables
- Export capabilities
- Real-time filtering

## Permissions

Internal plugins may require additional permissions beyond standard `Reader` access:

| Plugin | Required Permissions | API Dependencies |
|--------|---------------------|------------------|
| **zone-mapping** | Reader | Subscriptions API (locations endpoint) |
| **openai-throttling** | Reader + Monitoring Reader | Cognitive Services, Monitor Metrics |
| **carbon-emissions** | Reader | Carbon Optimization API |
| **sql-esu** | Reader | Azure Resource Graph |

**Recommended**: Assign `Reader` and `Monitoring Reader` roles at subscription or management group scope.

## Performance Considerations

Internal plugins add processing time to scans:

- **openai-throttling**: 1-3 minutes (depends on number of OpenAI accounts)
- **carbon-emissions**: 1-2 minutes (depends on subscription count)
- **zone-mapping**: <10 seconds (very fast, one API call per subscription)
- **sql-esu**: <30 seconds (single Azure Resource Graph query)

**Optimization Tips**:
- Enable only needed plugins
- Use subscription/resource group filters to reduce scope
