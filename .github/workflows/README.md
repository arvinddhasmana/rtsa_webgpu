<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA CI/CD Workflows

GitHub Actions pipelines for the RTSA Azure deployment. All cloud access uses
**OIDC federation** (no long-lived secrets). See
[docs/deployment/azure/05-devops-cicd-and-environments.md](../../docs/deployment/azure/05-devops-cicd-and-environments.md).

## Workflows

| File                          | Trigger                       | Purpose                                                             |
| ----------------------------- | ----------------------------- | ------------------------------------------------------------------- |
| `ci.yml`                      | PR to `main`, manual          | SG-1..SG-5 security gates (secrets, headers, build, test, security) |
| `cd-build.yml`                | push `main`, tag `v*`, manual | Build/sign/scan changed images → ACR, then auto-deploy `dev`        |
| `cd-deploy.yml`               | manual (`environment`, `tag`) | Promote a built image tag to `test`/`staging`/`prod`                |
| `infra-up.yml`                | manual (`environment`)        | `terraform apply` an environment landing zone (OIDC)                |
| `infra-down.yml`              | manual + nightly `cron`       | `terraform destroy` (nightly tears down `dev`+`test`)               |
| `_reusable/go-build-test.yml` | `workflow_call`               | Build all Go modules + race tests + coverage gate                   |
| `_reusable/container.yml`     | `workflow_call`               | Buildx + SBOM + cosign keyless sign + Trivy + push                  |
| `_reusable/deploy-helm.yml`   | `workflow_call`               | Helm deploy mesh + Redpanda + ClickHouse + services                 |

## Required configuration

Create four **GitHub Environments** — `dev`, `test`, `staging`, `prod` — and add
protection rules (required reviewers, wait timers) to `staging` and `prod`.

Set these as **Environment variables** (`Settings → Environments → <env> → Variables`).
None are secrets — OIDC + Key Vault carry all sensitive material.

| Variable                          | Example                        | Source                                  |
| --------------------------------- | ------------------------------ | --------------------------------------- |
| `AZURE_CLIENT_ID`                 | UAMI client id                 | bootstrap output (per-env federated id) |
| `AZURE_TENANT_ID`                 | Entra tenant id                | `az account show`                       |
| `AZURE_SUBSCRIPTION_ID`           | subscription id                | `az account show`                       |
| `ACR_NAME`                        | `acrrtsashared`                | bootstrap output (shared)               |
| `AKS_CLUSTER_NAME`                | `aks-rtsa-dev-cc`              | `infra-up` Terraform output             |
| `AKS_RESOURCE_GROUP`              | `rg-rtsa-dev-cc`               | `infra-up` Terraform output             |
| `KEY_VAULT_NAME`                  | `kv-rtsa-dev-cc`               | `infra-up` Terraform output             |
| `WEBTRANSPORT_IDENTITY_CLIENT_ID` | workload-identity client id    | `infra-up` Terraform output             |
| `ISTIO_REVISION`                  | `asm-1-22` (empty = default)   | `az aks show ... serviceMeshProfile`    |
| `TFSTATE_RESOURCE_GROUP`          | `rg-rtsa-tfstate`              | bootstrap output (shared)               |
| `TFSTATE_STORAGE_ACCOUNT`         | `strtsatfstate`                | bootstrap output (shared)               |
| `TFSTATE_CONTAINER`               | `tfstate`                      | bootstrap output (shared)               |
| `VITE_API_GATEWAY_URL`            | `https://api.rtsa.example`     | environment ingress hostname            |
| `VITE_WEBTRANSPORT_URL`           | `https://wt.rtsa.example:4443` | environment ingress hostname            |

> `ACR_NAME` and the `TFSTATE_*` values come from the shared bootstrap and are
> identical across environments — you may also set them once as **repository**
> variables instead of per-environment.

## Federated credential subjects

The bootstrap (`infra/terraform/bootstrap`) creates one federated credential per
environment. Each must match the GitHub OIDC subject exactly:

```
repo:<org>/<repo>:environment:dev
repo:<org>/<repo>:environment:test
repo:<org>/<repo>:environment:staging
repo:<org>/<repo>:environment:prod
repo:<org>/<repo>:pull_request        # for infra plan on PRs (optional)
```

## Notes

- Only the `dev` environment stack exists under `infra/terraform/environments/`
  today (P1/P2). `test`, `staging`, and `prod` stacks are added in a later phase;
  `infra-up`/`infra-down` for those will run once their directories exist.
- The coverage gate (`min-coverage`, default `80`) reflects the SDLC policy of
  80%+ line coverage. Tune per repository policy if required.
