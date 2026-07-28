# CLASSIFICATION: UNCLASSIFIED
# Key Vault module — RBAC-authorized vault for TLS certs, JWT secrets, etc.
# Secrets are projected into pods via the Secrets Store CSI driver + Workload Identity.

variable "name" {
  type        = string
  description = "Key Vault name (3-24 chars, globally unique)."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "resource_group_name" {
  type        = string
  description = "Resource group name."
}

variable "tenant_id" {
  type        = string
  description = "Entra ID tenant GUID."
}

variable "sku_name" {
  type        = string
  description = "Key Vault SKU."
  default     = "standard"
}

variable "purge_protection_enabled" {
  type        = bool
  description = "Enable purge protection (true for prod; false lets ephemeral envs be destroyed cleanly)."
  default     = false
}

variable "soft_delete_retention_days" {
  type        = number
  description = "Soft-delete retention window (days)."
  default     = 7
}

variable "public_network_access_enabled" {
  type        = bool
  description = "Allow public network access (lock down with private endpoints in hardening)."
  default     = true
}

variable "tags" {
  type        = map(string)
  description = "Tags."
  default     = {}
}
