# CLASSIFICATION: UNCLASSIFIED
# Provider configuration for the dev environment.
# Subscription is read from ARM_SUBSCRIPTION_ID (set locally or by GitHub OIDC) —
# no subscription/tenant IDs are committed to source control.

provider "azurerm" {
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
