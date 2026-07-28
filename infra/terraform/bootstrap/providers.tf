# CLASSIFICATION: UNCLASSIFIED
# Provider configuration for the RTSA bootstrap stack.

provider "azurerm" {
  features {}

  subscription_id = var.subscription_id
  tenant_id       = var.tenant_id != "" ? var.tenant_id : null
}
