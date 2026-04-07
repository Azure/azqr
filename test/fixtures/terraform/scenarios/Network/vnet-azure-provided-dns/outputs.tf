output "vnet_name" {
  description = "Name of the created virtual network"
  value       = azurerm_virtual_network.azure_provided_dns.name
}

output "vnet_id" {
  description = "ID of the created virtual network"
  value       = azurerm_virtual_network.azure_provided_dns.id
}

output "resource_group_name" {
  description = "Name of the resource group containing the virtual network"
  value       = azurerm_resource_group.test.name
}
