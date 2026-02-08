output "storage_account_name" {
  description = "Name of the created storage account"
  value       = azurerm_storage_account.no_immutable_versioning.name
}

output "storage_account_id" {
  description = "ID of the created storage account"
  value       = azurerm_storage_account.no_immutable_versioning.id
}

output "resource_group_name" {
  description = "Name of the resource group containing the storage account"
  value       = local.resource_group_name
}
