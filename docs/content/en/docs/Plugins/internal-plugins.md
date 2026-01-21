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
azqr scan --plugin openai-throttling --plugin carbon-emissions --plugin zone-mapping

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

```bash
# Run plugins as standalone commands (fastest)
azqr openai-throttling
azqr carbon-emissions
azqr zone-mapping

# Or run with full scan
azqr scan --plugin openai-throttling --plugin carbon-emissions --plugin zone-mapping
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

**Recommended**: Assign `Reader` and `Monitoring Reader` roles at subscription or management group scope.

## Performance Considerations

Internal plugins add processing time to scans:

- **openai-throttling**: 1-3 minutes (depends on number of OpenAI accounts)
- **carbon-emissions**: 1-2 minutes (depends on subscription count)
- **zone-mapping**: <10 seconds (very fast, one API call per subscription)

**Optimization Tips**:
- Enable only needed plugins
- Use subscription/resource group filters to reduce scope
