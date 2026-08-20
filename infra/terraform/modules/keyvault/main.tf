# CLASSIFICATION: UNCLASSIFIED
# Key Vault module — main.

resource "azurerm_key_vault" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  tenant_id           = var.tenant_id
  sku_name            = var.sku_name

  # RBAC data-plane (no access policies); workloads get "Key Vault Secrets User".
  rbac_authorization_enabled = true

  purge_protection_enabled      = var.purge_protection_enabled
  soft_delete_retention_days    = var.soft_delete_retention_days
  public_network_access_enabled = var.public_network_access_enabled

  network_acls {
    bypass         = "AzureServices"
    default_action = "Allow" # tighten to "Deny" + private endpoints in hardening phase
  }

  tags = var.tags
}

resource "azurerm_role_assignment" "secrets_officer" {
  count = var.secrets_officer_object_id == null ? 0 : 1

  scope                = azurerm_key_vault.this.id
  role_definition_name = "Key Vault Secrets Officer"
  principal_id         = var.secrets_officer_object_id
  principal_type       = "ServicePrincipal"
}
