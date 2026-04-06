variable "resource_group_name" {
  description = "Name of the resource group. If empty, a new one will be created."
  type        = string
  default     = ""
}

variable "cluster_name" {
  description = "Name of the Azure Managed Redis cluster. If empty, a random name will be generated."
  type        = string
  default     = ""
}

variable "location" {
  description = "Azure region for resources. Must support Azure Managed Redis (Balanced SKUs)."
  type        = string
  default     = "eastus"
}

variable "tags" {
  description = "Additional tags to apply to resources."
  type        = map(string)
  default     = {}
}
