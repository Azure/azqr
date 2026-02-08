output "storage_account_name" {
  description = "Name of the created storage account"
  value       = azurerm_storage_account.no_https.name
}

output "storage_account_id" {
  description = "ID of the created storage account"
  value       = azurerm_storage_account.no_https.id
}

output "resource_group_name" {
  description = "Name of the resource group containing the storage account"
  value       = local.resource_group_name
}

output "https_enabled" {
  description = "Whether HTTPS is enabled (should be false for this scenario)"
  value       = azurerm_storage_account.no_https.https_traffic_only_enabled
}
