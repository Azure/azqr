output "storage_account_name" {
  description = "Name of the created storage account"
  value       = azurerm_storage_account.old_tls.name
}

output "storage_account_id" {
  description = "ID of the created storage account"
  value       = azurerm_storage_account.old_tls.id
}

output "resource_group_name" {
  description = "Name of the resource group containing the storage account"
  value       = local.resource_group_name
}

output "min_tls_version" {
  description = "Minimum TLS version (should be TLS1_0 for this scenario)"
  value       = azurerm_storage_account.old_tls.min_tls_version
}
