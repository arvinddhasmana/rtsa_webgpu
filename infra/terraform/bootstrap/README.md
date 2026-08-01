<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Bootstrap (Shared Foundation + GitHub OIDC)

This stack provisions the **one-time, persistent foundation** that every RTSA
environment depends on, plus the **GitHub Actions OIDC** identities used by CI/CD.
It is **P0** of the [Azure deployment roadmap](../../../docs/deployment/azure/07-implementation-roadmap.md).

> Run once, by an admin, with **Owner** (or **Contributor + User Access Administrator**)
> on the target subscription. Everything else (environments) is created by CI/CD using
> the OIDC identities created here — no cloud secrets are ever stored in GitHub.

## What it creates

| Resource                    | Name pattern                                | Purpose                                           |
| --------------------------- | ------------------------------------------- | ------------------------------------------------- |
| Resource group              | `rg-rtsa-shared-<loc>`                      | Holds all shared, persistent resources            |
| Storage account + container | `strtsatf<rand>` / `tfstate`                | Terraform **remote state** (blob-lease locking)   |
| Container registry          | `acrrtsa<rand>`                             | Service + web images (Entra auth, admin disabled) |
| Log Analytics workspace     | `log-rtsa-shared-<loc>`                     | Central logs/metrics sink                         |
| User-assigned identities    | `id-rtsa-<env>-deployer`, `id-rtsa-ci-plan` | GitHub OIDC federation (per-env deploy + PR plan) |
| DNS zone (optional)         | `var.dns_zone_name`                         | Ingress hostnames                                 |

The **event backbone (Redpanda OSS)** and **ClickHouse** are deployed later as in-cluster
Helm workloads (P2) — not here.

## Identity & least privilege

- **Per-environment deployer** (`id-rtsa-dev/staging/prod-deployer`): federated to
  `repo:<owner>/<repo>:environment:<env>` so only jobs targeting the matching **GitHub
  Environment** (with its protection rules) can assume it. Roles: `Contributor` +
  `User Access Administrator` in the environment subscription, `User Access Administrator`
  on the shared foundation resource group, `AcrPush` (ACR), and `Storage Blob Data
Contributor` (state).
- **CI plan identity** (`id-rtsa-ci-plan`): federated to `repo:<owner>/<repo>:pull_request`,
  read-only (`Reader` + `Storage Blob Data Reader`) — runs `terraform plan` on PRs.

> `User Access Administrator` is required so the P1 landing zone can create role
> assignments (kubelet→ACR, workloads→Key Vault). For enterprise, replace it with a
> **constrained "Role Based Access Control Administrator"** (condition-scoped to the
> roles you actually delegate) or scope it to management groups / specific RGs.

## Run it

```bash
cd infra/terraform/bootstrap
cp terraform.tfvars.example terraform.tfvars    # edit subscription_id, github_owner, github_repo
az login                                         # admin identity

# from infra/terraform/
make bootstrap-up            # or: ../../scripts/azure/bootstrap.sh
make bootstrap-output        # copy IDs into GitHub (or use --set-github)
```

Wire GitHub automatically (needs `gh`):

```bash
scripts/azure/bootstrap.sh --set-github
```

That delegates to [scripts/azure/setup-github-environments.sh](../../../scripts/azure/setup-github-environments.sh),
which creates the `dev`, `test`, `staging`, and `prod` GitHub Environments and seeds
the current shared build variables plus the per-environment subscription and identity
variables.

If you need the legacy repository-level `AZURE_SUBSCRIPTION_ID` for transitional
jobs, pass `--set-legacy-repo-subscription-id` to the setup script directly.

| Scope                                   | Name                         | Source output                                      |
| --------------------------------------- | ---------------------------- | -------------------------------------------------- |
| repo variable                           | `AZURE_TENANT_ID`            | `tenant_id`                                        |
| repo variable                           | `TFSTATE_SUBSCRIPTION_ID`    | `subscription_id`                                  |
| repo variable                           | `REPO_AZURE_SUBSCRIPTION_ID` | bootstrap `subscription_id`                        |
| repo variable                           | `REPO_AZURE_CLIENT_ID`       | shared deployer identity (default dev)             |
| repo variable                           | `ACR_NAME`                   | `acr_name`                                         |
| repo variable                           | `TFSTATE_RESOURCE_GROUP`     | `shared_resource_group`                            |
| repo variable                           | `TFSTATE_STORAGE_ACCOUNT`    | `state_storage_account_name`                       |
| repo variable                           | `TFSTATE_CONTAINER`          | `state_container_name`                             |
| **environment** `dev/test/staging/prod` | `AZURE_CLIENT_ID`            | `deployer_identity_client_ids[env]`                |
| **environment** `dev/test/staging/prod` | `AZURE_SUBSCRIPTION_ID`      | `--environment-subscriptions` or bootstrap default |

## State model

Bootstrap uses **local state** by design — it _creates_ the remote backend. Keep
`terraform.tfstate` for the bootstrap stack out of git (already gitignored). The
shared RG and state storage account carry `prevent_destroy = true`.

### Migrating bootstrap state to remote (optional)

After the first apply you may move bootstrap state into the container it created:

```bash
# add a backend block, then:
terraform init -migrate-state \
  -backend-config="resource_group_name=$(terraform output -raw shared_resource_group)" \
  -backend-config="storage_account_name=$(terraform output -raw state_storage_account_name)" \
  -backend-config="container_name=$(terraform output -raw state_container_name)" \
  -backend-config="key=bootstrap.tfstate" \
  -backend-config="use_azuread_auth=true"
```

## Hardening (deferred to P6)

- Disable storage account keys (`shared_access_key_enabled = false`) once CI uses AAD.
- Private endpoints for ACR / Storage / Key Vault; restrict `public_network_access`.
- Replace `User Access Administrator` with a condition-scoped RBAC Administrator role.
