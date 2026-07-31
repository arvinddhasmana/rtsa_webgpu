# CLASSIFICATION: UNCLASSIFIED
# Outputs consumed by environment stacks + GitHub Actions configuration.

output "shared_resource_group" {
  description = "Name of the shared (persistent) resource group."
  value       = local.shared_rg_name
}

output "state_storage_account_name" {
  description = "Storage account holding Terraform remote state."
  value       = local.state_sa_name
}

output "state_container_name" {
  description = "Blob container for Terraform state files."
  value       = "tfstate"
}

output "acr_login_server" {
  description = "Login server of the shared Azure Container Registry."
  value       = "${local.acr_name}.azurecr.io"
}

output "acr_name" {
  description = "Name of the shared Azure Container Registry."
  value       = local.acr_name
}

output "log_analytics_workspace_id" {
  description = "Resource ID of the shared Log Analytics workspace."
  value       = "/subscriptions/${var.subscription_id}/resourceGroups/${local.shared_rg_name}/providers/Microsoft.OperationalInsights/workspaces/${local.law_name}"
}

output "tenant_id" {
  description = "Entra ID tenant ID (for azure/login)."
  value       = data.azurerm_client_config.current.tenant_id
}

output "subscription_id" {
  description = "Azure subscription ID that hosts the shared foundation and environment backends."
  value       = var.subscription_id
}

output "environment_subscription_ids" {
  description = "Map of environment name to deployment subscription ID. Falls back to the shared subscription when no per-environment mapping is provided."
  value = {
    for env in var.environments : env => lookup(var.environment_subscription_ids, env, var.subscription_id)
  }
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
    AZURE_TENANT_ID         = data.azurerm_client_config.current.tenant_id
    AZURE_SUBSCRIPTION_ID   = var.subscription_id
    ACR_LOGIN_SERVER        = "${local.acr_name}.azurecr.io"
    TFSTATE_RESOURCE_GROUP  = local.shared_rg_name
    TFSTATE_STORAGE_ACCOUNT = local.state_sa_name
    TFSTATE_CONTAINER       = "tfstate"

    # Legacy aliases retained for compatibility with older scripts.
    TFSTATE_RG = local.shared_rg_name
    TFSTATE_SA = local.state_sa_name
  }
}
