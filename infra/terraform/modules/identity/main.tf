# CLASSIFICATION: UNCLASSIFIED
# Workload Identity module — per-workload user-assigned identities federated to
# Kubernetes service accounts via the AKS OIDC issuer. Optionally grants each the
# "Key Vault Secrets User" role so pods can read secrets/certs via the CSI driver.

variable "name_prefix" {
  type        = string
  description = "Identity name prefix, e.g. rtsa-dev."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "resource_group_name" {
  type        = string
  description = "Resource group for the identities."
}

variable "oidc_issuer_url" {
  type        = string
  description = "AKS OIDC issuer URL."
}

variable "key_vault_id" {
  type        = string
  description = "Key Vault resource ID for Secrets User role assignments."
}

variable "workloads" {
  description = "Workloads needing an Azure identity, keyed by short name."
  type = map(object({
    namespace              = string
    service_account        = string
    key_vault_secrets_user = optional(bool, true)
  }))
  default = {}
}

variable "tags" {
  type        = map(string)
  description = "Tags."
  default     = {}
}

resource "azurerm_user_assigned_identity" "workload" {
  for_each = var.workloads

  name                = "id-${var.name_prefix}-${each.key}"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = var.tags
}

resource "azurerm_federated_identity_credential" "workload" {
  for_each = var.workloads

  name                = "fic-${each.key}"
  resource_group_name = var.resource_group_name
  parent_id           = azurerm_user_assigned_identity.workload[each.key].id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = var.oidc_issuer_url
  subject             = "system:serviceaccount:${each.value.namespace}:${each.value.service_account}"
}

resource "azurerm_role_assignment" "kv_secrets_user" {
  for_each = { for k, v in var.workloads : k => v if v.key_vault_secrets_user }

  scope                            = var.key_vault_id
  role_definition_name             = "Key Vault Secrets User"
  principal_id                     = azurerm_user_assigned_identity.workload[each.key].principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}

output "client_ids" {
  description = "Map of workload -> managed identity client ID (annotate the K8s ServiceAccount)."
  value       = { for k, id in azurerm_user_assigned_identity.workload : k => id.client_id }
}

output "principal_ids" {
  description = "Map of workload -> managed identity principal ID."
  value       = { for k, id in azurerm_user_assigned_identity.workload : k => id.principal_id }
}
