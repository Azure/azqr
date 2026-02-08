# Terraform Test Fixtures

This directory contains Terraform configurations that provision Azure infrastructure for integration testing. Each fixture represents either a **compliant** (baseline) or **non-compliant** (scenario) resource configuration.

## Directory Structure

```
terraform/
├── baseline/                            # Compliant resource configurations
│   ├── Resources/                       # Shared infrastructure namespace
│   │   └── resourceGroups/              # Standalone resource group
│   └── Storage/                         # Storage provider namespace
│       └── storageAccounts/             # Compliant storage account with all security best practices
└── scenarios/                           # Non-compliant configurations (violations)
    └── Storage/                         # Storage provider namespace
        └── storageAccounts/             # Storage Account scenarios
            ├── storage-no-https/        # Storage without HTTPS enforcement (st-007)
            ├── storage-public-access/   # Storage with public network access enabled
            └── storage-old-tls/         # Storage with TLS 1.0 (st-009)
```

**Organization**: Fixtures are organized hierarchically by Azure provider namespace (e.g., `Storage/`, `Compute/`) and resource type (e.g., `storageAccounts/`, `virtualMachines/`). This mirrors the structure in `internal/graph/aprl/azure-resources/` and `test/integration/azure-resources/`, providing consistency across the entire project.

## Fixture Types

### Baseline Fixtures

**Purpose**: Test that AZQR correctly identifies compliant resources (no recommendations).

Located in `baseline/`, these fixtures provision resources following Azure security best practices:

#### `baseline/Resources/resourceGroups`
- **Resource**: Azure Resource Group
- **Purpose**: Basic resource group for testing
- **Variables**:
  - `resource_group_name` - Name (auto-generated if empty)
  - `location` - Azure region (default: `eastus`)
  - `tags` - Additional tags

#### `baseline/Storage/storageAccounts`
- **Resource**: Azure Storage Account with security best practices
- **Configuration**:
  - ✅ HTTPS traffic only: `https_traffic_only_enabled = true`
  - ✅ Minimum TLS 1.2: `min_tls_version = "TLS1_2"`
  - ✅ Private network only: `public_network_access_enabled = false`
  - ✅ No public blob containers: `allow_nested_items_to_be_public = false`
  - ✅ Infrastructure encryption enabled
  - ✅ Blob versioning enabled
  - ✅ Soft delete for blobs (7 days)
- **Expected AZQR Result**: Zero recommendations (fully compliant)

### Scenario Fixtures

**Purpose**: Test that AZQR correctly detects specific violations and generates appropriate recommendations.

Located in `scenarios/`, each fixture has one or more intentional misconfigurations:

#### `scenarios/Storage/storageAccounts/storage-no-https`
- **Violation**: HTTPS not enforced
- **Configuration**: `https_traffic_only_enabled = false`
- **Expected Recommendation**: `st-007` - "Storage Account should use HTTPS only"
- **Impact**: High
- **Category**: Security

#### `scenarios/Storage/storageAccounts/storage-public-access`
- **Violation**: Public network access enabled
- **Configuration**:
  - `public_network_access_enabled = true`
  - `allow_nested_items_to_be_public = true`
- **Expected Recommendation**: (AZQR may or may not flag this currently)
- **Impact**: High
- **Category**: Security

#### `scenarios/Storage/storageAccounts/storage-old-tls`
- **Violation**: Old TLS version
- **Configuration**: `min_tls_version = "TLS1_0"`
- **Expected Recommendation**: `st-009` - "Storage Account should enforce TLS >= 1.2"
- **Impact**: Low
- **Category**: Security

## Common Patterns

### Resource Naming

All fixtures use auto-generated names with random suffixes to avoid conflicts:
```hcl
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

locals {
  storage_account_name = "azqrtest${random_string.suffix.result}"
}
```

### Resource Group Handling

Fixtures can either create their own resource group or use an existing one:
```hcl
# Create RG if not provided
resource "azurerm_resource_group" "test" {
  count    = var.resource_group_name == "" ? 1 : 0
  name     = "azqr-test-${random_string.suffix.result}"
  location = var.location
}

locals {
  resource_group_name = var.resource_group_name != "" ? var.resource_group_name : azurerm_resource_group.test[0].name
}
```

### Tagging

All resources are tagged for identification and cost tracking:
```hcl
tags = {
  "Purpose"     = "AZQR Integration Testing"
  "ManagedBy"   = "Terraform"
  "Environment" = "Test"
  "Violation"   = "HTTPS not enforced"  # For scenario fixtures
}
```

## Usage

### Individual Fixture Testing

```bash
cd test/fixtures/terraform/scenarios/Storage/storageAccounts/storage-no-https

# Initialize
terraform init

# Plan with default values
terraform plan

# Apply (creates resources)
terraform apply -auto-approve

# Get outputs
terraform output storage_account_name

# Destroy (cleanup)
terraform destroy -auto-approve
```

### With Custom Variables

```bash
terraform apply -auto-approve \
  -var="location=westus2" \
  -var="resource_group_name=my-test-rg"
```

### Outputs

Each fixture exposes relevant information via outputs:

```bash
$ terraform output
storage_account_name = "azqrtestab12cd34"
storage_account_id = "/subscriptions/.../storageAccounts/azqrtestab12cd34"
resource_group_name = "azqr-test-ab12cd34"
```

## Variables

Common variables across fixtures:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `resource_group_name` | string | `""` | Existing RG name (creates new if empty) |
| `location` | string | `"eastus"` | Azure region |
| `tags` | map(string) | `{}` | Additional tags |
| `storage_account_name` | string | `""` | Storage name (auto-generated if empty) |

## Adding New Fixtures

### Creating a Baseline Fixture

1. Create directory: `baseline/<Provider>/<resourceType>/`
   - Example: `baseline/Compute/virtualMachines/`
   - Example: `baseline/Sql/databases/`

2. Create files:
   - `main.tf` - Resource definition with best practices
   - `variables.tf` - Input variables
   - `outputs.tf` - Resource identifiers and properties

3. Follow security best practices from [Azure Well-Architected Framework](https://learn.microsoft.com/azure/well-architected/)

### Creating a Scenario Fixture

1. Create directory: `scenarios/<Provider>/<resourceType>/<scenario-name>/`
   - Example: `scenarios/Compute/virtualMachines/vm-no-encryption/`
   - Example: `scenarios/Sql/databases/sql-no-tde/`

2. Start from the baseline configuration
3. Intentionally violate one specific best practice
4. Document the violation in:
   - Resource tags
   - Output variables
   - Comments in `main.tf`

5. Reference the expected AZQR recommendation ID from [data/recommendations.json](../../../data/recommendations.json)

### File Template

**main.tf**:
```hcl
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "azurerm" {
  features {}
}

# Resource configuration here
```

**variables.tf**:
```hcl
variable "resource_group_name" {
  description = "Name of the resource group. If empty, a new one will be created."
  type        = string
  default     = ""
}

variable "location" {
  description = "Azure region for resources."
  type        = string
  default     = "eastus"
}
```

**outputs.tf**:
```hcl
output "resource_name" {
  description = "Name of the created resource"
  value       = azurerm_resource.example.name
}

output "resource_id" {
  description = "ID of the created resource"
  value       = azurerm_resource.example.id
}
```

## Maintenance

### Terraform Version Updates

When updating Terraform version requirements:
1. Update `required_version` in all `main.tf` files
2. Test all fixtures with new version
3. Update CI/CD workflow Terraform version

### Azure Provider Updates

When updating azurerm provider:
1. Update version in all `main.tf` files
2. Review [provider changelog](https://github.com/hashicorp/terraform-provider-azurerm/blob/main/CHANGELOG.md)
3. Test for breaking changes in resource arguments

### Validation

Validate all fixtures before committing:

```bash
# Format all Terraform files
make terraform-fmt

# Validate syntax
make terraform-validate

# Or manually
cd test/fixtures/terraform
find . -name "*.tf" -exec terraform fmt {} \;
find . -type d -exec sh -c 'cd "{}" && terraform init -backend=false && terraform validate' \;
```

## Cost Optimization

To minimize Azure costs:

- Use **Standard LRS** replication (cheapest)
- Use **eastus** or **westus** regions (typically lower cost)
- **Destroy resources immediately** after tests
- Run tests in **dedicated test subscription** with spending limits

Estimated costs per resource (monthly, if not deleted):
- Resource Group: Free
- Storage Account (LRS, no data): ~$0.05/month

## Troubleshooting

### Name Conflicts

```
Error: storage account name already exists
```

**Solution**: Storage account names are globally unique. The random suffix should prevent conflicts, but you can override:
```bash
terraform apply -var="storage_account_name=uniquename123"
```

### Authentication Errors

```
Error: building account: Error getting authenticated object ID
```

**Solution**: Authenticate to Azure:
```bash
az login
# Or set service principal environment variables
```

### Quota Limits

```
Error: Quota exceeded for resource type
```

**Solution**: 
- Request quota increase, or
- Use a different Azure region, or
- Clean up existing test resources

## Resources

- [Terraform Azure Provider Docs](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [Azure Storage Account Best Practices](https://learn.microsoft.com/azure/storage/common/storage-security-guide)
- [AZQR Recommendations](../../data/recommendations.json)
- [Terratest Best Practices](https://terratest.gruntwork.io/docs/testing-best-practices/)
