output "cluster_name" {
  description = "Name of the Azure Managed Redis instance"
  value       = azurerm_managed_redis.violations.name
}

output "cluster_id" {
  description = "Resource ID of the Azure Managed Redis instance"
  value       = azurerm_managed_redis.violations.id
}

output "resource_group_name" {
  description = "Name of the resource group containing the instance"
  value       = local.resource_group_name
}

output "high_availability_enabled" {
  description = "Whether high availability is enabled (should be false for the violation scenario)"
  value       = azurerm_managed_redis.violations.high_availability_enabled
}

output "location" {
  description = "Azure region where the instance is deployed"
  value       = azurerm_managed_redis.violations.location
}
