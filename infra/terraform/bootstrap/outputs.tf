# CLASSIFICATION: UNCLASSIFIED
# Outputs consumed by environment stacks + GitHub Actions configuration.

output "shared_resource_group" {
  description = "Name of the shared (persistent) resource group."
  value       = azurerm_resource_group.shared.name
}

output "state_storage_account_name" {
  description = "Storage account holding Terraform remote state."
  value       = azurerm_storage_account.tfstate.name
}

output "state_container_name" {
  description = "Blob container for Terraform state files."
  value       = azurerm_storage_container.tfstate.name
}

output "acr_login_server" {
  description = "Login server of the shared Azure Container Registry."
  value       = azurerm_container_registry.shared.login_server
}

output "acr_name" {
  description = "Name of the shared Azure Container Registry."
  value       = azurerm_container_registry.shared.name
}

output "log_analytics_workspace_id" {
  description = "Resource ID of the shared Log Analytics workspace."
  value       = azurerm_log_analytics_workspace.shared.id
}

output "tenant_id" {
  description = "Entra ID tenant ID (for azure/login)."
  value       = data.azurerm_client_config.current.tenant_id
}

output "deployer_identity_client_ids" {
  description = "Map of environment -> deployer managed-identity client ID (set as the AZURE_CLIENT_ID secret in the matching GitHub Environment)."
  value       = { for env, id in azurerm_user_assigned_identity.deployer : env => id.client_id }
}

output "ci_plan_identity_client_id" {
  description = "Client ID of the PR/CI plan identity (pull_request OIDC)."
  value       = azurerm_user_assigned_identity.ci.client_id
}

output "github_actions_variables" {
  description = "Repository-level variables/secrets to configure for the pipelines."
  value = {
    AZURE_TENANT_ID       = data.azurerm_client_config.current.tenant_id
    AZURE_SUBSCRIPTION_ID = var.subscription_id
    ACR_LOGIN_SERVER      = azurerm_container_registry.shared.login_server
    TFSTATE_RG            = azurerm_resource_group.shared.name
    TFSTATE_SA            = azurerm_storage_account.tfstate.name
    TFSTATE_CONTAINER     = azurerm_storage_container.tfstate.name
  }
}
