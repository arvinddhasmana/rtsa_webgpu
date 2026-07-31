<!-- CLASSIFICATION: UNCLASSIFIED -->

# 12 - Multi-Subscription Implementation Plan (Phase-Wise Execution)

> Parent: [README](README.md) · Prev: [11 Multi-Subscription Architecture Model](11-multi-subscription-analysis.md)
> Classification: UNCLASSIFIED · Status: Implementation planning

---

## 1. Objective and Scope

Implement the environment-isolated subscription model as the standard Azure
deployment pattern:

- Dev subscription
- Test subscription
- Staging subscription
- Prod subscription
- Shared subscription for common assets (ACR, tfstate storage, shared LAW)

This plan keeps the existing CI/CD architecture and promotion-by-digest behavior.

---

## 2. Phase Map

```mermaid
flowchart LR
    P0[Phase 0\nDesign freeze and inventory] --> P1[Phase 1\nVariable model and workflow wiring]
    P1 --> P2[Phase 2\nBootstrap identity and RBAC refactor]
    P2 --> P3[Phase 3\nTerraform provider and backend split]
    P3 --> P4[Phase 4\nEnvironment roots expansion]
    P4 --> P5[Phase 5\nPipeline validation and promotion drill]
    P5 --> P6[Phase 6\nProd cutover and hardening]
```

---

## 3. Detailed Phases

## Phase 0 - Design Freeze and Inventory

### Purpose

Create an approved source of truth for subscription mapping and ownership before
any code changes.

### Inputs

1. Subscription ids for shared, dev, test, staging, prod
2. Ownership and approver contacts per environment
3. Required RBAC policy constraints

### Tasks

1. Define canonical mapping table:
   - env name
   - subscription id
   - deployer identity name
   - approvals and reviewers
2. Validate policy constraints with cloud governance owners.
3. Freeze naming conventions and tagging keys.

### Exit Criteria

1. Mapping table approved.
2. Security and platform owners sign off.

### Deliverables

1. Updated architecture decision entry in [03-technology-options-and-decisions.md](03-technology-options-and-decisions.md)
2. Migration tracker checklist in platform board

---

## Phase 1 - Variable Model and Workflow Wiring

### Purpose

Define the environment-scoped subscription and shared build wiring used by all
workflows.

### Tasks

1. Define environment-level `AZURE_SUBSCRIPTION_ID` values for dev, test, staging,
   and prod.
2. Add shared subscription variable(s) for build/state tasks where needed.
3. Update helper scripts so they seed the correct scope for each variable.
4. Validate the workflow model against the shared + per-environment subscription
   split.

### Files in scope

1. [../../.github/workflows/infra-up.yml](../../.github/workflows/infra-up.yml)
2. [../../.github/workflows/infra-down.yml](../../.github/workflows/infra-down.yml)
3. [../../.github/workflows/cd-build.yml](../../.github/workflows/cd-build.yml)
4. [../../.github/workflows/\_reusable/deploy-helm.yml](../../.github/workflows/_reusable/deploy-helm.yml)
5. [../../scripts/azure/setup-github-environments.sh](../../scripts/azure/setup-github-environments.sh)

### Validation

1. Run infra-up in dev with env-level subscription var.
2. Run cd-build and confirm ACR push still succeeds.
3. Confirm no job depends on removed repo-level value.

### Exit Criteria

1. All workflows pass in dev using the target variable model.
2. No deploy job depends on a repo-level subscription variable.

---

## Phase 2 - Bootstrap Identity and RBAC Refactor

### Purpose

Scope identities and roles to the correct subscription per environment.

### Tasks

1. Extend bootstrap variables to support environment-to-subscription mapping.
2. Refactor identity module logic to assign:
   - env deploy identity roles in env subscription scope
   - shared resource roles in shared subscription scope
3. Keep CI plan identity with minimum read permissions only where required.
4. Add verification outputs for assigned scopes and principal ids.

### Files in scope

1. [../../infra/terraform/bootstrap/variables.tf](../../infra/terraform/bootstrap/variables.tf)
2. [../../infra/terraform/bootstrap/identity.tf](../../infra/terraform/bootstrap/identity.tf)
3. [../../infra/terraform/bootstrap/outputs.tf](../../infra/terraform/bootstrap/outputs.tf)
4. [../../infra/terraform/bootstrap/terraform.tfvars.example](../../infra/terraform/bootstrap/terraform.tfvars.example)

### Validation

1. Terraform plan shows role assignments per correct subscription.
2. OIDC login tests pass for each environment identity.
3. ACR and tfstate access checks pass from each env identity.

### Exit Criteria

1. Bootstrap apply completes with per-env scope assignments.
2. Evidence captured for least privilege review.

---

## Phase 3 - Terraform Provider and Backend Split

### Purpose

Allow environment roots to create environment resources in env subscription while
reading shared resources from shared subscription.

### Tasks

1. Add provider aliases in environment roots:
   - azurerm.env
   - azurerm.shared
2. Bind shared data sources to shared provider alias.
3. Keep environment resources and modules on env provider.
4. Add explicit backend subscription configuration in init path.

### Files in scope

1. [../../infra/terraform/environments/dev/providers.tf](../../infra/terraform/environments/dev/providers.tf)
2. [../../infra/terraform/environments/dev/main.tf](../../infra/terraform/environments/dev/main.tf)
3. [../../infra/terraform/environments/dev/variables.tf](../../infra/terraform/environments/dev/variables.tf)
4. [../../.github/workflows/infra-up.yml](../../.github/workflows/infra-up.yml)
5. [../../.github/workflows/infra-down.yml](../../.github/workflows/infra-down.yml)

### Validation

1. terraform init works with backend in shared subscription.
2. terraform plan resolves shared ACR and LAW via shared provider.
3. terraform apply completes in dev env subscription.

### Exit Criteria

1. Cross-subscription provider pattern proven in dev.
2. No hidden same-subscription assumptions remain in root module.

---

## Phase 4 - Environment Roots Expansion

### Purpose

Create and validate test, staging, prod roots using the proven multi-subscription pattern.

### Tasks

1. Add folders:
   - infra/terraform/environments/test
   - infra/terraform/environments/staging
   - infra/terraform/environments/prod
2. Create tfvars for each environment with appropriate sizing and policy.
3. Set environment-specific workflow vars (subscription id, client id, cluster vars).
4. Add optional policy checks for production constraints.

### Validation

1. `infra-up` and `infra-down` succeed in test, with `verify-infrastructure-deployment.sh --env test` passing after apply.
2. `infra-up` succeeds in staging and prod with approval gates and the same environment-specific verification.
3. State keys and resource names are isolated per environment.

### Exit Criteria

1. All four environments provision and destroy as designed.
2. Promotion path can be executed across all environment subscriptions.

---

## Phase 5 - Pipeline Validation and Promotion Drill

### Purpose

Prove end-to-end CI/CD behavior with digest promotion across different subscriptions.

### Tasks

1. Run a full rehearsal using one candidate release sha.
2. Execute sequence:
   - CI on PR
   - merge to main
   - cd-build
   - deploy dev
   - deploy test
   - deploy staging
   - approval
   - deploy prod
3. Validate all identity exchanges and AKS credential retrieval per subscription.

### Validation

1. `verify-workload-deployment.sh --expected-image-tag <release-tag>` passes in each environment and proves the same promoted image tag is deployed.
2. No cross-environment subscription mix-up.
3. Audit logs match expected identity usage.

### Exit Criteria

1. Rehearsal report approved.
2. Runbook updates completed.

---

## Phase 6 - Production Cutover and Hardening

### Purpose

Complete the rollout and lock controls.

### Tasks

1. Enable strict protection rules for staging and prod.
2. Remove transitional fallback logic that is no longer needed.
3. Finalize alerting and budget rules per subscription.
4. Execute post-rollout review and close implementation action items.

### Validation

1. Two consecutive prod releases succeed after rollout.
2. Rollback path validated post-cutover.

### Exit Criteria

1. Migration accepted by platform and security owners.
2. Transitional assumptions retired.

---

## 4. Required Documentation Changes in Existing Files

The following updates are required to keep documents aligned with implementation.

| File                                                                                 | Current issue                                                | Required update                                                         |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------ | ----------------------------------------------------------------------- |
| [README.md](README.md)                                                               | index does not include multi-sub docs                        | add references to docs 11 and 12; update reading flow                   |
| [02-target-architecture.md](02-target-architecture.md)                               | baseline diagram needs shared + per-env subscription framing | update diagram and landing-zone wording                                 |
| [05-devops-cicd-and-environments.md](05-devops-cicd-and-environments.md)             | examples need explicit build/deploy split                    | document env-level subscription variables and shared build subscription |
| [06-iac-terraform-and-lifecycle.md](06-iac-terraform-and-lifecycle.md)               | provider/backend narrative needs env and shared aliases      | add provider alias and backend subscription guidance                    |
| [09-operator-runbook-manual-validation.md](09-operator-runbook-manual-validation.md) | runbook is environment-driven but needs validation examples  | add per-environment verification commands                               |
| [10-getting-started-guide.md](10-getting-started-guide.md)                           | onboarding flow needs pure multi-sub setup steps             | document environment setup and shared build setup                       |

---

## 5. Change Package Recommendation

To reduce risk, split implementation into small pull requests:

1. PR-A: workflow vars and script changes
2. PR-B: bootstrap identity and RBAC mapping
3. PR-C: provider alias plus backend split in dev
4. PR-D: add test/staging/prod roots
5. PR-E: documentation and runbook finalization

Each PR should include:

1. explicit rollback steps
2. evidence artifacts
3. updated docs in docs/deployment/azure

---

## 6. Rollback Strategy per Phase

| Phase   | Rollback approach                                                            |
| ------- | ---------------------------------------------------------------------------- |
| Phase 1 | revert environment and shared variable updates                               |
| Phase 2 | restore previous bootstrap state version and role assignment set             |
| Phase 3 | revert provider alias changes and return to single provider context          |
| Phase 4 | disable new environments and keep dev only                                   |
| Phase 5 | stop at staging and do not approve prod deployment                           |
| Phase 6 | revert workflow protection and subscription mapping to the prior release tag |

---

## 7. Success Metrics

1. 100 percent environment deployments use the correct environment subscription id.
2. Zero production deployments executed by a non-prod identity.
3. No tfstate backend auth failures after rollout.
4. Promotion drill completed twice without manual credential fixes.
5. Documentation and runbook accuracy validated by a new team member dry run.

---

## 8. Immediate Next Steps

1. Approve target design option B in [11-multi-subscription-analysis.md](11-multi-subscription-analysis.md).
2. Start Phase 1 changes as PR-A.
3. Update impacted docs in same PR stream to prevent drift.
