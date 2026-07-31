# CLASSIFICATION: UNCLASSIFIED
# Terraform + provider constraints and remote backend for the production environment.

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

  # Partial backend — coordinates are supplied via `-backend-config` (see the
  # infra/terraform Makefile `env-init` target, which reads them from bootstrap).
  backend "azurerm" {}
}
