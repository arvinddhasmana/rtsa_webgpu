<!-- CLASSIFICATION: UNCLASSIFIED -->

# 10 — Getting Started: Zero to Production (New Team Member Guide)

> **Parent**: [README](README.md) · **Prev**: [09 Operator Runbook](09-operator-runbook-manual-validation.md)
> **Classification**: UNCLASSIFIED · **Audience**: Developers and operators new to the project
> **Scope**: Everything from an empty laptop to a live production deployment

---

## How to Read This Guide

This guide is sequential. Every section ends with a **System State Checkpoint** diagram
so you can see exactly what has been created, what still doesn't exist, and what comes next.
Skip nothing on a first run — later sections depend on earlier steps having succeeded.

Scope note: this guide walks the baseline implementation path. For environment-isolated
subscriptions (dev/test/staging/prod separated), use this guide together with
[11 — Multi-Subscription Analysis](11-multi-subscription-analysis.md) and
[12 — Multi-Subscription Migration Plan](12-multi-subscription-migration-plan.md).

```mermaid
flowchart LR
    S0["0 · Mental\nModel"] --> S1["1 · Prerequisites\n& Tools"]
    S1 --> S2["2 · Bootstrap\n(shared foundation)"]
    S2 --> S3["3 · GitHub\nEnvironments"]
    S3 --> S4["4 · First Infra\nDeploy (infra-up)"]
    S4 --> S5["5 · First PR\n(CI gates)"]
    S5 --> S6["6 · Merge → main\n(CD Build)"]
    S6 --> S7["7 · CD Deploy\n(dev)"]
    S7 --> S8["8 · Promote\ntest → staging"]
    S8 --> S9["9 · Production\nDeployment"]
    S9 --> S10["10 · Ongoing Ops\n& Tear-down"]

    style S0 fill:#f0f0f0
    style S9 fill:#d4edda
```

---

## Section 0 — Mental Model: How Everything Fits Together

Before touching a keyboard, understand the three-layer model that everything in RTSA follows.

### 0.1 The Three Layers

```mermaid
flowchart TB
    subgraph L1["Layer 1 — Shared Foundation (runs once, lives forever)"]
        direction LR
        ACR["Azure Container\nRegistry (ACR)\nacrrtsabzkem9"]
        TFSA["Terraform State\nStorage\nstrtsatfbzkem9"]
        LAW["Log Analytics\nWorkspace\nlog-rtsa-shared-cc"]
        MI["Managed Identities\n+ OIDC Federated\nCredentials"]
    end

    subgraph L2["Layer 2 — Environment Platform (ephemeral, created/destroyed per need)"]
        direction LR
        AKS["AKS Cluster\naks-rtsa-dev"]
        KV["Key Vault\nkv-rtsa-dev-rg7auf"]
        VNET["VNet + Subnets\n10.0.0.0/16"]
        WKID["Workload\nIdentities (3)"]
    end

    subgraph L3["Layer 3 — Application Workloads (deployed via Helm/GitOps)"]
        direction LR
        RP["Redpanda\n(event backbone)"]
        CH["ClickHouse\n(OLAP store)"]
        SVCS["16 Go Services\n+ WebGPU frontend"]
        MESH["Istio mesh\nmTLS · rate-limit"]
    end

    L1 -->|"provides pull access\nand state storage"| L2
    L2 -->|"provides the K8s\nplatform to run on"| L3
```

**Key insight**: Destroying an environment (Layer 2) never touches images in ACR or Terraform
state files in the Storage Account. You can rebuild from scratch in ~20 minutes at any time.

### 0.2 The Full Deployment Journey

```mermaid
flowchart LR
    subgraph DEV["Developer Machine"]
        CODE["Code change\n(Go/TS/WGSL)"]
        GITPR["git push\n→ Pull Request"]
    end

    subgraph GH["GitHub"]
        subgraph CI["CI Pipeline (on PR)"]
            SG1["SG-1\nPre-build\nsecrets/headers/fmt"]
            SG2["SG-2\nBuild\ngo build · buf gen"]
            SG3["SG-3\nTests\nrace · 80% coverage"]
            SG4["SG-4\nSecurity\ngosec · trivy · SARIF"]
            SG5["SG-5\nIntegration\ncompose e2e smoke"]
            SG1-->SG2-->SG3-->SG4-->SG5
        end
        MERGE["Merge to main"]
        subgraph CD["CD Pipeline (on merge)"]
            BUILD["CD Build\nDocker → sign → ACR"]
            DEPLOY["CD Deploy\nHelm upgrade → AKS"]
        end
        PROMOTE["Promote digest\ntest → staging → prod"]
    end

    subgraph AZ["Azure"]
        ACRREG["ACR\n(image store)"]
        DEVENV["Dev AKS\n(auto-deploy)"]
        STAGENV["Staging AKS\n(manual gate)"]
        PRODENV["Prod AKS\n(approval required)"]
    end

    CODE --> GITPR --> CI
    SG5 --> MERGE --> BUILD
    BUILD -->|"push digest"| ACRREG
    ACRREG --> DEPLOY --> DEVENV
    DEVENV -->|"green + tag"| STAGENV
    STAGENV -->|"human approval"| PRODENV
```

### 0.3 Terraform vs Helm — What Goes Where

This distinction matters. Getting it wrong causes frustration.

| Managed by    | Tool                         | What it creates                                                                           | When                |
| ------------- | ---------------------------- | ----------------------------------------------------------------------------------------- | ------------------- |
| **Terraform** | `infra-up.yml` workflow      | Azure resources: VNet, AKS, Key Vault, Managed Identities, role assignments               | Before any pods run |
| **Helm**      | `cd-deploy.yml` workflow     | Kubernetes resources: Deployments, Services, ConfigMaps, KEDA ScaledObjects, Istio config | After AKS exists    |
| **Bootstrap** | `scripts/azure/bootstrap.sh` | Shared resources: ACR, state storage, OIDC app registrations                              | Once, by an admin   |

Terraform does not know about your Go code. Helm does not know about VNets. They each
manage their own state independently.

### 0.4 Identity and Trust: No Passwords Anywhere

```mermaid
flowchart TB
    subgraph GH["GitHub Actions"]
        RUNNER["Job Runner\n(ubuntu-latest)"]
        TOKEN["GitHub OIDC\nJWT Token\n(short-lived, auto-issued)"]
    end

    subgraph ENTRA["Microsoft Entra ID"]
        APP["User-Assigned\nManaged Identity\nid-rtsa-dev-deployer"]
        FED["Federated\nCredential\n(trusts github.com OIDC)"]
    end

    subgraph AZ["Azure Resources"]
        TFSTATE["Storage Account\n(Contributor)"]
        AKSRG["Resource Group\n(Contributor)"]
        ACRRG["ACR\n(AcrPush)"]
    end

    RUNNER -->|"request OIDC token"| TOKEN
    TOKEN -->|"exchange via WorkloadIdentity\nno password needed"| APP
    FED -->|"trust anchors"| APP
    APP -->|"RBAC role assignments"| TFSTATE
    APP -->|"RBAC role assignments"| AKSRG
    APP -->|"RBAC role assignments"| ACRRG
```

**This is OIDC federation.** GitHub proves its identity to Azure by signing a JWT. No Azure
client secrets are ever stored in GitHub. This is set up by the bootstrap step and never
needs changing unless you rotate the federation.

---

## Section 1 — Prerequisites: What You Need Before You Start

Run through this checklist completely. Every item marked _(one-time)_ only needs doing once
per developer machine.

### 1.1 Required Tools

| Tool              | Minimum version    | Install command                                                                                  | Check                      |
| ----------------- | ------------------ | ------------------------------------------------------------------------------------------------ | -------------------------- |
| `az` (Azure CLI)  | 2.60+              | `brew install azure-cli` or [docs.microsoft.com/cli/azure](https://docs.microsoft.com/cli/azure) | `az --version`             |
| `terraform`       | 1.10.5             | `brew install terraform` or [releases.hashicorp.com](https://releases.hashicorp.com/terraform/)  | `terraform version`        |
| `gh` (GitHub CLI) | 2.40+              | `brew install gh` or [cli.github.com](https://cli.github.com)                                    | `gh --version`             |
| `kubectl`         | 1.30+              | `az aks install-cli`                                                                             | `kubectl version --client` |
| `helm`            | 3.15+              | `brew install helm`                                                                              | `helm version`             |
| `go`              | 1.24+              | [go.dev/dl](https://go.dev/dl)                                                                   | `go version`               |
| `node` / `pnpm`   | Node 22+ / pnpm 9+ | `brew install node pnpm`                                                                         | `pnpm --version`           |
| `buf`             | 1.40+              | `brew install bufbuild/buf/buf`                                                                  | `buf --version`            |
| `docker`          | 25+                | [docs.docker.com/get-docker](https://docs.docker.com/get-docker/)                                | `docker --version`         |
| `cosign`          | 2.4+               | `brew install cosign`                                                                            | `cosign version`           |

### 1.2 Accounts and Permissions

| Account                | What you need                                                          | Who grants it            |
| ---------------------- | ---------------------------------------------------------------------- | ------------------------ |
| **Azure subscription** | Contributor + User Access Administrator (or Owner) on the subscription | Azure subscription owner |
| **GitHub repository**  | Write access + Actions enabled                                         | Repo admin               |
| **GitHub token**       | `gh auth login` with `repo`, `workflow`, `write:variables` scopes      | Yourself                 |

### 1.3 Log in to Everything

```bash
# Azure — interactive browser login
az login

# Confirm you are on the right subscription
az account show --query '{name:name, id:id}' -o json

# If wrong subscription, switch:
az account set --subscription "11f614f9-a6d3-419b-9437-37a84c75f27a"

# GitHub CLI — interactive login, select HTTPS, copy+paste the one-time code
gh auth login

# Verify
gh auth status
```

### 1.4 Clone the Repository

```bash
git clone https://github.com/arvinddhasmana/rtsa_webgpu.git
cd rtsa_webgpu

# Verify you have all workspace modules
go work sync
```

---

## Section 2 — Step 1: Bootstrap (One-Time, Admin Only)

The bootstrap creates the _shared foundation_ that every environment depends on.
Run this exactly once. If it already exists (ask the team), skip to Section 3.

### 2.1 What Bootstrap Provisions

```mermaid
flowchart TB
    subgraph BEFORE["Before Bootstrap: Empty Azure subscription"]
        EMPTY["Nothing RTSA-related exists"]
    end

    subgraph BOOTSTRAP["bootstrap.sh runs terraform apply"]
        direction TB
        TF["infra/terraform/bootstrap/\nmain.tf · variables.tf · outputs.tf"]
    end

    subgraph AFTER["After Bootstrap: rg-rtsa-shared-cc"]
        direction TB
        SA["Storage Account\nstrtsatfbzkem9\n(Terraform remote state)"]
        CONT["Blob container: tfstate\n(holds .tfstate per environment)"]
        ACR_R["ACR: acrrtsabzkem9\n(all container images)"]
        LAW_R["Log Analytics: log-rtsa-shared-cc\n(AKS Container Insights)"]
        MI_DEV["MI: id-rtsa-dev-deployer\nClient ID: 698f32f3-..."]
        MI_STG["MI: id-rtsa-staging-deployer"]
        MI_PRD["MI: id-rtsa-prod-deployer"]
        MI_CI["MI: id-rtsa-ci-plan\n(CI Terraform plan)"]
        FED_DEV["Federated Cred on MI: dev\nTrusts github.com main branch"]
        FED_STG["Federated Cred on MI: staging"]
        FED_PRD["Federated Cred on MI: prod"]
    end

    BEFORE --> BOOTSTRAP --> AFTER
```

**28 resources total.** None of these are expensive. The ACR Standard tier costs ~$0.10/day.
The Storage Account is negligible. These resources survive environment teardowns permanently.

### 2.2 Configure Bootstrap Variables

```bash
cd infra/terraform/bootstrap

# terraform.tfvars is gitignored — your real values stay local
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with your real values:

```hcl
subscription_id = "11f614f9-a6d3-419b-9437-37a84c75f27a"
tenant_id       = "b13ceef6-6c21-40a4-ad1c-bbaac66921c2"
location        = "canadacentral"
project         = "rtsa"

github_owner = "arvinddhasmana"
github_repo  = "rtsa_webgpu"

environments = ["dev", "staging", "prod"]
```

> **Never commit `terraform.tfvars` to git.** It is in `.gitignore`. Only `.tfvars.example`
> files belong in version control.

### 2.3 Run Bootstrap

```bash
cd /path/to/rtsa_webgpu

# Dry run first — review what will be created (no changes)
scripts/azure/bootstrap.sh --plan-only

# Apply when satisfied
scripts/azure/bootstrap.sh

# After apply — save these outputs; you will need them in Section 3
terraform -chdir=infra/terraform/bootstrap output
```

Bootstrap grants each environment deployer identity the Azure management-plane roles and
the `Azure Kubernetes Service RBAC Cluster Admin` data-plane role required by Helm and the
post-deployment verifier. Re-run bootstrap after introducing a new environment identity or
changing these role assignments.

Expected output (values vary by run):

```
acr_name                      = "acrrtsabzkem9"
acr_login_server              = "acrrtsabzkem9.azurecr.io"
shared_resource_group         = "rg-rtsa-shared-cc"
state_storage_account_name    = "strtsatfbzkem9"
state_container_name          = "tfstate"
tenant_id                     = "b13ceef6-6c21-40a4-ad1c-bbaac66921c2"
deployer_identity_client_ids  = {
  "dev"     = "698f32f3-60d4-49a8-9317-94c34c99197b"
  "staging" = "a77c5b05-5dd0-484d-96e-..."
  "prod"    = "13795bc9-6518-4b61-abeb-..."
}
ci_plan_identity_client_id    = "a223d804-258c-4b83-8526-e808cda67eeb"
```

Save these. The setup script in Step 2 reads them automatically if you run it from the
same machine, but having them written down helps debugging.

### 2.4 System State After Bootstrap

```mermaid
flowchart TB
    subgraph AZ["Azure — rg-rtsa-shared-cc (persistent)"]
        ACR_S["ACR ✅\nacrrtsabzkem9"]
        SA_S["Storage ✅\nstrtsatfbzkem9\n  tfstate/ (empty)"]
        LAW_S["Log Analytics ✅\nlog-rtsa-shared-cc"]
        MI_S["4 Managed Identities ✅\n(deployer×3 + ci-plan)"]
        FED_S["OIDC Federated Creds ✅\n(GitHub trusted)"]
    end

    subgraph GH["GitHub"]
        ENV_S["Environments ❌ not yet\nVariables ❌ not yet"]
    end

    subgraph ENVS["Azure — Environments (all ephemeral)"]
        DEV_S["dev ❌ not provisioned"]
        STG_S["staging ❌ not provisioned"]
        PRD_S["prod ❌ not provisioned"]
    end
```

**What works right now**: An admin with the Managed Identity can authenticate to Azure via
OIDC from a GitHub Actions runner. **What doesn't work yet**: No GitHub Environment
variables exist, so workflows would fail at `az login` / `terraform init`.

---

## Section 3 — Step 2: Configure GitHub Environments

GitHub Environments are where per-environment configuration and protection rules live.
They hold the values that workflows need to deploy to a specific environment without
hardcoding anything.

### 3.1 What GitHub Environments Provide

```mermaid
flowchart LR
    subgraph GHE["GitHub Environment: dev"]
        direction TB
        PROT["Protection Rules\n(required reviewers, wait timer)"]
        VARS["Environment Variables\nAZURE_CLIENT_ID\nAKS_CLUSTER_NAME\nKEY_VAULT_NAME\netc."]
        SECS["Environment Secrets\n(none — we use OIDC, no secrets needed)"]
    end

    subgraph WF["GitHub Actions Workflow"]
        JOB["job:\n  environment: dev"]
    end

    subgraph AZ["Azure"]
        OIDC_T["OIDC Token exchange\n(using AZURE_CLIENT_ID from env)"]
        RESRC["Azure Resources\n(deployed to this env)"]
    end

    WF -->|"inherits vars/secrets"| GHE
    WF --> OIDC_T -->|"Managed Identity auth"| RESRC
```

A workflow job that declares `environment: dev` automatically:

1. Waits for any required reviewers to approve (prod only)
2. Gets all variables from that environment injected as `vars.*`
3. Uses those to authenticate to Azure and target the right cluster

### 3.2 Variables Reference: What Each Variable Does

| Variable                          | Scope   | Source                                              | Purpose                                                   |
| --------------------------------- | ------- | --------------------------------------------------- | --------------------------------------------------------- |
| `AZURE_TENANT_ID`                 | Repo    | Bootstrap output                                    | Entra tenant for OIDC exchange                            |
| `REPO_AZURE_SUBSCRIPTION_ID`      | Repo    | bootstrap `terraform.tfvars`                        | Shared build subscription used by CD Build                |
| `REPO_AZURE_CLIENT_ID`            | Repo    | setup script (default: dev deployer identity)       | Shared build identity for CD Build                        |
| `ACR_NAME`                        | Repo    | Bootstrap output `acr_name`                         | Image registry used by all environments                   |
| `TFSTATE_RESOURCE_GROUP`          | Repo    | Bootstrap output `shared_resource_group`            | Where the state storage account lives                     |
| `TFSTATE_STORAGE_ACCOUNT`         | Repo    | Bootstrap output `state_storage_account_name`       | Storage account name for Terraform backends               |
| `TFSTATE_CONTAINER`               | Repo    | Bootstrap output `state_container_name`             | Blob container holding `.tfstate` files                   |
| `AZURE_SUBSCRIPTION_ID`           | Per-env | setup script / env override                         | Target subscription for the environment                   |
| `AZURE_CLIENT_ID`                 | Per-env | Bootstrap output `deployer_identity_client_ids.dev` | **Which Managed Identity to impersonate in this env**     |
| `AKS_CLUSTER_NAME`                | Per-env | infra-up output `cluster_name`                      | AKS cluster name for `kubectl` context                    |
| `AKS_RESOURCE_GROUP`              | Per-env | infra-up output `resource_group`                    | Resource group containing AKS                             |
| `KEY_VAULT_NAME`                  | Per-env | infra-up output `key_vault_uri` (parsed)            | Key Vault for CSI secrets + Helm values                   |
| `WEBTRANSPORT_IDENTITY_CLIENT_ID` | Per-env | infra-up output                                     | Client ID for the WebTransport workload identity          |
| `ISTIO_REVISION`                  | Per-env | `dev.tfvars` `istio_revisions[0]`                   | Which Istio ASM revision is active (for injection labels) |
| `VITE_API_GATEWAY_URL`            | Per-env | Manual — set after first Helm deploy                | gRPC-Web gateway URL injected into the frontend build     |
| `VITE_WEBTRANSPORT_URL`           | Per-env | Manual — set after first Helm deploy                | WebTransport server URL for the COP hot path              |

### 3.3 Automated Setup (Recommended)

```bash
# Creates all 4 environments + seeds repo-level + env-level variables
# from live bootstrap Terraform outputs
scripts/azure/setup-github-environments.sh \
  --repo arvinddhasmana/rtsa_webgpu

# Optional multi-subscription override:
# scripts/azure/setup-github-environments.sh \
#   --repo arvinddhasmana/rtsa_webgpu \
#   --environment-subscriptions \
#   dev=<dev-sub>,test=<test-sub>,staging=<stg-sub>,prod=<prod-sub>

# After infra-up (Section 4), sync the remaining env-level variables:
scripts/azure/sync-dev-github-vars-from-tf.sh \
  --repo arvinddhasmana/rtsa_webgpu \
  --environment dev
```

These two scripts together cover all variables except `VITE_API_GATEWAY_URL` and
`VITE_WEBTRANSPORT_URL`, which can only be known after the first successful Helm deploy.

### 3.4 Manual Setup Fallback (if automated fails)

If the scripts fail, the fallback is the GitHub web UI:

1. Go to **Settings → Environments** in the repository.
2. Create four environments: `dev`, `test`, `staging`, `prod`.
3. For `staging` and `prod`, add a **required reviewer** (yourself initially).
4. Add each variable from the table above under the correct scope.

> Add repo-level variables at **Settings → Secrets and variables → Actions → Variables tab**.

### 3.5 System State After GitHub Setup

```mermaid
flowchart TB
    subgraph GH["GitHub"]
        subgraph ENVGH["Environments ✅"]
            DEVG["dev\nAZURE_CLIENT_ID=698f32f3-...\nAKS_CLUSTER_NAME=__SET_ME__"]
            TSTG["test\nAZURE_CLIENT_ID=..."]
            STGG["staging\nRequired reviewers: enabled"]
            PRDG["prod\nRequired reviewers: enabled"]
        end
        subgraph REPOV["Repo-level Variables ✅"]
            RV1["AZURE_TENANT_ID"]
            RV2["REPO_AZURE_SUBSCRIPTION_ID"]
            RV3["ACR_NAME = acrrtsabzkem9"]
            RV4["TFSTATE_STORAGE_ACCOUNT = strtsatfbzkem9"]
        end
    end

    subgraph AZ["Azure — Still Missing"]
        DEVAKS["dev AKS ❌ not provisioned yet\n(AKS_CLUSTER_NAME = __SET_ME__)"]
    end
```

The `AKS_CLUSTER_NAME` and related variables remain `__SET_ME__` until infra-up runs.
The CI pipeline will still work because it does not need the cluster to lint/test code.

---

## Section 4 — Step 3: Provision Infrastructure (infra-up)

`infra-up.yml` is the workflow that runs Terraform for a given environment.
It creates the AKS cluster, Key Vault, VNet, and workload identities.

### 4.1 What infra-up Creates (dev environment)

```mermaid
flowchart TB
    subgraph SHARED["rg-rtsa-shared-cc (already exists from bootstrap)"]
        ACR_U["ACR ✅\nacrrtsabzkem9"]
        LAW_U["Log Analytics ✅\nlog-rtsa-shared-cc"]
    end

    subgraph DEV["rg-rtsa-dev-cc (created by infra-up)"]
        subgraph VNET_U["VNet 10.0.0.0/16"]
            SNET1["snet-system\n10.0.1.0/24"]
            SNET2["snet-user\n10.0.2.0/24"]
            SNET3["snet-private-endpoints\n10.0.3.0/24"]
        end
        subgraph AKS_U["AKS: aks-rtsa-dev (Standard, Free tier)"]
            SYS_U["system pool\n1-2 × Standard_D2pds_v6\nARM v6, no AZ"]
            ING_U["ingest pool\n0-3 × Standard_D2pds_v6\nSpot, scale-to-zero"]
            PROC_U["process pool\n0-3 × Standard_D2pds_v6\nSpot, scale-to-zero"]
            STAT_U["stateful pool\n1-2 × Standard_D2pds_v6\nOn-demand, 128 GB disk"]
        end
        KV_U["Key Vault\nkv-rtsa-dev-rg7auf"]
        SA_U["Storage Account\nstrtsadevrg7auf\ncontainers: tiered, backups"]
        subgraph WKID_U["Workload Identities (3)"]
            WI1["id-rtsa-dev-webtransport\n+ federated cred (K8s SA)"]
            WI2["id-rtsa-dev-track\n+ federated cred (K8s SA)"]
            WI3["id-rtsa-dev-query\n+ federated cred (K8s SA)"]
        end
    end

    SHARED -->|"kubelet MI pull"| AKS_U
    SHARED -->|"Container Insights"| LAW_U
    AKS_U -->|"CSI driver"| KV_U
    WKID_U -->|"Key Vault Secrets User role"| KV_U
```

### 4.2 How Terraform Modules Compose

```mermaid
flowchart TB
    ROOT["environments/dev/main.tf\n(root module)"]
    ROOT --> NET["modules/network\nVNet · subnets · NSGs"]
    ROOT --> AKS_M["modules/aks\nCluster · node pools\nIstio · KEDA · Workload Identity"]
    ROOT --> ACRA["modules/acr-access\nKubelet MI → ACR pull role"]
    ROOT --> KVM["modules/keyvault\nKey Vault · RBAC · CSI"]
    ROOT --> IDM["modules/identity\nWorkload MIs · federated creds\nKV role assignments"]
    ROOT --> STOM["modules/storage\nStorage account · containers"]
    ROOT --> OBS["modules/observability\nManaged Prometheus · Grafana\n(disabled in dev)"]

    NET -->|"subnet IDs"| AKS_M
    AKS_M -->|"kubelet object ID"| ACRA
    AKS_M -->|"OIDC issuer URL"| IDM
    KVM -->|"KV ID"| IDM
```

Each module is independently versioned and reused across dev/staging/prod by passing
different `tfvars`. The modules never contain environment-specific hardcoding.

### 4.3 Run infra-up from GitHub Actions (preferred)

```
GitHub UI → Actions tab → "Infra Up" → Run workflow
  environment: dev
  → Run
```

Or via CLI:

```bash
gh workflow run infra-up.yml \
  --repo arvinddhasmana/rtsa_webgpu \
  --field environment=dev
```

Watch it run:

```bash
gh run list --workflow infra-up.yml --limit 5
gh run watch <run-id>
```

### 4.4 Run infra-up Locally (admin / debugging)

```bash
# The environment provider targets nonprod; the backend remains in shared.
export ARM_SUBSCRIPTION_ID="bb2b8549-9693-40f2-9287-3bd5afcc6633"

scripts/azure/preflight-environment-deploy.sh --env dev
make -C infra/terraform env-plan ENV=dev
make -C infra/terraform env-up ENV=dev
make -C infra/terraform env-output ENV=dev
```

> **Expected apply time**: ~15 minutes for the first full apply (AKS cluster creation
> takes 8-12 minutes, Key Vault ~3 minutes).

### 4.5 After infra-up: Sync Variables to GitHub

Once apply completes, copy the Terraform outputs into the GitHub environment variables:

```bash
scripts/azure/sync-dev-github-vars-from-tf.sh \
  --repo arvinddhasmana/rtsa_webgpu \
  --environment dev
```

Expected output:

```
Updated environment variables for arvinddhasmana/rtsa_webgpu / dev:
  AKS_RESOURCE_GROUP=rg-rtsa-dev-cc
  AKS_CLUSTER_NAME=aks-rtsa-dev
  KEY_VAULT_NAME=kv-rtsa-dev-rg7auf
  ACR_NAME=acrrtsabzkem9
  WEBTRANSPORT_IDENTITY_CLIENT_ID=641ebe6a-34b8-...
  ISTIO_REVISION=asm-1-29
```

### 4.6 Verify the Cluster

```bash
export ARM_SUBSCRIPTION_ID="bb2b8549-9693-40f2-9287-3bd5afcc6633"
scripts/azure/verify-infrastructure-deployment.sh --env dev
```

The verifier checks subscription isolation, zero Terraform drift, Azure provisioning and
power states, OIDC and Workload Identity, node-pool health, Ready nodes, and `kube-system`
pods. It uses a temporary kubeconfig and does not replace your current context. The
`infra-up.yml` workflow runs the same command after every successful apply and blocks on
failure.

Local operators also need AKS data-plane access. An Azure administrator can grant it at the
specific cluster scope:

```bash
OPERATOR_ID=$(az ad signed-in-user show --query id -o tsv)
AKS_ID=$(az aks show \
    --subscription "$ARM_SUBSCRIPTION_ID" \
    --resource-group rg-rtsa-dev-cc \
    --name aks-rtsa-dev \
    --query id -o tsv)

az role assignment create \
    --assignee-object-id "$OPERATOR_ID" \
    --assignee-principal-type User \
    --role "Azure Kubernetes Service RBAC Cluster Admin" \
    --scope "$AKS_ID"
```

### 4.7 System State After infra-up

```mermaid
flowchart TB
    subgraph AZ["Azure — All Created ✅"]
        SHARED2["rg-rtsa-shared-cc\nACR · State SA · Log Analytics · 4 MIs"]
        DEV2["rg-rtsa-dev-cc\nVNet · AKS · KV · Storage · 3 workload MIs"]
    end

    subgraph GH2["GitHub — All Variables Populated ✅"]
        DEVENV2["dev environment\nAZURE_CLIENT_ID · AKS_CLUSTER_NAME\nKEY_VAULT_NAME · ISTIO_REVISION\n...all except VITE_*"]
    end

    subgraph K8S["Kubernetes — Platform Running ✅"]
        NS["Namespaces: kube-system · istio-system\nkeda · cert-manager · external-secrets"]
        ADDONS["Istio asm-1-29 · KEDA · CSI · Azure Policy"]
        NOPODS["Application pods: ❌ none yet\n(Helm deploy not run)"]
    end

    AZ --> GH2
    AZ --> K8S
```

---

## Section 5 — Step 4: Your First Pull Request (CI Pipeline)

Every code change goes through a Pull Request. The CI pipeline enforces five sequential
security gates before any merge is allowed.

### 5.1 Create a Branch and Open a PR

```bash
git checkout -b feature/my-change
# make your changes
git add -A && git commit -m "feat: describe your change"
git push origin feature/my-change

# Open PR
gh pr create --title "feat: describe your change" --body "What this does"
```

### 5.2 The Five Security Gates (SG-1 through SG-5)

```mermaid
flowchart TB
    subgraph SG1["SG-1 · Pre-build (~2 min)"]
        direction LR
        GL["gitleaks\nsecret scan\nall files"]
        CH["classification\nheader check\nchanged files only"]
        FMT["gofmt\nbuf lint + format"]
        GL --> CH --> FMT
    end

    subgraph SG2["SG-2 · Build (~3 min)"]
        direction LR
        GB["go build ./...\nall 13 services"]
        BG["buf generate\nproto → Go + TS"]
        SB["syft SBOM\nattachment"]
        GB --> BG --> SB
    end

    subgraph SG3["SG-3 · Tests (~5 min)"]
        direction LR
        GT["-race unit tests\n80% line coverage gate"]
        VT["vitest\nTypeScript/SolidJS tests"]
        GT --> VT
    end

    subgraph SG4["SG-4 · Security (~4 min)"]
        direction LR
        GS["gosec\nGo static analysis"]
        SM["semgrep\ncross-language rules"]
        GV["govulncheck\nknown CVEs"]
        TV["trivy\nimage + repo scan\nSARIF → GitHub Security"]
        GS --> SM --> GV --> TV
    end

    subgraph SG5["SG-5 · Integration (~6 min)"]
        direction LR
        DC["docker compose up\n(Redpanda + ClickHouse\n+ OTel stub)"]
        E2E["e2e smoke tests\n(send → ingest → fuse → track)"]
        DC --> E2E
    end

    PR["Pull Request opened"] --> SG1 --> SG2 --> SG3 --> SG4 --> SG5

    SG5 -->|"all green → merge allowed"| MERGE["PR mergeable ✅"]
    SG1 -->|"fail → PR blocked"| FAIL["PR blocked ❌\nfix required"]
```

### 5.3 What Each Gate Checks and Why

| Gate                | Blocks merge on...                            | Why it matters for RTSA              |
| ------------------- | --------------------------------------------- | ------------------------------------ |
| SG-1 `gitleaks`     | Any secret pattern in any file                | ITSG-33 SA-4: no credentials in code |
| SG-1 classification | Missing `CLASSIFICATION: UNCLASSIFIED` header | DND classification policy            |
| SG-1 `gofmt`        | Unformatted Go                                | Readability and diff noise           |
| SG-2 build          | Compile errors                                | Catch obvious breaks early           |
| SG-3 coverage       | < 80% line coverage                           | SDLC mandates test completeness      |
| SG-4 `gosec`        | CWE-780, SQL injection, hardcoded creds, etc. | OWASP Top 10 / NIST 800-53           |
| SG-4 `trivy`        | CRITICAL/HIGH CVEs in dependencies            | Supply chain security                |
| SG-5 e2e            | Broken service integration                    | Catches contract regressions         |

### 5.4 Making CI Pass: Common Fixes

```bash
# SG-1: Add missing classification header to a new file
echo "// CLASSIFICATION: UNCLASSIFIED" | cat - myfile.go > tmp && mv tmp myfile.go

# SG-1: Fix gofmt
gofmt -w ./...

# SG-1: Fix buf lint
buf lint && buf format -w

# SG-3: Check your coverage
go test -coverprofile=cover.out ./... && go tool cover -func=cover.out | tail -1

# SG-4: Run gosec locally before push
gosec ./...

# SG-5: Reproduce the e2e smoke locally
docker compose -f deploy/docker-compose.yml up -d
go test -tags=integration ./tests/e2e/... -v
docker compose -f deploy/docker-compose.yml down
```

---

## Section 6 — Step 5: Merge to main (CD Build)

When the PR merges, the CD Build pipeline automatically fires. It builds only the
services whose source files changed (path-filtered), signs the images with `cosign`,
and pushes them to ACR.

### 6.1 CD Build Pipeline Flow

```mermaid
flowchart TB
    subgraph TRIGGER["Trigger: push to main (or tag v*.*.*  or manual dispatch)"]
        PUSH["git merge → main"]
    end

    subgraph DETECT["Job: changes — Detect changed services"]
        FILTER["dorny/paths-filter\nchecks changed paths"]
        FORCE["If: shared pkg changed\n  OR workflow_dispatch\n  OR release tag\n→ rebuild ALL services"]
        FILTER --> FORCE
    end

    subgraph BUILD["Job: build-go (matrix over changed services)"]
        direction TB
        CHECKOUT["checkout"]
        OIDC_ACR["az login via OIDC\n(ci-plan identity)"]
        DOCKER["docker buildx build\n--platform linux/arm64\n(ARM v6 target cluster)"]
        TAG["Tag: {full-commit-sha}"]
        PUSH_IMG["docker push → ACR\nacrrtsabzkem9.azurecr.io/{svc}:{sha}"]
        SIGN["cosign sign --key less\n(keyless / Sigstore)"]
        SBOM["cosign attest SBOM\n(syft SPDX)"]
        TRIVY["trivy image scan\nSARIF → GitHub Security"]
        CHECKOUT --> OIDC_ACR --> DOCKER --> TAG --> PUSH_IMG --> SIGN --> SBOM --> TRIVY
    end

    subgraph BUILDWEB["Job: build-web (SolidJS + WebGPU)"]
        PNPM["pnpm install"]
        VITE["vite build\n--mode production"]
        PUSHW["push → ACR\nweb-cop-gpu:{sha}"]
        PNPM --> VITE --> PUSHW
    end

    PUSH --> DETECT --> BUILD
    DETECT --> BUILDWEB
```

### 6.2 Image Naming and the Promotion Contract

All images share the same commit SHA tag. This is the **promotion contract** — the exact
same digest validated in dev is what gets deployed to staging and then production.
No rebuilds happen during promotion. Only the Kubernetes config changes.

```
acrrtsabzkem9.azurecr.io/svc-radar-ingestion:<full-commit-sha>
acrrtsabzkem9.azurecr.io/svc-fusion-engine:<full-commit-sha>
acrrtsabzkem9.azurecr.io/svc-track:<full-commit-sha>
acrrtsabzkem9.azurecr.io/svc-webtransport:<full-commit-sha>
acrrtsabzkem9.azurecr.io/svc-query:<full-commit-sha>
acrrtsabzkem9.azurecr.io/web-cop-gpu:<full-commit-sha>
```

### 6.3 Platform Architecture: Why ARM64

The dev AKS cluster uses `Standard_D2pds_v6` nodes (Ampere ARM64). The reusable
container workflow configures QEMU and Buildx with `platforms: linux/arm64`; Go
Dockerfiles consume BuildKit's `TARGETARCH` for binaries and health probes. If an
x86 image reaches an ARM node, the pod fails with `exec format error`.

```bash
# Verify your local Docker can build ARM64 (requires QEMU emulation or Apple M-series)
docker buildx ls
# You should see linux/arm64 in the platforms list

# Test a local ARM64 build of one service
docker buildx build --platform linux/arm64 -t test-arm64 ./svc-radar-ingestion
```

### 6.4 System State After CD Build

```mermaid
flowchart TB
    subgraph ACR_STATE["ACR: acrrtsabzkem9 ✅ Images exist"]
        IMG1["svc-radar-ingestion:sha-a1b2c3d ✅"]
        IMG2["svc-fusion-engine:sha-a1b2c3d ✅"]
        IMG3["svc-track:sha-a1b2c3d ✅"]
        IMG4["svc-webtransport:sha-a1b2c3d ✅"]
        IMG5["svc-query:sha-a1b2c3d ✅"]
        IMG6["web-cop-gpu:sha-a1b2c3d ✅"]
    end

    subgraph COSIGN["Image Provenance ✅"]
        SIGN2["cosign signature attached\n(keyless / Sigstore)"]
        SBOM2["SBOM (SPDX) attached\nvia cosign attest"]
    end

    subgraph K8S2["Kubernetes — Still no app pods ❌"]
        NOTE["Images are in ACR but\nno Helm deploy has run yet"]
    end
```

---

## Section 7 — Step 6: CD Deploy to Dev

`cd-deploy.yml` runs `helm upgrade --install` against the AKS cluster. It uses the image
tag built in the previous step and targets the `rtsa` namespace.

### 7.1 Helm Chart Structure

```
deploy/charts/
├── rtsa-backbone/        # Redpanda (single-broker dev) + ClickHouse + OTel collector
├── rtsa-ingestion/       # Radar + AIS + EW + ELINT + ISR + Cyber ingestion services
├── rtsa-processing/      # Fusion engine + anomaly detection + feedback + audit
└── rtsa-presentation/    # svc-track · svc-query · svc-webtransport · web-cop-gpu
```

Each chart has a `values.yaml` (defaults) and per-environment overrides in
`deploy/charts/<chart>/values-<env>.yaml`.

### 7.2 CD Deploy Flow

```mermaid
flowchart TB
    subgraph TRIGGER2["Trigger: workflow_dispatch with image-tag input"]
        ITAG["image-tag: sha-a1b2c3d"]
    end

    subgraph AUTH2["Authentication (OIDC)"]
        AZLOG["az login\nusing AZURE_CLIENT_ID from env"]
        KUBE["az aks get-credentials\nAKS_CLUSTER_NAME · AKS_RESOURCE_GROUP"]
    end

    subgraph DEPLOY_STEPS["Helm Deploy (reusable: reusable-deploy-helm.yml)"]
        VERIFY["cosign verify image\n(reject unsigned images)"]
        NSCREATE["kubectl create namespace rtsa\n(if not exists)"]
        LABEL["kubectl label namespace rtsa\nistio.io/rev=asm-1-29"]
        HELM1["helm upgrade --install rtsa-backbone\n--set image.tag=sha-a1b2c3d"]
        HELM2["helm upgrade --install rtsa-ingestion\n--set image.tag=sha-a1b2c3d"]
        HELM3["helm upgrade --install rtsa-processing\n--set image.tag=sha-a1b2c3d"]
        HELM4["helm upgrade --install rtsa-presentation\n--set image.tag=sha-a1b2c3d"]
        SMOKE["verify-workload-deployment.sh\n(releases · rollouts · identity · images)"]
        VERIFY --> NSCREATE --> LABEL --> HELM1 --> HELM2 --> HELM3 --> HELM4 --> SMOKE
    end

    TRIGGER2 --> AUTH2 --> DEPLOY_STEPS
```

### 7.3 Bootstrap Development WebTransport Secrets

Before the first dev Helm deployment, create the JWT and TLS material directly in
Key Vault. The operator needs `Key Vault Secrets Officer` at the dev vault scope;
workloads retain the read-only `Key Vault Secrets User` role.

```bash
VAULT_ID=$(az keyvault show \
    --resource-group rg-rtsa-dev-cc \
    --name kv-rtsa-dev-aa67qb \
    --query id -o tsv)
OPERATOR_ID=$(az ad signed-in-user show --query id -o tsv)

az role assignment create \
    --assignee-object-id "$OPERATOR_ID" \
    --assignee-principal-type User \
    --role "Key Vault Secrets Officer" \
    --scope "$VAULT_ID"

scripts/azure/bootstrap-dev-key-vault-secrets.sh --env dev
```

The script refuses non-dev environments, validates the Terraform subscription and
vault, retains existing versions unless `--force` is explicit, and never prints
secret values. Use `--dry-run` to validate prerequisites only.

### 7.4 Trigger CD Deploy Manually

```bash
# Get the full SHA of the last successful CD Build
LAST_SHA=$(gh run list --workflow cd-build.yml --status success --limit 1 --json headSha --jq '.[0].headSha')

gh workflow run cd-deploy.yml \
  --repo arvinddhasmana/rtsa_webgpu \
  --field environment=dev \
    --field image-tag="$LAST_SHA"
```

### 7.5 Verify the Workload Deployment

The reusable Helm workflow automatically runs the workload verifier after all Helm
operations. Run the same gate locally when diagnosing or validating a deployment:

```bash
scripts/azure/verify-workload-deployment.sh \
    --namespace rtsa \
    --expected-image-tag "$LAST_SHA"
```

It verifies every expected Helm release, StatefulSet and Deployment rollout, pod health,
Istio sidecar injection, the WebTransport workload-identity annotation, and the promoted
image tag. Kubernetes warning events are printed for diagnosis but do not fail an otherwise
healthy deployment.

### 7.6 What Runs Inside Kubernetes After Deploy

```mermaid
flowchart TB
    subgraph NS_BACKBONE["namespace: rtsa-backbone"]
        RP_POD["Redpanda pod\n(StatefulSet, stateful pool)"]
        CH_POD["ClickHouse pod\n(StatefulSet, stateful pool)"]
        OTEL_POD["OTel Collector\n(DaemonSet)"]
    end

    subgraph NS_APP["namespace: rtsa"]
        ING_POD["svc-radar-ingestion\n(Deployment, ingest pool)"]
        FUSE_POD["svc-fusion-engine\n(Deployment, process pool)"]
        TRACK_POD["svc-track\n(Deployment, process pool)"]
        WT_POD["svc-webtransport\n(Deployment, process pool)"]
        QUERY_POD["svc-query\n(Deployment, process pool)"]
        COP_POD["web-cop-gpu\n(Deployment, process pool)"]
    end

    subgraph NS_MESH["Istio mesh — rtsa namespace (asm-1-29)"]
        SIDECAR["Envoy sidecar injected\nin every pod\nmTLS enforced"]
        GW_COLD["Istio Gateway\n(gRPC-Web cold path)"]
        GW_HOT["Standard LB UDP/443\n(WebTransport hot path)"]
    end

    RP_POD -->|"events"| FUSE_POD
    ING_POD -->|"publishes to Redpanda"| RP_POD
    FUSE_POD -->|"track events"| TRACK_POD
    TRACK_POD -->|"track state"| WT_POD
    TRACK_POD -->|"ClickHouse sink"| CH_POD
    WT_POD -->|"QUIC datagrams"| GW_HOT
    QUERY_POD -->|"gRPC-Web"| GW_COLD
```

### 7.6 Set the Frontend URLs in GitHub

After the first successful deploy, retrieve the ingress endpoint addresses and set the
last two `__SET_ME__` variables:

```bash
# Get the Istio gateway external IP
GATEWAY_IP=$(kubectl get svc istio-ingressgateway -n istio-system \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Get the WebTransport service external IP (UDP/443)
WT_IP=$(kubectl get svc svc-webtransport -n rtsa \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Set them in GitHub
gh variable set VITE_API_GATEWAY_URL \
  --repo arvinddhasmana/rtsa_webgpu \
  --env dev \
  --body "https://${GATEWAY_IP}"

gh variable set VITE_WEBTRANSPORT_URL \
  --repo arvinddhasmana/rtsa_webgpu \
  --env dev \
  --body "https://${WT_IP}:443"
```

### 7.7 System State After CD Deploy (dev)

```mermaid
flowchart TB
    subgraph FULL_STATE["Full System State: Dev Environment ✅"]
        direction TB

        subgraph SHARED3["rg-rtsa-shared-cc ✅"]
            ACR3["ACR: 6 service images"]
        end

        subgraph DEV3["rg-rtsa-dev-cc ✅"]
            AKS3["AKS aks-rtsa-dev\n(nodes healthy)"]
        end

        subgraph K8S3["Kubernetes: rtsa + rtsa-backbone ✅"]
            ALLPODS["All pods Running ✅\nIstio sidecars injected ✅\nmTLS between all services ✅"]
        end

        subgraph DATA3["Data Flow ✅"]
            FLOW["simulator → radar-ingestion\n→ Redpanda → fusion-engine\n→ track → webtransport\n→ COP browser (WebGPU)"]
        end

        subgraph GH3["GitHub ✅"]
            ALLVARS["All env vars set\n(including VITE_* URLs)"]
        end
    end
```

---

## Section 8 — Step 7: Promote to Test and Staging

Promotion means deploying the same image digest that passed in dev to a higher
environment. No code changes. No rebuilds. Just config and approval gates.

### 8.1 Full Promotion Flow

```mermaid
flowchart LR
    subgraph DEV_FLOW["Dev (auto on merge)"]
        D_INFRA["infra-up dev\n(if not running)"]
        D_BUILD["CD Build\n(on merge → main)"]
        D_DEPLOY["CD Deploy dev\n(auto trigger)"]
        D_TEST["automated smoke\n(5 min)"]
        D_INFRA --> D_BUILD --> D_DEPLOY --> D_TEST
    end

    subgraph TEST_FLOW["Test (auto after dev green)"]
        T_INFRA["infra-up test\n(ephemeral)"]
        T_DEPLOY["CD Deploy test\n(same sha)"]
        T_E2E["full SG-5 e2e\n(integration suite)"]
        T_DOWN["infra-down test\n(destroy)"]
        T_INFRA --> T_DEPLOY --> T_E2E --> T_DOWN
    end

    subgraph STG_FLOW["Staging (release tag only)"]
        S_TAG["git tag v1.2.3\n→ push → GitHub"]
        S_INFRA["infra-up staging\n(prod-like sizing)"]
        S_DEPLOY["CD Deploy staging\n(same sha)"]
        S_PERF["perf + chaos tests\n(30 min)"]
        S_GATE{"Manual approval\n(required reviewer)"}
        S_TAG --> S_INFRA --> S_DEPLOY --> S_PERF --> S_GATE
    end

    D_TEST -->|"green"| TEST_FLOW
    TEST_FLOW -->|"green + tag"| STG_FLOW
```

### 8.2 Create a Release Tag and Trigger Staging

```bash
# Ensure you are on main and up to date
git checkout main && git pull

# Tag the release
git tag v1.0.0
git push origin v1.0.0

# This triggers CD Build (force-rebuild all) + Staging promotion
# Monitor:
gh run list --workflow cd-build.yml --limit 3
```

### 8.3 Approve the Staging → Prod Gate

After staging tests pass, GitHub will pause the workflow at the `prod` environment
pending required reviewer approval.

1. Go to **Actions** tab in GitHub.
2. Find the pending workflow run.
3. Click **Review deployments**.
4. Select `prod` and click **Approve and deploy**.

---

## Section 9 — Step 8: Production Deployment

Production is identical to staging — same images, same Helm charts — with:

- On-demand (non-Spot) nodes across all pools
- Larger node sizes (configured in `staging.tfvars` / `prod.tfvars`)
- Purge protection enabled on Key Vault
- Managed Prometheus + Grafana enabled
- Multiple required reviewers on the GitHub Environment

### 9.1 Production Topology Compared to Dev

```mermaid
flowchart TB
    subgraph DEV_TOPO["Dev Cluster (lean)"]
        direction TB
        DSYS["system: 1-2 × D2pds_v6\nARM, Free SKU"]
        DING["ingest: 0-3 × D2pds_v6\nSpot, scale-to-zero"]
        DPROC["process: 0-3 × D2pds_v6\nSpot, scale-to-zero"]
        DSTAT["stateful: 1-2 × D2pds_v6\n128 GB disk"]
    end

    subgraph PROD_TOPO["Prod Cluster (HA)"]
        direction TB
        PSYS["system: 2-3 × D4pds_v6\nStandard SKU, HA"]
        PING["ingest: 2-10 × D4pds_v6\nOn-demand, autoscale"]
        PPROC["process: 2-10 × D4pds_v6\nOn-demand, autoscale"]
        PSTAT["stateful: 3 × D8pds_v6\nPremium SSD v2, ZRS"]
    end

    subgraph DIFFERENCES["Key Differences in prod.tfvars"]
        DIFFLIST["aks_sku_tier = 'Standard'\nsystem_node_min_count = 2\nenable_managed_prometheus = true\nenable_grafana = true\nkey_vault_purge_protection = true\nttl_hours = 0  ← never auto-destroy"]
    end
```

### 9.2 Production Pre-flight Checklist

Before approving prod:

```
□  All staging smoke tests pass (check Actions run summary)
□  Trivy scan shows no CRITICAL CVEs (check GitHub Security tab)
□  Change record raised (link in PR description)
□  On-call engineer notified
□  Rollback plan documented (previous image tag noted)
□  Cost estimate reviewed (08-cost-model-and-teardown.md)
```

### 9.3 Rollback

If prod deploy fails or causes a regression, roll back with the previous tag:

```bash
# Find the previous successful deploy tag
gh run list --workflow cd-deploy.yml --status success --limit 5

# Re-deploy previous image
gh workflow run cd-deploy.yml \
  --repo arvinddhasmana/rtsa_webgpu \
  --field environment=prod \
  --field image-tag="sha-{previous-sha}"
```

Helm keeps the previous release. `helm rollback` can also be used directly:

```bash
az aks get-credentials --resource-group rg-rtsa-prod-cc --name aks-rtsa-prod
helm history rtsa-presentation -n rtsa
helm rollback rtsa-presentation <previous-revision> -n rtsa
```

---

## Section 10 — Ongoing Operations

### 10.1 Cost Control: Tear Down Environments When Not Needed

The nightly schedule tears down `dev` and `test` automatically at 07:00 UTC:

```yaml
# infra-down.yml schedule:
schedule:
  - cron: "0 7 * * *" # tears down dev + test every night
```

Manual teardown:

```bash
gh workflow run infra-down.yml \
  --repo arvinddhasmana/rtsa_webgpu \
  --field environment=dev
```

Re-provision next morning:

```bash
gh workflow run infra-up.yml \
  --repo arvinddhasmana/rtsa_webgpu \
  --field environment=dev
```

> Rebuild time: ~15 min. All images remain in ACR. State is preserved in the Storage Account.
> The cluster starts empty — run CD Deploy again with the same SHA to restore application pods.

### 10.2 Environment Lifecycle Summary

```mermaid
stateDiagram-v2
    [*] --> NotProvisioned : initial state

    NotProvisioned --> Provisioning : infra-up triggered
    Provisioning --> PlatformReady : terraform apply complete\n(AKS healthy, no app pods)
    PlatformReady --> AppDeployed : CD Deploy complete\n(all pods Running)
    AppDeployed --> AppDeployed : new deploys via CD Deploy\n(rolling update)
    AppDeployed --> Tearing : infra-down triggered\nor nightly schedule
    PlatformReady --> Tearing : infra-down triggered
    Tearing --> NotProvisioned : terraform destroy complete\n(rg-rtsa-ENV deleted)\n(ACR + state preserved)

    note right of NotProvisioned
        GitHub env variables\nstill exist and are correct.\nNo Azure cost.
    end note

    note right of AppDeployed
        Full system running.\nAzure cost: ~$4-8/day (dev)\nwith Spot nodes.
    end note
```

### 10.3 Monitoring and Observability

```bash
# Check pod status across all namespaces
kubectl get pods -A

# View logs for a specific service
kubectl logs -n rtsa deployment/svc-radar-ingestion --follow

# Check Redpanda topic lag (KEDA watches this to scale)
kubectl exec -n rtsa-backbone -it redpanda-0 -- \
  rpk topic list

# Open Grafana (dev: port-forward since Managed Grafana is disabled in dev)
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Then open http://localhost:3000

# View Istio mesh metrics
kubectl port-forward -n istio-system svc/kiali 20001:20001
```

### 10.4 Troubleshooting Common Issues

#### Terraform: `state blob is already locked`

A previous Terraform process was killed mid-apply. The Azure Blob lease is held.

```bash
# Step 1: Kill any lingering Terraform processes
pkill -9 -f "terraform-provider-azurerm"
sleep 3

# Step 2: Break the blob lease using the storage account key
ACCT_KEY=$(az storage account keys list \
  --account-name strtsatfbzkem9 \
  --resource-group rg-rtsa-shared-cc \
  --query '[0].value' -o tsv)

az storage blob lease break \
    --blob-name "dev.tfstate" \
  --container-name "tfstate" \
  --account-name "strtsatfbzkem9" \
  --account-key "$ACCT_KEY"

# Step 3: Force-unlock using the lock ID from the error message
terraform -chdir=infra/terraform/environments/dev \
  force-unlock -force <LOCK-ID-FROM-ERROR>

# Step 4: Retry plan
terraform -chdir=infra/terraform/environments/dev plan -var-file=dev.tfvars
```

Or use the helper script:

```bash
scripts/azure/tf-break-lock.sh \
  --environment dev \
  --storage-account strtsatfbzkem9 \
  --container tfstate
```

#### AKS node not starting: `Insufficient vcpu quota`

This subscription (canadacentral) restricts to ARM VMs and has limited vCPU quota.
Only `Standard_D2pds_v6` (family: `Dpdsv6`) has available quota.
Check your current quota:

```bash
az vm list-usage --location canadacentral \
  --query "[?contains(name.value,'Dpds')].{Name:name.value, Used:currentValue, Limit:limit}" \
  -o table
```

If all 10 vCPUs are used, tear down the environment first:

```bash
gh workflow run infra-down.yml --field environment=dev
# Then wait ~5 min for resources to release, then re-provision
```

#### Pod in `CrashLoopBackOff`

```bash
# See why
kubectl describe pod <pod-name> -n rtsa

# Check logs (including init containers)
kubectl logs <pod-name> -n rtsa --previous

# Most common causes:
# 1. Missing Key Vault secret — check CSI driver logs
kubectl logs -n kube-system -l app=secrets-store-csi-driver

# 2. Wrong Workload Identity client ID
kubectl get sa -n rtsa svc-webtransport -o yaml | grep azure.workload

# 3. Image not found in ACR — check if CD Build ran
az acr repository show-tags --name acrrtsabzkem9 --repository svc-webtransport
```

#### CI failing: `Missing 'CLASSIFICATION: UNCLASSIFIED' header`

Every `.go`, `.ts`, `.yaml`, `.yml`, `.tf`, `.sh`, and `.md` file must have the
classification header as its **first line** (or within the first 5 lines):

```go
// CLASSIFICATION: UNCLASSIFIED
package main
```

```yaml
# CLASSIFICATION: UNCLASSIFIED
name: My Workflow
```

```hcl
# CLASSIFICATION: UNCLASSIFIED
variable "name" {}
```

---

## Section 11 — Complete State Reference

This diagram shows what exists after every step has been completed successfully.
Use it as a verification checklist.

```mermaid
flowchart TB
    subgraph STEP0["After Step 1 — Bootstrap"]
        B1["✅ rg-rtsa-shared-cc created"]
        B2["✅ ACR: acrrtsabzkem9"]
        B3["✅ Terraform state SA: strtsatfbzkem9"]
        B4["✅ 4 Managed Identities + OIDC federated creds"]
        B5["✅ Log Analytics: log-rtsa-shared-cc"]
    end

    subgraph STEP1["After Step 2 — GitHub Setup"]
        G1["✅ GitHub Environments: dev, test, staging, prod"]
        G2["✅ Repo vars: AZURE_TENANT_ID, SUBSCRIPTION_ID, ACR_NAME, TFSTATE_*"]
        G3["✅ Dev env: AZURE_CLIENT_ID = 698f32f3-..."]
        G4["⚠️  AKS_CLUSTER_NAME = __SET_ME__ (needs infra-up)"]
    end

    subgraph STEP2["After Step 3 — infra-up dev"]
        I1["✅ rg-rtsa-dev-cc created"]
        I2["✅ AKS: aks-rtsa-dev (3 node pools healthy)"]
        I3["✅ KV: kv-rtsa-dev-rg7auf"]
        I4["✅ Storage: strtsadevrg7auf"]
        I5["✅ 3 Workload Managed Identities"]
        I6["✅ GitHub dev vars: AKS_CLUSTER_NAME, KEY_VAULT_NAME, ISTIO_REVISION"]
    end

    subgraph STEP3["After Step 4+5 — CI + CD Build"]
        C1["✅ PR gates passing (SG-1..SG-5)"]
        C2["✅ Images in ACR: sha-{commit}"]
        C3["✅ cosign signature + SBOM attached"]
        C4["✅ SARIF scan uploaded to GitHub Security"]
    end

    subgraph STEP4["After Step 6+7 — CD Deploy dev + set VITE URLs"]
        D1["✅ All pods Running in rtsa namespace"]
        D2["✅ Istio mTLS enforced (asm-1-29)"]
        D3["✅ Backbone (Redpanda + ClickHouse) healthy"]
        D4["✅ VITE_API_GATEWAY_URL set"]
        D5["✅ VITE_WEBTRANSPORT_URL set"]
        D6["✅ COP loads in browser, tracks visible"]
    end

    subgraph STEP5["After Step 8+9 — Promoted to Staging → Prod"]
        P1["✅ Release tag vX.Y.Z pushed"]
        P2["✅ Staging tests + chaos/perf passed"]
        P3["✅ Manual approval recorded"]
        P4["✅ Same image digest running in prod"]
        P5["✅ Prod monitoring (Managed Prometheus/Grafana) active"]
    end

    STEP0 --> STEP1 --> STEP2 --> STEP3 --> STEP4 --> STEP5
```

---

## Quick Reference Card

Keep this handy. These are the 10 commands you will run most often.

```bash
# 1. Log in
az login && gh auth login

# 2. Bootstrap (one-time)
scripts/azure/bootstrap.sh

# 3. Create GitHub environments (one-time)
scripts/azure/setup-github-environments.sh --repo arvinddhasmana/rtsa_webgpu

# 4. Provision dev infrastructure
gh workflow run infra-up.yml --field environment=dev

# 5. Sync Terraform outputs to GitHub variables
scripts/azure/sync-dev-github-vars-from-tf.sh \
  --repo arvinddhasmana/rtsa_webgpu --environment dev

# 6. Connect kubectl to dev
az aks get-credentials --resource-group rg-rtsa-dev-cc --name aks-rtsa-dev

# 7. Check what's running
kubectl get pods -A

# 8. Trigger a deploy manually
gh workflow run cd-deploy.yml \
  --field environment=dev --field image-tag="sha-$(git rev-parse --short HEAD)"

# 9. Tear down dev (saves ~$6/day)
gh workflow run infra-down.yml --field environment=dev

# 10. Release to staging/prod
git tag v1.0.0 && git push origin v1.0.0
```

---

_For architecture details, see [02-target-architecture.md](02-target-architecture.md)._
_For cost breakdown, see [08-cost-model-and-teardown.md](08-cost-model-and-teardown.md)._
_For troubleshooting Terraform state issues in depth, see [09-operator-runbook-manual-validation.md](09-operator-runbook-manual-validation.md)._
