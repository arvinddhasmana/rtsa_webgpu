# CLASSIFICATION: UNCLASSIFIED
# Provider configuration for the staging environment.
# Subscription is read from ARM_SUBSCRIPTION_ID (set locally or by GitHub OIDC) —
# no subscription/tenant IDs are committed to source control.

provider "azurerm" {
  # Skip bulk resource provider registration enumeration; only register what
  # this stack actually uses. Prevents the ~2 min silent "hang" during plan.
  resource_provider_registrations = "core"

  features {
    resource_group {
      # Ephemeral environments must destroy cleanly even if stray resources remain.
      prevent_deletion_if_contains_resources = false
    }
    key_vault {
      purge_soft_delete_on_destroy    = true
      recover_soft_deleted_key_vaults = true
    }
  }
}

provider "azurerm" {
  alias = "shared"

  # Shared foundation reads use the shared subscription so ACR / Log Analytics
  # can live outside the environment subscription boundary.
  resource_provider_registrations = "core"

  subscription_id = var.shared_subscription_id

  features {}
}
