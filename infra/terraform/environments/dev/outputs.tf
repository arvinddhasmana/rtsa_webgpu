# CLASSIFICATION: UNCLASSIFIED
# Dev environment outputs.

output "resource_group" {
  description = "Environment resource group."
  value       = azurerm_resource_group.env.name
}

output "cluster_name" {
  description = "AKS cluster name."
  value       = module.aks.cluster_name
}

output "node_resource_group" {
  description = "AKS-managed node resource group."
  value       = module.aks.node_resource_group
}

output "oidc_issuer_url" {
  description = "AKS OIDC issuer URL."
  value       = module.aks.oidc_issuer_url
}

output "key_vault_uri" {
  description = "Key Vault URI."
  value       = module.keyvault.uri
}

output "acr_login_server" {
  description = "Shared ACR login server."
  value       = data.azurerm_container_registry.shared.login_server
}

output "storage_account_name" {
  description = "Environment storage account."
  value       = module.storage.name
}

output "workload_identity_client_ids" {
  description = "Workload -> managed identity client ID (annotate ServiceAccounts)."
  value       = module.identity.client_ids
}
