# CLASSIFICATION: UNCLASSIFIED
# RTSA Infrastructure — Terraform lifecycle targets (bootstrap + ephemeral environments).
#
# Usage:
#   make bootstrap-up                         # one-time shared foundation (run by an admin)
#   make env-up ENV=dev                       # create/refresh an environment
#   make env-plan ENV=dev
#   make env-down ENV=dev                     # terraform destroy (clean teardown)
#   make env-nuke ENV=dev                     # emergency: delete all RGs tagged for the env
#
# Auth: `az login` locally, or GitHub OIDC in CI. Environment stacks read the
# subscription from ARM_SUBSCRIPTION_ID (no secrets committed).

SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

BOOTSTRAP_DIR := bootstrap
ENV ?= dev
ENV_DIR := environments/$(ENV)
TFVARS := $(ENV).tfvars

# Backend coordinates: taken from CI env vars if present, else from bootstrap outputs.
define BACKEND_CONFIG
sa=$${TFSTATE_SA:-$$(terraform -chdir=$(BOOTSTRAP_DIR) output -raw state_storage_account_name)}; \
rg=$${TFSTATE_RG:-$$(terraform -chdir=$(BOOTSTRAP_DIR) output -raw shared_resource_group)}; \
container=$${TFSTATE_CONTAINER:-$$(terraform -chdir=$(BOOTSTRAP_DIR) output -raw state_container_name)}
endef

.PHONY: help fmt validate bootstrap-init bootstrap-plan bootstrap-up bootstrap-output \
        env-init env-plan env-up env-down env-nuke env-output

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

fmt: ## Format all Terraform recursively
	terraform fmt -recursive

validate: ## Validate bootstrap + every environment stack (no backend)
	terraform -chdir=$(BOOTSTRAP_DIR) init -backend=false -input=false >/dev/null && \
	  terraform -chdir=$(BOOTSTRAP_DIR) validate
	@for d in environments/*/; do \
	  echo "== validate $$d =="; \
	  terraform -chdir=$$d init -backend=false -input=false >/dev/null && \
	    terraform -chdir=$$d validate || exit 1; \
	done

# ── Bootstrap (shared, persistent — run once by an admin) ─────────────────────
bootstrap-init: ## terraform init for the bootstrap stack (local state)
	terraform -chdir=$(BOOTSTRAP_DIR) init -input=false

bootstrap-plan: bootstrap-init ## Plan the shared foundation
	terraform -chdir=$(BOOTSTRAP_DIR) plan -input=false

bootstrap-up: bootstrap-init ## Apply the shared foundation
	terraform -chdir=$(BOOTSTRAP_DIR) apply -input=false

bootstrap-output: ## Show bootstrap outputs (state SA, ACR, identity client IDs)
	terraform -chdir=$(BOOTSTRAP_DIR) output

# ── Environments (ephemeral) ──────────────────────────────────────────────────
env-init: ## Init an env stack against the remote backend  (ENV=dev|staging|prod)
	@$(BACKEND_CONFIG); \
	terraform -chdir=$(ENV_DIR) init -input=false -reconfigure \
	  -backend-config="resource_group_name=$$rg" \
	  -backend-config="storage_account_name=$$sa" \
	  -backend-config="container_name=$$container" \
	  -backend-config="key=$(ENV).tfstate" \
	  -backend-config="use_azuread_auth=true"

env-plan: env-init ## Plan an environment
	terraform -chdir=$(ENV_DIR) plan -input=false -var-file=$(TFVARS)

env-up: env-init ## Create/refresh an environment
	terraform -chdir=$(ENV_DIR) apply -input=false -auto-approve -var-file=$(TFVARS)

env-down: env-init ## Destroy an environment (clean teardown)
	terraform -chdir=$(ENV_DIR) destroy -input=false -auto-approve -var-file=$(TFVARS)

env-output: env-init ## Show environment outputs
	terraform -chdir=$(ENV_DIR) output

env-nuke: ## Emergency: delete every resource group tagged for the env
	@echo "Deleting all resource groups tagged project=rtsa environment=$(ENV) ..."; \
	az group list --query "[?tags.project=='rtsa' && tags.environment=='$(ENV)'].name" -o tsv | \
	  xargs -r -n1 -I{} az group delete --name {} --yes --no-wait
