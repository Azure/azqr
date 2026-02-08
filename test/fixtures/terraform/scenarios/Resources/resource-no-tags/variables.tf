variable "resource_group_name" {
  description = "Name of the resource group. If empty, a new one will be created."
  type        = string
  default     = ""
}

variable "storage_account_name" {
  description = "Name of the storage account. If empty, a random name will be generated."
  type        = string
  default     = ""
  validation {
    condition     = var.storage_account_name == "" || (length(var.storage_account_name) >= 3 && length(var.storage_account_name) <= 24 && can(regex("^[a-z0-9]+$", var.storage_account_name)))
    error_message = "Storage account name must be between 3 and 24 characters, and can only contain lowercase letters and numbers."
  }
}

variable "location" {
  description = "Azure region for resources."
  type        = string
  default     = "eastus"
}
