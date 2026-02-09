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

resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# Create resource group if not provided
resource "azurerm_resource_group" "test" {
  count    = var.resource_group_name == "" ? 1 : 0
  name     = "azqr-test-diag-${random_string.suffix.result}"
  location = var.location

  tags = {
    "Purpose"     = "AZQR Integration Testing"
    "ManagedBy"   = "Terraform"
    "Environment" = "Test"
  }
}

locals {
  resource_group_name = var.resource_group_name != "" ? var.resource_group_name : azurerm_resource_group.test[0].name
}

# Storage account WITH diagnostic settings enabled (compliant - should NOT trigger st-001)
resource "azurerm_storage_account" "with_diag" {
  name                     = "azqrwdiag${random_string.suffix.result}"
  resource_group_name      = local.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = merge(
    var.tags,
    {
      "Purpose"     = "AZQR Integration Testing - With Diagnostic Settings"
      "ManagedBy"   = "Terraform"
      "Environment" = "Test"
    }
  )
}

# Storage account WITHOUT diagnostic settings (violation - should trigger st-001)
resource "azurerm_storage_account" "without_diag" {
  name                     = "azqrnodiag${random_string.suffix.result}"
  resource_group_name      = local.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = merge(
    var.tags,
    {
      "Purpose"     = "AZQR Integration Testing - Without Diagnostic Settings"
      "ManagedBy"   = "Terraform"
      "Environment" = "Test"
      "Violation"   = "No diagnostic settings"
    }
  )
}

# Log Analytics workspace to serve as the diagnostic settings destination
resource "azurerm_log_analytics_workspace" "test" {
  name                = "azqr-law-${random_string.suffix.result}"
  resource_group_name = local.resource_group_name
  location            = var.location
  sku                 = "PerGB2018"
  retention_in_days   = 30

  tags = {
    "Purpose"     = "AZQR Integration Testing"
    "ManagedBy"   = "Terraform"
    "Environment" = "Test"
  }
}

# Diagnostic settings for the compliant storage account
resource "azurerm_monitor_diagnostic_setting" "with_diag" {
  name                       = "azqr-diag-setting"
  target_resource_id         = azurerm_storage_account.with_diag.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.test.id

  metric {
    category = "Transaction"
    enabled  = true
  }
}
