variable "resource_group_name" {
  description = "Name of the resource group. If empty, a new one will be created."
  type        = string
  default     = ""
}

variable "location" {
  description = "Azure region for resources."
  type        = string
  default     = "eastus"
}

variable "tags" {
  description = "Additional tags to apply to resources."
  type        = map(string)
  default     = {}
}
