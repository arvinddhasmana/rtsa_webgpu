# CLASSIFICATION: UNCLASSIFIED
# Shared Azure Container Registry for RTSA service + web images.

resource "azurerm_container_registry" "shared" {
  name                          = local.acr_name
  resource_group_name           = azurerm_resource_group.shared.name
  location                      = azurerm_resource_group.shared.location
  sku                           = var.acr_sku
  admin_enabled                 = false # pull/push via Entra (AcrPull/AcrPush), never admin creds
  public_network_access_enabled = true  # restrict + private endpoint in hardening phase

  tags = local.base_tags
}
