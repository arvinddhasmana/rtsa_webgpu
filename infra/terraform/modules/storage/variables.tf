# CLASSIFICATION: UNCLASSIFIED
# Storage module — env-scoped blob storage for backups / Redpanda tiered storage.

variable "name" {
  type        = string
  description = "Storage account name (3-24 lowercase alphanumeric, globally unique)."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "resource_group_name" {
  type        = string
  description = "Resource group name."
}

variable "account_replication_type" {
  type        = string
  description = "Replication type."
  default     = "LRS"
}

variable "containers" {
  type        = list(string)
  description = "Blob containers to create."
  default     = ["backups", "tiered"]
}

variable "tags" {
  type        = map(string)
  description = "Tags."
  default     = {}
}
