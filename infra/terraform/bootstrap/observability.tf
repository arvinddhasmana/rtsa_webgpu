# CLASSIFICATION: UNCLASSIFIED
# Shared Log Analytics workspace — central logs/metrics sink for all environments.

resource "azurerm_log_analytics_workspace" "shared" {
  name                = local.law_name
  resource_group_name = azurerm_resource_group.shared.name
  location            = azurerm_resource_group.shared.location
  sku                 = "PerGB2018"
  retention_in_days   = var.log_retention_days

  tags = local.base_tags
}
