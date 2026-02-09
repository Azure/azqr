output "storage_account_with_diag_name" {
  description = "Name of the storage account with diagnostic settings enabled"
  value       = azurerm_storage_account.with_diag.name
}

output "storage_account_without_diag_name" {
  description = "Name of the storage account without diagnostic settings"
  value       = azurerm_storage_account.without_diag.name
}

output "resource_group_name" {
  description = "Name of the resource group containing the storage accounts"
  value       = local.resource_group_name
}

output "log_analytics_workspace_id" {
  description = "ID of the Log Analytics workspace used for diagnostic settings"
  value       = azurerm_log_analytics_workspace.test.id
}
