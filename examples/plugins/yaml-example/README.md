# YAML Plugin Example

This example demonstrates how to create a YAML-based plugin for azqr that executes custom Azure Resource Graph queries.

## Overview

YAML plugins allow you to define custom Azure Resource Graph (ARG) queries without writing Go code. They are perfect for:
- Custom governance checks
- Organization-specific policies
- Quick prototyping of new checks
- Sharing queries across teams

## File Structure

```
examples/plugins/yaml-example/
├── custom-checks.yaml    # Main plugin configuration
├── kql/                  # Optional: External KQL query files
│   └── unused-public-ips.kql
└── README.md            # This file
```

## Plugin Configuration

### Basic Structure

```yaml
name: my-custom-plugin          # Unique plugin identifier
version: 1.0.0                  # Semantic version
description: My custom checks   # Brief description
author: Your Name              # Optional
license: MIT                   # Optional

queries:                       # List of queries to execute
  - description: Check description
    aprlGuid: unique-id-001
    # ... query configuration
```

### Query Configuration

Each query requires:

```yaml
- description: What this query checks for
  aprlGuid: unique-identifier  # Must be unique across all queries
  recommendationControl: Category  # See categories below
  recommendationImpact: High|Medium|Low
  recommendationResourceType: Microsoft.Service/resourceType
  recommendationMetadataState: Active|Preview|Deprecated
  longDescription: |
    Detailed explanation of the recommendation
    and why it matters.
  potentialBenefits: What you gain by following this
  pgVerified: true|false  # Verified by product group
  automationAvailable: true|false
  tags:
    - tag1
    - tag2
  learnMoreLink:
    - name: Link Title
      url: "https://learn.microsoft.com/..."
  
  # Option 1: Inline query
  query: |
    resources
    | where type =~ 'Microsoft.Network/networkInterfaces'
    | where properties.virtualMachine == ""
    | project recommendationId="unique-id-001", name, id, tags
  
  # Option 2: External query file
  queryFile: kql/my-query.kql
```

### Categories

Use these for `recommendationControl`:
- `HighAvailability` - Availability and redundancy
- `Security` - Security and access control
- `DisasterRecovery` - Backup and disaster recovery
- `Monitoring` or `Alerting` - Monitoring and alerting
- `Governance` - Resource management and governance
- `Scalability` - Performance and scaling
- `BusinessContinuity` - Business continuity
- Other values default to `OtherBestPractices`

### Impact Levels

- `High` - Critical issues that should be addressed immediately
- `Medium` - Important issues that should be planned
- `Low` - Nice-to-have improvements

## Query Requirements

Your KQL query **must**:

1. **Include `recommendationId` in project**: This links results to the recommendation
   ```kql
   | project recommendationId="your-unique-id", name, id, tags
   ```

2. **Return these standard fields**:
   - `recommendationId` - Your query's aprlGuid
   - `name` - Resource name
   - `id` - Resource ID
   - `tags` - Resource tags (optional)

3. **Optionally include**:
   - `param1`, `param2`, etc. - Additional context shown in reports
   - `resourceGroup` - Explicitly specify resource group
   - `location` - Explicitly specify location

### Example Queries

**Unused Network Interfaces:**
```kql
resources
| where type =~ 'Microsoft.Network/networkInterfaces'
| where properties.virtualMachine == "" or isnull(properties.virtualMachine)
| project recommendationId="yaml-001", name, id, tags,
          param1=strcat("Location: ", location)
```

**Storage Without Encryption:**
```kql
resources
| where type =~ 'Microsoft.Storage/storageAccounts'
| where properties.encryption.services.blob.enabled == false
| project recommendationId="yaml-002", name, id, tags,
          param1=strcat("SKU: ", sku.name)
```

## Installation

### Option 1: User Plugin Directory (Recommended)

```bash
# Copy to user plugin directory
mkdir -p ~/.azqr/plugins
cp custom-checks.yaml ~/.azqr/plugins/
cp -r kql ~/.azqr/plugins/

# Verify plugin is loaded
azqr plugins list
```

### Option 2: Local Plugin Directory

```bash
# Copy to local plugins directory
mkdir -p ./plugins
cp custom-checks.yaml ./plugins/
cp -r kql ./plugins/

# Run azqr from this directory
azqr plugins list
```

### Option 3: Custom Plugin Directory

```bash
# Use AZQR_PLUGIN_DIR environment variable
export AZQR_PLUGIN_DIR=/path/to/your/plugins
azqr plugins list
```

## Usage

Once installed, YAML plugins work like any other scanner:

```bash
# List all plugins (including YAML plugins)
azqr plugins list

# Show plugin details
azqr plugins info example-custom-checks

# Run scan with YAML plugin
azqr scan --subscription-id <sub-id>

# YAML plugin results are included in all output formats:
# - Excel report (recommendations tab)
# - CSV files
# - JSON output
```

## Testing Your Plugin

1. **Validate YAML syntax:**
   ```bash
   # Use any YAML validator
   yamllint custom-checks.yaml
   ```

2. **Test query in Azure Portal:**
   - Go to Azure Portal → Resource Graph Explorer
   - Paste your query
   - Verify it returns expected results

3. **Load and inspect:**
   ```bash
   azqr plugins info your-plugin-name
   ```

4. **Run a test scan:**
   ```bash
   azqr scan --subscription-id <test-sub-id>
   ```

## Best Practices

### Query Design

1. **Be Specific**: Target exact resource types
   ```kql
   | where type =~ 'Microsoft.Network/networkInterfaces'
   ```

2. **Exclude Known Cases**: Filter out expected resources
   ```kql
   | where not(name endswith "-asr")  # Exclude Azure Site Recovery resources
   ```

3. **Handle Nulls**: Check for both empty strings and nulls
   ```kql
   | where properties.field == "" or isnull(properties.field)
   ```

4. **Add Context**: Include helpful parameters
   ```kql
   | project ... param1=strcat("Cost/month: $", properties.cost)
   ```

5. **Optimize Performance**: Use specific type filters early
   ```kql
   resources
   | where type =~ 'Microsoft.Network/publicIPAddresses'  # Filter first
   | where properties.ipConfiguration == ""                # Then check properties
   ```

### Plugin Organization

1. **Group Related Checks**: Put related queries in one plugin
2. **Use Descriptive Names**: Clear, meaningful plugin names
3. **Version Properly**: Follow semantic versioning
4. **Document Well**: Clear descriptions and learn more links
5. **External vs Inline**: Use external files for complex queries

### Naming Conventions

1. **Plugin Names**: `lowercase-with-hyphens`
2. **APR L GUIDs**: `category-number-description` (e.g., `yaml-001-unused-nics`)
3. **KQL Files**: Match the check they implement

## Troubleshooting

### Plugin Not Loaded

```bash
# Check plugin directories
echo $HOME/.azqr/plugins
ls -la ~/.azqr/plugins/

# Enable debug logging
azqr scan --debug
```

### Query Fails

1. **Test in Azure Portal**: Verify query works
2. **Check Permissions**: Ensure Reader access on subscriptions
3. **Review Logs**: Look for error messages
4. **Validate YAML**: Ensure proper formatting

### No Results

1. **Verify Subscriptions**: Check you're scanning correct subscriptions
2. **Test Query**: Run directly in Resource Graph Explorer
3. **Check Filters**: Ensure filters aren't too restrictive

## Examples

See the included `custom-checks.yaml` for complete examples of:
- Inline queries
- External query files
- Different recommendation categories
- Various impact levels
- Multiple parameters

## Advanced Usage

### Multiple Query Files

```yaml
queries:
  - description: Check 1
    queryFile: kql/check1.kql
  - description: Check 2
    queryFile: kql/check2.kql
  - description: Check 3
    queryFile: kql/subfolder/check3.kql
```

### Dynamic Queries

Use KQL functions for reusability:

```kql
let excludedTags = dynamic(["DoNotDelete", "Production"]);
resources
| where type =~ 'Microsoft.Network/publicIPAddresses'
| where not(tags has_any excludedTags)
| project recommendationId="yaml-001", name, id, tags
```

### Cross-Resource Queries

```kql
resources
| where type =~ 'Microsoft.Network/publicIPAddresses'
| where properties.ipConfiguration == ""
| join kind=leftanti (
    resources
    | where type =~ 'Microsoft.Network/natGateways'
    | mvexpand publicIp = properties.publicIpAddresses
    | project publicIpId = tostring(publicIp.id)
) on $left.id == $right.publicIpId
| project recommendationId="yaml-001", name, id, tags
```

## Contributing

To share your YAML plugin with the community:

1. Test thoroughly across multiple subscriptions
2. Document the plugin and queries
3. Create a pull request to the azqr repository
4. Include in `examples/plugins/yaml-example/community/`

## Learn More

- [Azure Resource Graph Query Language](https://learn.microsoft.com/azure/governance/resource-graph/concepts/query-language)
- [KQL Quick Reference](https://learn.microsoft.com/azure/data-explorer/kql-quick-reference)
- [Resource Graph Sample Queries](https://learn.microsoft.com/azure/governance/resource-graph/samples/starter)
- [APRL Recommendations](https://azure.github.io/Azure-Proactive-Resiliency-Library/)
- [Plugin Development Guide](../../../docs/content/en/plugins.md)
