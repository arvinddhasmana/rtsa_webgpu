# CLASSIFICATION: UNCLASSIFIED
# Provider configuration for the RTSA bootstrap stack.

provider "azurerm" {
  features {}

  # Keep provider registration bounded; legacy/full registration can take many
  # minutes and looks like a hang during `terraform plan`.
  resource_provider_registrations = "core"

  subscription_id = var.subscription_id
  tenant_id       = var.tenant_id != "" ? var.tenant_id : null
}
