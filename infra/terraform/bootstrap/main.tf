# CLASSIFICATION: UNCLASSIFIED
# Core locals, naming, and the shared (persistent) resource group for RTSA.

data "azurerm_client_config" "current" {}

resource "random_string" "suffix" {
  length  = 6
  lower   = true
  upper   = false
  numeric = true
  special = false
}

locals {
  # CAF region short codes (extend as needed).
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

  # CAF-aligned resource names.
  shared_rg_name = "rg-${var.project}-shared-${local.loc_short}"
  state_sa_name  = "st${var.project}tf${local.suffix}"
  acr_name       = "acr${var.project}${local.suffix}"
  law_name       = "log-${var.project}-shared-${local.loc_short}"

  # Mandatory tags. Shared resources are persistent (no ttl_hours teardown).
  base_tags = merge({
    project        = var.project
    environment    = "shared"
    managed_by     = "terraform"
    classification = "UNCLASSIFIED"
    lifecycle      = "persistent"
    owner          = "${var.github_owner}/${var.github_repo}"
  }, var.tags)
}

resource "azurerm_resource_group" "shared" {
  name     = local.shared_rg_name
  location = var.location
  tags     = local.base_tags

  lifecycle {
    prevent_destroy = true
  }
}
