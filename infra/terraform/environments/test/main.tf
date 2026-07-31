# CLASSIFICATION: UNCLASSIFIED
# Test environment — resource group + landing-zone module composition.

data "azurerm_client_config" "current" {}

data "azurerm_subscription" "current" {}

data "azurerm_container_registry" "shared" {
  provider = azurerm.shared

  name                = var.acr_name
  resource_group_name = var.shared_resource_group
}

data "azurerm_log_analytics_workspace" "shared" {
  provider = azurerm.shared

  name                = var.law_name
  resource_group_name = var.shared_resource_group
}

resource "random_string" "suffix" {
  length  = 6
  lower   = true
  upper   = false
  numeric = true
  special = false
}

locals {
  name_prefix = "${var.project}-${var.environment}"

  region_short = {
    canadacentral = "cc"
    canadaeast    = "ce"
    eastus        = "eus"
    eastus2       = "eus2"
    westus2       = "wus2"
    westus3       = "wus3"
    westeurope    = "weu"
    northeurope   = "neu"
  }
  loc_short = lookup(local.region_short, var.location, "reg")
  suffix    = random_string.suffix.result

  rg_name = "rg-${var.project}-${var.environment}-${local.loc_short}"
  kv_name = "kv-${var.project}-${var.environment}-${local.suffix}"
  sa_name = "st${var.project}${var.environment}${local.suffix}"

  base_tags = merge({
    project        = var.project
    environment    = var.environment
    managed_by     = "terraform"
    classification = "UNCLASSIFIED"
    ttl_hours      = tostring(var.ttl_hours)
    owner          = "rtsa"
  }, var.tags)
}

resource "azurerm_resource_group" "env" {
  name     = local.rg_name
  location = var.location
  tags     = local.base_tags

  lifecycle {
    precondition {
      condition     = data.azurerm_subscription.current.subscription_id == var.expected_subscription_id
      error_message = "Environment deployment subscription mismatch. Set ARM_SUBSCRIPTION_ID to the expected environment subscription before running Terraform for this root."
    }
  }
}

module "network" {
  source = "../../modules/network"

  name_prefix         = local.name_prefix
  location            = var.location
  resource_group_name = azurerm_resource_group.env.name
  vnet_address_space  = var.vnet_address_space
  subnet_aks_prefix   = var.subnet_aks_prefix
  subnet_pe_prefix    = var.subnet_pe_prefix
  tags                = local.base_tags
}

module "aks" {
  source = "../../modules/aks"

  name                       = "aks-${local.name_prefix}"
  location                   = var.location
  resource_group_name        = azurerm_resource_group.env.name
  dns_prefix                 = local.name_prefix
  kubernetes_version         = var.kubernetes_version
  sku_tier                   = var.aks_sku_tier
  aks_subnet_id              = module.network.aks_subnet_id
  log_analytics_workspace_id = data.azurerm_log_analytics_workspace.shared.id

  system_node_vm_size   = var.system_node_vm_size
  system_node_min_count = var.system_node_min_count
  system_node_max_count = var.system_node_max_count
  system_node_zones     = var.system_node_zones
  additional_node_pools = var.additional_node_pools

  enable_istio           = var.enable_istio
  istio_revisions        = var.istio_revisions
  enable_keda            = var.enable_keda
  admin_group_object_ids = var.admin_group_object_ids

  tags = local.base_tags
}

module "acr_access" {
  source = "../../modules/acr-access"

  acr_id            = data.azurerm_container_registry.shared.id
  kubelet_object_id = module.aks.kubelet_identity_object_id
}

module "keyvault" {
  source = "../../modules/keyvault"

  name                     = local.kv_name
  location                 = var.location
  resource_group_name      = azurerm_resource_group.env.name
  tenant_id                = data.azurerm_client_config.current.tenant_id
  purge_protection_enabled = var.key_vault_purge_protection
  tags                     = local.base_tags
}

module "identity" {
  source = "../../modules/identity"

  name_prefix         = local.name_prefix
  location            = var.location
  resource_group_name = azurerm_resource_group.env.name
  oidc_issuer_url     = module.aks.oidc_issuer_url
  key_vault_id        = module.keyvault.id
  workloads           = var.workloads
  tags                = local.base_tags
}

module "storage" {
  source = "../../modules/storage"

  name                = local.sa_name
  location            = var.location
  resource_group_name = azurerm_resource_group.env.name
  tags                = local.base_tags
}

module "observability" {
  source = "../../modules/observability"

  name_prefix               = local.name_prefix
  location                  = var.location
  resource_group_name       = azurerm_resource_group.env.name
  enable_managed_prometheus = var.enable_managed_prometheus
  enable_grafana            = var.enable_grafana
  grafana_admin_object_ids  = var.grafana_admin_object_ids
  tags                      = local.base_tags
}

module "dns_tls" {
  source = "../../modules/dns-tls"

  create_records          = var.create_dns_records
  dns_zone_name           = var.dns_zone_name
  dns_zone_resource_group = var.dns_zone_resource_group
  records                 = var.dns_records
}
