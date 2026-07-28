# CLASSIFICATION: UNCLASSIFIED
# Input variables for the RTSA bootstrap stack.

variable "subscription_id" {
  type        = string
  description = "Azure subscription ID (GUID) that will host RTSA shared + environment resources."

  validation {
    condition     = can(regex("^[0-9a-fA-F-]{36}$", var.subscription_id))
    error_message = "subscription_id must be a 36-character GUID."
  }
}

variable "tenant_id" {
  type        = string
  description = "Entra ID tenant GUID. Leave empty to use the tenant of the current az login."
  default     = ""
}

variable "location" {
  type        = string
  description = "Primary Azure region for shared resources."
  default     = "canadacentral"
}

variable "project" {
  type        = string
  description = "Short project token used in resource names (lowercase alphanumeric)."
  default     = "rtsa"

  validation {
    condition     = can(regex("^[a-z0-9]{2,8}$", var.project))
    error_message = "project must be 2-8 lowercase alphanumeric characters."
  }
}

variable "github_owner" {
  type        = string
  description = "GitHub organization or user that owns the RTSA repository (for OIDC subjects)."
}

variable "github_repo" {
  type        = string
  description = "GitHub repository name (for OIDC federated-credential subjects)."
}

variable "environments" {
  type        = list(string)
  description = "Environments to provision CI/CD deployer identities for."
  default     = ["dev", "staging", "prod"]
}

variable "state_replication_type" {
  type        = string
  description = "Replication for the Terraform state storage account."
  default     = "GRS"

  validation {
    condition     = contains(["LRS", "ZRS", "GRS", "RAGRS", "GZRS", "RAGZRS"], var.state_replication_type)
    error_message = "state_replication_type must be a valid Azure Storage replication option."
  }
}

variable "acr_sku" {
  type        = string
  description = "SKU for the shared Azure Container Registry."
  default     = "Standard"

  validation {
    condition     = contains(["Basic", "Standard", "Premium"], var.acr_sku)
    error_message = "acr_sku must be Basic, Standard or Premium."
  }
}

variable "log_retention_days" {
  type        = number
  description = "Retention for the shared Log Analytics workspace (days)."
  default     = 30

  validation {
    condition     = var.log_retention_days >= 30 && var.log_retention_days <= 730
    error_message = "log_retention_days must be between 30 and 730."
  }
}

variable "deployer_subscription_role" {
  type        = string
  description = "Built-in role granted to per-environment deployer identities for resource management."
  default     = "Contributor"
}

variable "deployer_rbac_role" {
  type        = string
  description = "Role that lets deployers create role assignments (required by the P1 landing zone). Tighten to a constrained 'Role Based Access Control Administrator' for enterprise."
  default     = "User Access Administrator"
}

variable "create_dns_zone" {
  type        = bool
  description = "Whether to create a shared public DNS zone."
  default     = false
}

variable "dns_zone_name" {
  type        = string
  description = "FQDN of the public DNS zone to create when create_dns_zone = true."
  default     = ""
}

variable "tags" {
  type        = map(string)
  description = "Additional tags merged onto all shared resources."
  default     = {}
}
