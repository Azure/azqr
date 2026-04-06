terraform {
  required_version = ">= 1.5.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
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

resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "azurerm_resource_group" "test" {
  count    = var.resource_group_name == "" ? 1 : 0
  name     = "azqr-test-amr-${random_string.suffix.result}"
  location = var.location

  tags = {
    "Purpose"     = "AZQR Integration Testing"
    "ManagedBy"   = "Terraform"
    "Environment" = "Test"
  }
}

locals {
  resource_group_name = var.resource_group_name != "" ? var.resource_group_name : azurerm_resource_group.test[0].name
  cluster_name        = var.cluster_name != "" ? var.cluster_name : "azqr-amr-${random_string.suffix.result}"
}

# Azure Managed Redis instance with multiple deliberate policy violations.
# Uses the smallest Balanced_B0 SKU to minimise cost during integration tests.
# NOTE: provisioning takes approximately 10-20 minutes; plan test timeouts accordingly.
#
# NOTE: database-level properties (accessKeysAuthentication, clientProtocol, persistence)
# are on the child resource Microsoft.Cache/redisEnterprise/databases which is NOT indexed
# by Azure Resource Graph. Those rules (redis-017, 018, 019) are therefore documented as
# automationAvailable: false and not covered by integration tests.
resource "azurerm_managed_redis" "violations" {
  name                = local.cluster_name
  resource_group_name = local.resource_group_name
  location            = var.location
  sku_name            = "Balanced_B0"

  # VIOLATION: HA disabled — triggers redis-013
  high_availability_enabled = false

  # VIOLATION: public network access left Enabled (default) — triggers redis-014
  # VIOLATION: no private endpoints (default: none) — triggers redis-015
  # VIOLATION: no customer-managed key (default: none) — triggers redis-016
  # VIOLATION: no zones (default: none) — triggers redis-012

  default_database {}

  tags = merge(
    var.tags,
    {
      "Purpose"     = "AZQR Integration Testing - AMR Violations"
      "ManagedBy"   = "Terraform"
      "Environment" = "Test"
    }
  )
}


