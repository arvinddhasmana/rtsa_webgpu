# CLASSIFICATION: UNCLASSIFIED
# Remote Terraform state backend (Azure Blob) consumed by every environment stack.
# State locking uses native blob leases — no separate lock table is required.

resource "azurerm_storage_account" "tfstate" {
  name                            = local.state_sa_name
  resource_group_name             = azurerm_resource_group.shared.name
  location                        = azurerm_resource_group.shared.location
  account_tier                    = "Standard"
  account_replication_type        = var.state_replication_type
  account_kind                    = "StorageV2"
  min_tls_version                 = "TLS1_2"
  https_traffic_only_enabled      = true
  allow_nested_items_to_be_public = false

  # Keys stay enabled so the container can be created without a data-plane RBAC
  # race on first run. State ACCESS from CI uses Entra (backend use_azuread_auth).
  # Disable keys + add private endpoints in the hardening phase (P6).
  shared_access_key_enabled     = true
  public_network_access_enabled = true

  blob_properties {
    versioning_enabled = true

    delete_retention_policy {
      days = 30
    }
    container_delete_retention_policy {
      days = 30
    }
  }

  tags = local.base_tags

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_storage_container" "tfstate" {
  name                  = "tfstate"
  storage_account_id    = azurerm_storage_account.tfstate.id
  container_access_type = "private"
}
