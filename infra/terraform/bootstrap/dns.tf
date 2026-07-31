# CLASSIFICATION: UNCLASSIFIED
# Optional shared public DNS zone (enable with create_dns_zone = true).

resource "azurerm_dns_zone" "shared" {
  count               = var.create_dns_zone ? 1 : 0
  name                = var.dns_zone_name
  resource_group_name = azurerm_resource_group.shared.name
  tags                = local.base_tags
}
