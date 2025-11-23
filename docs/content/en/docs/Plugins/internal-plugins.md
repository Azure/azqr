---
title: Internal Plugins
description: Built-in analysis plugins for advanced Azure resource insights
weight: 2
---

## Overview

Azure Quick Review (azqr) includes **internal plugins** - specialized built-in scanners that provide advanced analytics beyond standard best practice recommendations. Unlike YAML plugins (which add custom Resource Graph queries), internal plugins perform complex data analysis, API integrations, and multi-source data correlation.

Internal plugins are disabled by default and must be explicitly enabled using command-line flags.

## Available Internal Plugins

### 1. Region Selection

**Plugin Name**: `region-selection`  
**Flag**: `--region-selection`  
**Version**: 1.0.0

Analyzes optimal Azure region selection based on service availability, network latency, and cost comparison.

**Key Features**:
- Multi-factor region scoring (availability, latency, cost)
- Service availability validation across regions
- Network latency analysis using Azure RTT statistics
- Cost comparison based on Azure Retail Prices
- Identifies best alternative regions for migration or DR

**Scoring Weights**:
- **Service Availability**: 50% - Ensures all resource types are available
- **Network Latency**: 30% - Evaluates performance using Azure RTT data
- **Cost Comparison**: 20% - Compares regional pricing differences

**Use Cases**:
- Disaster recovery planning and region pairing
- Cost optimization through region migration
- Multi-region strategy evaluation
- Migration feasibility assessment

**Output Columns**:
- Source Region, Target Region
- Recommendation Score (0-100)
- Availability Percentage
- Average Latency (ms)
- Average Cost Difference (%)
- Missing Resource Types

**Data Source**: Azure Resource Graph, Resource Providers API, Retail Prices API, Cost Management API

---

### 2. OpenAI Throttling

**Plugin Name**: `openai-throttling`  
**Flag**: `--openai-throttling`  
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

### 3. Carbon Emissions

**Plugin Name**: `carbon-emissions`  
**Flag**: `--carbon-emissions`  
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

### 4. Zone Mapping

**Plugin Name**: `zone-mapping`  
**Flag**: `--zone-mapping`  
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

---

## Usage

### Enabling Internal Plugins

Internal plugins are opt-in and must be enabled individually using command-line flags:

```bash
# Enable region selection plugin
azqr scan --region-selection

# Enable OpenAI throttling plugin
azqr scan --openai-throttling

# Enable carbon emissions plugin
azqr scan --carbon-emissions

# Enable zone mapping plugin
azqr scan --zone-mapping

# Enable multiple plugins
azqr scan --region-selection --openai-throttling --carbon-emissions --zone-mapping

# Combine with other scan options
azqr scan --subscription-id <sub-id> --region-selection --output-name analysis
```

### Listing Available Plugins

View all registered plugins (internal and YAML):

```bash
azqr plugins list
```

**Sample Output**:
```
NAME                  VERSION    TYPE       DESCRIPTION
region-selection      1.0.0      internal   Analyzes optimal Azure region selection based on...
openai-throttling     1.0.0      internal   Checks OpenAI/Cognitive Services accounts for...
carbon-emissions      1.0.0      internal   Analyzes carbon emissions by Azure resource type
zone-mapping          1.0.0      internal   Retrieves logical-to-physical availability zone mappings...
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
- **Region Selection** sheet
- **Zone Mapping** sheet
- **OpenAI Throttling** sheet  
- **Carbon Emissions** sheet

```bash
azqr scan --region-selection --openai-throttling --carbon-emissions --zone-mapping
# Generates: azqr_action_plan_YYYY_MM_DD_THHMMSS.xlsx
```

### JSON

Plugin results are included in the `pluginResults` array:

```bash
azqr scan --zone-mapping --json
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
azqr scan --zone-mapping --csv
# Generates: 
#   <filename>.zone-mapping.csv
#   <filename>.recommendations.csv
#   <filename>.inventory.csv
#   ...
```

## Interactive Dashboard

View plugin results interactively using the `show` command:

```bash
# Generate report with plugins
azqr scan --openai-throttling --carbon-emissions --zone-mapping --output-name analysis

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
| **region-selection** | Reader | Resource Graph, Resource Providers, Retail Prices, Cost Management |
| **zone-mapping** | Reader | Subscriptions API (locations endpoint) |
| **openai-throttling** | Reader + Monitoring Reader | Cognitive Services, Monitor Metrics |
| **carbon-emissions** | Reader | Carbon Optimization API |

**Recommended**: Assign `Reader` and `Monitoring Reader` roles at subscription or management group scope.

## Performance Considerations

Internal plugins add processing time to scans:

- **region-selection**: 2-5 minutes (depends on number of regions and resource types)
- **openai-throttling**: 1-3 minutes (depends on number of OpenAI accounts)
- **carbon-emissions**: 1-2 minutes (depends on subscription count)
- **zone-mapping**: <10 seconds (very fast, one API call per subscription)

**Optimization Tips**:
- Enable only needed plugins
- Use subscription/resource group filters to reduce scope
