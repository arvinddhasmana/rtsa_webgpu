<!-- CLASSIFICATION: UNCLASSIFIED -->

# 09 — Operator Runbook (Setup, Manual Validation, and First Pipeline Run)

> **Parent**: [README](README.md) · **Prev**: [08 Cost & Teardown](08-cost-model-and-teardown.md)
> **Classification**: UNCLASSIFIED · **Status**: IMPLEMENTATION RUNBOOK

---

## 1. What This Runbook Covers

This runbook answers the practical execution questions for the current implementation state:

1. Where and how to create GitHub environments and variables.
2. How to run `infra-up.yml` for `dev` and map Terraform outputs to environment variables.
3. Whether you can test manually before CI/CD deployment to `main`.
4. What happens when you push to `main`.

## 2. Important Current-State Constraint

At this phase, only the `dev` Terraform environment exists under `infra/terraform/environments/dev`.

- `infra-up` and `infra-down` are productionized workflow files.
- Running them for `test`, `staging`, or `prod` will fail until those Terraform directories are added.

For the planned multi-subscription operating model and phased rollout across
dev/test/staging/prod, see [11 — Multi-Subscription Analysis](11-multi-subscription-analysis.md)
and [12 — Multi-Subscription Migration Plan](12-multi-subscription-migration-plan.md).

## 3. Prerequisites

- You have repo admin or maintainer access for:
  - Settings -> Environments
  - Actions workflows
- Azure OIDC identity/federation from bootstrap already exists.
- Shared bootstrap resources already exist (state storage, ACR, etc.).

### 3.1 If prerequisites are missing

If OIDC identities/federation and shared bootstrap resources are not present yet,
create them first.

1. Prepare bootstrap variables:

```bash
cp infra/terraform/bootstrap/terraform.tfvars.example infra/terraform/bootstrap/terraform.tfvars
# edit infra/terraform/bootstrap/terraform.tfvars
```

2. Run bootstrap:

```bash
scripts/azure/bootstrap.sh
```

3. Verify bootstrap outputs exist:

```bash
terraform -chdir=infra/terraform/bootstrap output
```

Expected outputs include shared state/registry and identity data such as:

- `shared_resource_group`
- `state_storage_account_name`
- `state_container_name`
- `acr_name`
- `deployer_identity_client_ids`
- `tenant_id`

If bootstrap fails, fix that before any GitHub Environment setup.

## 4. Create GitHub Environments and Variables (UI)

### 4.0 Automated option (recommended)

Prefer automation over manual setup:

```bash
scripts/azure/setup-github-environments.sh --repo <owner>/<repo>
```

What this script does:

- Creates `dev,test,staging,prod` GitHub Environments.
- Seeds shared build variables from bootstrap outputs:
  - `AZURE_TENANT_ID`
  - `REPO_AZURE_SUBSCRIPTION_ID`
  - `REPO_AZURE_CLIENT_ID`
  - `ACR_NAME`
  - `TFSTATE_RESOURCE_GROUP`
  - `TFSTATE_STORAGE_ACCOUNT`
  - `TFSTATE_CONTAINER`
- Seeds environment-level `AZURE_CLIENT_ID` values from bootstrap
  `deployer_identity_client_ids`.
- Seeds environment-level `AZURE_SUBSCRIPTION_ID` values for `dev`, `test`,
  `staging`, and `prod`.
- Creates placeholders (`__SET_ME__`) for environment values that are not known
  until `infra-up` runs.

You can still use the UI flow below if preferred.

### 4.1 Create environments

In GitHub:

1. Open repository.
2. Go to Settings -> Environments.
3. Create: `dev`, `test`, `staging`, `prod`.

Recommended protections:

- `dev`: no manual approval requirement.
- `test`: optional reviewer requirement.
- `staging`: required reviewers + optional wait timer.
- `prod`: required reviewers + branch restrictions + optional wait timer.

### 4.2 Set required variables

Set environment-scoped values in Settings -> Environments -> <env> -> Variables,
and shared build values in Settings -> Secrets and variables -> Actions -> Variables.

| Variable                          | Description                                                     |
| --------------------------------- | --------------------------------------------------------------- |
| `AZURE_CLIENT_ID`                 | OIDC federated identity client id for that environment          |
| `AZURE_TENANT_ID`                 | Entra tenant id                                                 |
| `AZURE_SUBSCRIPTION_ID`           | Environment subscription id for infra and deploy jobs           |
| `REPO_AZURE_CLIENT_ID`            | Shared build identity client id for CD Build                    |
| `REPO_AZURE_SUBSCRIPTION_ID`      | Shared build subscription id for CD Build                       |
| `ACR_NAME`                        | ACR name without `.azurecr.io`                                  |
| `AKS_CLUSTER_NAME`                | AKS cluster name for the environment                            |
| `AKS_RESOURCE_GROUP`              | AKS resource group for the environment                          |
| `KEY_VAULT_NAME`                  | Key Vault name used by workload secrets                         |
| `WEBTRANSPORT_IDENTITY_CLIENT_ID` | Workload identity for `svc-webtransport` service account        |
| `ISTIO_REVISION`                  | Istio revision label (empty if default injection label is used) |
| `TFSTATE_RESOURCE_GROUP`          | Shared Terraform state RG                                       |
| `TFSTATE_STORAGE_ACCOUNT`         | Shared Terraform state storage account                          |
| `TFSTATE_CONTAINER`               | Shared Terraform state container                                |
| `VITE_API_GATEWAY_URL`            | Frontend cold-path API URL                                      |
| `VITE_WEBTRANSPORT_URL`           | Frontend WebTransport URL                                       |

Notes:

- `ACR_NAME` and `TFSTATE_*` are shared values from bootstrap and can be repo-level variables instead of per-environment variables.
- `AZURE_SUBSCRIPTION_ID` is required per environment.
- Keep value naming consistent across environments to avoid workflow drift.

Access note:

- The environment deployer identity needs `Contributor` + `User Access Administrator`
  in its environment subscription.
- It also needs `User Access Administrator` on the shared foundation resource group
  because `infra-up` creates role assignments on shared ACR and related resources.
- The shared ACR and Log Analytics workspace themselves remain in the shared
  subscription and are read through the shared provider alias.

## 5. Run infra-up for dev and Wire Outputs

### 5.1 Run workflow

In GitHub:

1. Actions tab.
2. Open workflow: `Infra Up`.
3. Click Run workflow.
4. Branch: `main`.
5. Environment input: `dev`.
6. Run.

### 5.2 Collect outputs

When the run completes, open the job summary for step `Publish outputs to job summary`.

Current Terraform output names are:

- `resource_group`
- `cluster_name`
- `node_resource_group`
- `oidc_issuer_url`
- `key_vault_uri`
- `acr_login_server`
- `storage_account_name`
- `workload_identity_client_ids`

### 5.3 Map outputs to environment variables

Automated sync (recommended):

```bash
scripts/azure/sync-dev-github-vars-from-tf.sh --repo <owner>/<repo> --environment dev
```

This script reads Terraform outputs and updates environment variables automatically.

Verify the live Azure and AKS platform immediately after `infra-up`:

```bash
export ARM_SUBSCRIPTION_ID=<environment-subscription-id>
scripts/azure/verify-infrastructure-deployment.sh --env dev
```

The same verification is a blocking post-apply step in `infra-up.yml`. A passing result
proves that Terraform has no post-apply drift, the resources are in the expected
subscription, AKS and its node pools are healthy, and Kubernetes system pods are Ready.
GitHub deployer identities receive `Azure Kubernetes Service RBAC Cluster Admin` through
bootstrap. A local operator needs the same role at the target AKS cluster scope; use the
role-assignment example in the Getting Started guide before running Kubernetes checks.

Manual mapping (if not using script):

Use this mapping for `dev` environment variables:

- `AKS_RESOURCE_GROUP` <- `resource_group`
- `AKS_CLUSTER_NAME` <- `cluster_name`
- `WEBTRANSPORT_IDENTITY_CLIENT_ID` <- `workload_identity_client_ids.webtransport` (key name depends on identity module map key)
- `KEY_VAULT_NAME` <- derived from `key_vault_uri`
  - Example: `https://kv-rtsa-dev-cc.vault.azure.net/` -> `kv-rtsa-dev-cc`
- `ACR_NAME` <- derived from `acr_login_server`
  - Example: `acrrtsashared.azurecr.io` -> `acrrtsashared`

Get `ISTIO_REVISION` if needed:

```bash
az aks show -g <AKS_RESOURCE_GROUP> -n <AKS_CLUSTER_NAME> --query "serviceMeshProfile.istio.revisions" -o tsv
```

If this returns empty, keep `ISTIO_REVISION` empty and rely on default injection label behavior.

Optional: set frontend URLs during sync:

```bash
scripts/azure/sync-dev-github-vars-from-tf.sh \
  --repo <owner>/<repo> \
  --environment dev \
  --vite-api-url https://api.example.mil \
  --vite-webtransport-url https://wt.example.mil:4443
```

## 6. Can You Run and Test Manually Before CI/CD?

Yes. Use this sequence:

### 6.1 Local Terraform validation (no cloud changes)

```bash
export PATH="$HOME/.local/bin:$PATH"
cd infra/terraform/environments/dev
terraform init
terraform validate
terraform plan -var-file=dev.tfvars
```

### 6.2 Local Helm/chart validation (no cluster changes)

```bash
cd /home/arvind/workspace/rtsa_webgpu
helm lint deploy/charts/rtsa-service
helm lint deploy/charts/redpanda-dev
helm lint deploy/charts/clickhouse-dev
helm lint deploy/charts/rtsa-mesh
```

Optional render checks:

```bash
helm template svc-track deploy/charts/rtsa-service -f deploy/charts/values/svc-track.yaml --set image.repository=example.azurecr.io/svc-track --set image.tag=dryrun --namespace rtsa > /tmp/svc-track.yaml
```

### 6.3 Manual workflow tests before main push

You can run these from Actions using `workflow_dispatch`:

- `Infra Up` for `dev`.
- `Infra Down` for `dev` (cleanup test).
- `CI` for validation gates.
- `CD Build` manual run for image build/sign/scan.

Recommended dry run sequence:

1. `scripts/azure/setup-github-environments.sh --repo <owner>/<repo>`
2. Run `Infra Up` for `dev`.
3. Confirm the workflow's `Verify infrastructure deployment` step passes.
4. Run `scripts/azure/sync-dev-github-vars-from-tf.sh --repo <owner>/<repo> --environment dev`.
5. Run `CI` manually.
6. Run `CD Build` manually from branch (build/sign/scan only).
7. Merge/push to `main` for auto dev deploy and confirm `Verify workload deployment` passes.

Important:

- A manual `CD Build` run from a non-main branch validates build/sign/scan but does not auto-deploy to dev.

### 6.4 Optional manual AKS deployment test (without pushing main)

Use the same values files and command shape implemented in workflows:

```bash
helm upgrade --install svc-track deploy/charts/rtsa-service \
  -f deploy/charts/values/svc-track.yaml \
  --namespace rtsa \
  --set image.repository=<ACR_NAME>.azurecr.io/svc-track \
  --set image.tag=<EXISTING_IMAGE_TAG>
```

After the complete release set is installed, run the same post-Helm gate used by CD:

```bash
scripts/azure/verify-workload-deployment.sh \
  --namespace rtsa \
  --expected-image-tag <EXISTING_IMAGE_TAG>
```

## 7. First main Push Behavior

After configuration is complete:

1. Push changes to `main`.
2. `cd-build.yml` triggers.
3. Changed services are built and pushed to ACR.
4. Trivy scan and cosign signing run.
5. `deploy-dev` runs automatically on `main` and deploys:
   - `rtsa-mesh`
   - `redpanda-dev`
   - `clickhouse-dev`
   - RTSA services and frontend values in `deploy/charts/values/`

## 8. Troubleshooting Quick Guide

- `terraform init` backend errors:
  - Check `TFSTATE_RESOURCE_GROUP`, `TFSTATE_STORAGE_ACCOUNT`, `TFSTATE_CONTAINER`.
- Azure login failures in Actions:
  - Check `AZURE_CLIENT_ID`, tenant/subscription vars, and federated credential subject format.
- `AuthorizationFailed` on `Microsoft.Authorization/roleAssignments/write` for shared ACR:
  - Check that the deployer identity has `User Access Administrator` on the shared
    foundation resource group, not just the environment subscription.
- `infra-up` succeeds but deploy fails:
  - Run `scripts/azure/verify-infrastructure-deployment.sh --env <environment>`.
- `svc-webtransport` deployment issues:
  - Run `scripts/azure/verify-workload-deployment.sh --namespace rtsa --expected-image-tag <tag>`.
  - Verify `WEBTRANSPORT_IDENTITY_CLIENT_ID`, `KEY_VAULT_NAME`, and Key Vault object names.
- No `deploy-dev` execution in CD Build:
  - Confirm workflow run was on `main`.

## 9. Suggested Operator Sequence (Current Phase)

1. Configure `dev` environment variables first.
2. Run `Infra Up` for `dev`.
3. Map outputs to variables and re-run `Infra Up` if needed.
4. Run `CI` manually to verify gates.
5. Push a small controlled change to `main` and observe `CD Build` + dev auto-deploy.
6. Confirm both deployment verification gates pass, then run `Infra Down` to validate teardown.

This sequence gives full confidence before broader environment rollout.
