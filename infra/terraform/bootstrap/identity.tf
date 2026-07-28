# CLASSIFICATION: UNCLASSIFIED
# GitHub Actions OIDC federation for RTSA.
#
# Design: GitHub federates directly to user-assigned managed identities — NO client
# secrets are ever created or stored. Each environment has its own deployer identity
# (least privilege, gated by GitHub Environments); a separate read-only identity runs
# `terraform plan` on pull_request.

locals {
  environments = toset(var.environments)

  oidc_issuer        = "https://token.actions.githubusercontent.com"
  oidc_audience      = ["api://AzureADTokenExchange"]
  subscription_scope = "/subscriptions/${var.subscription_id}"
}

# ── Per-environment deployer identities (used by gated CD jobs) ───────────────
resource "azurerm_user_assigned_identity" "deployer" {
  for_each = local.environments

  name                = "id-${var.project}-${each.key}-deployer"
  resource_group_name = azurerm_resource_group.shared.name
  location            = azurerm_resource_group.shared.location
  tags                = merge(local.base_tags, { environment = each.key })
}

# Subject is gated on the GitHub Environment, so only jobs targeting that
# environment (with its protection rules) can obtain the token.
resource "azurerm_federated_identity_credential" "deployer_env" {
  for_each = local.environments

  name                = "gha-${each.key}-environment"
  resource_group_name = azurerm_resource_group.shared.name
  parent_id           = azurerm_user_assigned_identity.deployer[each.key].id
  audience            = local.oidc_audience
  issuer              = local.oidc_issuer
  subject             = "repo:${var.github_owner}/${var.github_repo}:environment:${each.key}"
}

resource "azurerm_role_assignment" "deployer_contributor" {
  for_each = local.environments

  scope                            = local.subscription_scope
  role_definition_name             = var.deployer_subscription_role
  principal_id                     = azurerm_user_assigned_identity.deployer[each.key].principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}

# Required so P1 can create role assignments (kubelet->ACR, workloads->Key Vault).
resource "azurerm_role_assignment" "deployer_rbac" {
  for_each = local.environments

  scope                            = local.subscription_scope
  role_definition_name             = var.deployer_rbac_role
  principal_id                     = azurerm_user_assigned_identity.deployer[each.key].principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "deployer_acr_push" {
  for_each = local.environments

  scope                            = azurerm_container_registry.shared.id
  role_definition_name             = "AcrPush"
  principal_id                     = azurerm_user_assigned_identity.deployer[each.key].principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "deployer_state_blob" {
  for_each = local.environments

  scope                            = azurerm_storage_account.tfstate.id
  role_definition_name             = "Storage Blob Data Contributor"
  principal_id                     = azurerm_user_assigned_identity.deployer[each.key].principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}

# ── Shared PR/CI plan identity (read-only; runs `terraform plan` on pull_request) ─
resource "azurerm_user_assigned_identity" "ci" {
  name                = "id-${var.project}-ci-plan"
  resource_group_name = azurerm_resource_group.shared.name
  location            = azurerm_resource_group.shared.location
  tags                = merge(local.base_tags, { environment = "ci" })
}

resource "azurerm_federated_identity_credential" "ci_pull_request" {
  name                = "gha-pull-request"
  resource_group_name = azurerm_resource_group.shared.name
  parent_id           = azurerm_user_assigned_identity.ci.id
  audience            = local.oidc_audience
  issuer              = local.oidc_issuer
  subject             = "repo:${var.github_owner}/${var.github_repo}:pull_request"
}

resource "azurerm_role_assignment" "ci_reader" {
  scope                            = local.subscription_scope
  role_definition_name             = "Reader"
  principal_id                     = azurerm_user_assigned_identity.ci.principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "ci_state_blob_reader" {
  scope                            = azurerm_storage_account.tfstate.id
  role_definition_name             = "Storage Blob Data Reader"
  principal_id                     = azurerm_user_assigned_identity.ci.principal_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}
