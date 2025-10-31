---
title: Plugins
description: Documentation for creating and using YAML plugins in azqr
weight: 1
---

## Overview

YAML plugins provide a simple, declarative way to extend azqr with custom Azure Resource Graph (ARG) queries without writing Go code. This is ideal for:
- Quick custom checks and validations
- Organization-specific compliance rules
- Temporary or experimental recommendations
- Non-developers who want to extend azqr

## Plugin Structure

A YAML plugin consists of a single `.yaml` file containing plugin metadata and one or more Azure Resource Graph queries.

### Basic Structure

```yaml
name: plugin-name
version: 1.0.0
description: Brief description of what this plugin does
author: Plugin Author Name (optional)
license: MIT (optional)

queries:
  - aprlGuid: unique-id-001
    description: Short description of the recommendation
    longDescription: |
      Detailed description of the issue and why it matters.
      Can be multiple lines.
    recommendationControl: Category Name
    recommendationImpact: Impact Level
    recommendationResourceType: Microsoft.Service/resourceType
    learnMoreLink:
      - name: Documentation Title
        url: https://learn.microsoft.com/...
    query: |
      resources
      | where type =~ 'microsoft.service/resourcetype'
      | where some_condition == true
      | project id, name, resourceGroup, location
```

### Required Fields

#### Plugin Level
- **name**: Unique plugin identifier (string)
- **version**: Semantic version (e.g., "1.0.0")
- **description**: What the plugin does

#### Query Level
- **aprlGuid**: Unique identifier for the recommendation
- **description**: Short recommendation title
- **recommendationControl**: Category (see Categories section)
- **recommendationImpact**: Impact level (High, Medium, Low)
- **recommendationResourceType**: Azure resource type (e.g., "Microsoft.Storage/storageAccounts")
- **query** OR **queryFile**: The KQL query to execute

### Optional Fields

#### Plugin Level
- **author**: Plugin author name
- **license**: License type (e.g., MIT, Apache-2.0)

#### Query Level
- **longDescription**: Detailed explanation
- **learnMoreLink**: Array of documentation links
- **recommendationTypeId**: Azure Policy or APRL recommendation ID
- **recommendationMetadataState**: State (Active, Deprecated)
- **potentialBenefits**: Benefits of implementing the recommendation
- **pgVerified**: Whether verified by product group (boolean)
- **automationAvailable**: Whether automation is available (boolean)
- **tags**: Array of tags for categorization

## Categories

Valid values for `recommendationControl`:
- **High Availability**: Availability and redundancy
- **Security**: Security and access control
- **Disaster Recovery**: Backup and recovery
- **Scalability**: Scaling and performance
- **Governance**: Compliance and governance
- **Monitoring and Alerting**: Observability
- **Business Continuity**: Business continuity planning
- **Service Upgrade and Retirement**: Service lifecycle
- **Other Best Practices**: General best practices (default)

## Impact Levels

Valid values for `recommendationImpact`:
- **High**: Critical issues that should be addressed immediately
- **Medium**: Important issues that should be addressed soon (default)
- **Low**: Nice-to-have improvements

## Query Format

### Inline Query

Include the KQL query directly in the YAML file:

```yaml
queries:
  - aprlGuid: example-001
    description: Example recommendation
    recommendationControl: Security
    recommendationImpact: High
    recommendationResourceType: Microsoft.Storage/storageAccounts
    query: |
      resources
      | where type =~ 'microsoft.storage/storageaccounts'
      | where properties.supportsHttpsTrafficOnly == false
      | project id, name, resourceGroup, location
```

### External Query File

Reference an external `.kql` file:

```yaml
queries:
  - aprlGuid: example-002
    description: Example with external query
    recommendationControl: Governance
    recommendationImpact: Medium
    recommendationResourceType: Microsoft.Network/publicIPAddresses
    queryFile: ./kql/unused-public-ips.kql
```

The path is relative to the YAML file location. Create a `kql/` subdirectory next to your plugin YAML file.

## Query Requirements

Your Azure Resource Graph queries must:

1. **Return resources**: Query the `resources` table
2. **Project required fields**: Include at minimum:
   - `id`: Resource ID
   - `name`: Resource name
   - `resourceGroup`: Resource group name
   - `location`: Azure region

3. **Filter appropriately**: Use `where` clauses to identify non-compliant resources
4. **Use case-insensitive comparisons**: Use `=~` instead of `==` for type comparisons

### Query Example

```kql
resources
| where type =~ 'microsoft.network/networkinterfaces'
| where properties.virtualMachine == "" or isnull(properties.virtualMachine)
| where properties.privateEndpoint == "" or isnull(properties.privateEndpoint)
| project id, name, resourceGroup, location,
          tags,
          sku = properties.ipConfigurations[0].properties.privateIPAllocationMethod
```

## Plugin Discovery

YAML plugins are discovered from the following locations:

1. **Current directory**: `./plugins/*.yaml`
2. **User plugins directory**: `~/.azqr/plugins/*.yaml`
3. **System plugins directory**: `/etc/azqr/plugins/*.yaml` (Linux/macOS)

azqr searches recursively in these directories for any `.yaml` or `.yml` files.

## Complete Example

Here's a complete example plugin (`custom-checks.yaml`):

```yaml
name: example-custom-checks
version: 1.0.0
description: Example YAML plugin with custom Azure Resource Graph queries
author: Azure Quick Review Team
license: MIT

queries:
  # Check for unused network interfaces
  - description: Network interfaces not attached to any VM
    aprlGuid: yaml-001-unused-nics
    recommendationTypeId: null
    recommendationControl: Governance
    recommendationImpact: Low
    recommendationResourceType: Microsoft.Network/networkInterfaces
    recommendationMetadataState: Active
    longDescription: |
      Network interfaces that are not attached to any virtual machine.
      These resources incur costs and should be reviewed for cleanup.
    potentialBenefits: Cost optimization and resource cleanup
    pgVerified: false
    automationAvailable: false
    tags:
      - cost-optimization
      - cleanup
    learnMoreLink:
      - name: Network Interface Overview
        url: "https://learn.microsoft.com/azure/virtual-network/virtual-network-network-interface"
    query: |
      resources
      | where type =~ 'Microsoft.Network/networkInterfaces'
      | where properties.virtualMachine == "" or isnull(properties.virtualMachine)
      | where properties.privateEndpoint == "" or isnull(properties.privateEndpoint)
      | project id, name, resourceGroup, location, tags

  # Check for unused public IPs (from external file)
  - description: Public IP addresses not associated with any resource
    aprlGuid: yaml-002-unused-public-ips
    recommendationControl: Governance
    recommendationImpact: Medium
    recommendationResourceType: Microsoft.Network/publicIPAddresses
    longDescription: |
      Public IP addresses that are not associated with any Azure resource.
      These IPs cost money even when not in use.
    learnMoreLink:
      - name: Public IP Addresses
        url: "https://learn.microsoft.com/azure/virtual-network/ip-services/public-ip-addresses"
    queryFile: kql/unused-public-ips.kql

  # Security check
  - description: Storage accounts without secure transfer enabled
    aprlGuid: yaml-003-storage-secure-transfer
    recommendationControl: Security
    recommendationImpact: High
    recommendationResourceType: Microsoft.Storage/storageAccounts
    longDescription: |
      Storage accounts that do not have secure transfer (HTTPS) required.
      This is a security risk as data can be transmitted over insecure connections.
    potentialBenefits: Improved security and data protection
    learnMoreLink:
      - name: Require secure transfer
        url: "https://learn.microsoft.com/azure/storage/common/storage-require-secure-transfer"
    query: |
      resources
      | where type =~ 'Microsoft.Storage/storageAccounts'
      | where properties.supportsHttpsTrafficOnly == false
      | project id, name, resourceGroup, location,
                sku = sku.name,
                tier = sku.tier
```

## Usage

Once you've created your YAML plugin:

1. **Place the file** in one of the plugin directories
2. **Run azqr scan** as normal:
   ```bash
   azqr scan
   ```

3. **View plugin info**:
   ```bash
   azqr plugins list
   azqr plugins info <plugin-name>
   ```

The recommendations from your YAML plugin will be included in all outputs (Excel, CSV, JSON) alongside built-in recommendations.

## Best Practices

### 1. Use Descriptive Names
```yaml
name: org-security-checks
description: Organization-specific security compliance checks
```

### 2. Group Related Checks
Put related recommendations in the same plugin file:
```yaml
name: network-optimization
queries:
  - aprlGuid: net-001-unused-nics
    description: Unused network interfaces
    ...
  - aprlGuid: net-002-unused-ips
    description: Unused public IPs
    ...
```

### 3. Version Your Plugins
Follow semantic versioning:
- **1.0.0**: Initial release
- **1.1.0**: Add new queries
- **2.0.0**: Breaking changes

### 4. Provide Learn More Links
Always include documentation links:
```yaml
learnMoreLink:
  - name: Official Documentation
    url: https://learn.microsoft.com/...
  - name: Best Practices Guide
    url: https://learn.microsoft.com/...
```

### 5. Test Your Queries
Test queries in Azure Resource Graph Explorer first:
- https://portal.azure.com/#view/HubsExtension/ArgQueryBlade

### 6. Use External Files for Complex Queries
For queries over ~10 lines, use external `.kql` files:
```
my-plugin/
├── custom-checks.yaml
└── kql/
    ├── query1.kql
    ├── query2.kql
    └── query3.kql
```

### 7. Document Your Plugin
Include comprehensive descriptions:
```yaml
longDescription: |
  This check identifies resources that...
  
  Why it matters:
  - Cost implications
  - Security risks
  - Performance impact
  
  How to fix:
  1. Step one
  2. Step two
```

## Troubleshooting

### Plugin Not Discovered

1. Check the file extension (`.yaml` or `.yml`)
2. Verify the file is in a plugin directory
3. Run with debug logging:
   ```bash
   azqr scan --debug
   ```

### Query Errors

1. **Syntax errors**: Test the query in Azure Resource Graph Explorer
2. **No results**: Verify the resource type filter
3. **Permission errors**: Ensure you have Reader access to subscriptions

### Invalid YAML

Use a YAML validator to check syntax:
```bash
yamllint custom-checks.yaml
```

## Limitations

1. **Query-based only**: YAML plugins can only use Azure Resource Graph queries, not ARM API calls
3. **Subscription scope**: Queries run within subscription context
4. **No custom logic**: Cannot include complex evaluation logic (use built-in plugins for that)

## Migration from Graph Queries

If you have existing ARG queries, convert them to YAML plugins:

**Before** (separate .kql files):
```kql
resources
| where type =~ 'microsoft.storage/storageaccounts'
| where properties.supportsHttpsTrafficOnly == false
```

**After** (YAML plugin):
```yaml
name: my-checks
version: 1.0.0
description: My custom checks
queries:
  - aprlGuid: check-001
    description: Storage accounts without HTTPS
    recommendationControl: Security
    recommendationImpact: High
    recommendationResourceType: Microsoft.Storage/storageAccounts
    query: |
      resources
      | where type =~ 'microsoft.storage/storageaccounts'
      | where properties.supportsHttpsTrafficOnly == false
      | project id, name, resourceGroup, location
```
