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
  name     = "azqr-test-${random_string.suffix.result}"
  location = var.location

  tags = {
    "Purpose"     = "AZQR Integration Testing"
    "ManagedBy"   = "Terraform"
    "Environment" = "Test"
  }
}

locals {
  resource_group_name  = var.resource_group_name != "" ? var.resource_group_name : azurerm_resource_group.test[0].name
  storage_account_name = var.storage_account_name != "" ? var.storage_account_name : "azqrtest${random_string.suffix.result}"
}

# Storage account with old TLS version (VIOLATION: should trigger AZQR recommendation)
resource "azurerm_storage_account" "old_tls" {
  name                     = local.storage_account_name
  resource_group_name      = local.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  https_traffic_only_enabled = true
  # VIOLATION: Using old TLS version (should be TLS1_2 minimum)
  min_tls_version                 = "TLS1_0"
  public_network_access_enabled   = false
  allow_nested_items_to_be_public = false

  tags = merge(
    var.tags,
    {
      "Purpose"     = "AZQR Integration Testing - TLS Violation"
      "ManagedBy"   = "Terraform"
      "Environment" = "Test"
      "Violation"   = "Old TLS version (TLS1_0)"
    }
  )
}
