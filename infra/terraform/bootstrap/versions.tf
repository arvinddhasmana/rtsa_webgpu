# CLASSIFICATION: UNCLASSIFIED
# Terraform + provider version constraints for the RTSA bootstrap stack.

terraform {
  required_version = ">= 1.9.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.14"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Bootstrap uses LOCAL state by design: it *creates* the remote-state backend
  # (storage account + container) that every environment stack consumes.
  # After the first apply you may optionally migrate this state into the created
  # container — see README.md ("Migrating bootstrap state to remote").
}
