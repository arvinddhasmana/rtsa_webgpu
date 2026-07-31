## 2. AAD Backend Authentication Example

Terraform resources are deployed into nonprod, but Terraform state is stored in the shared subscription:

```text
Operator login
    |
    +-- Terraform provider --> nonprod subscription --> AKS, VNet, Key Vault
    |
    +-- Terraform backend  --> shared subscription --> tfstate/test.tfstate
```

These are separate authorization paths.

For example, your account may have `Contributor` on nonprod, allowing it to create AKS resources, but still receive this during `terraform init`:

```text
AuthorizationPermissionMismatch
```

That happens because `Contributor` is a management-plane role. Reading and writing state blobs requires the data-plane role `Storage Blob Data Contributor`.

### Grant your operator access

```bash
export SHARED_SUB="11f614f9-a6d3-419b-9437-37a84c75f27a"
export TFSTATE_RG="rg-rtsa-shared-cc"
export TFSTATE_SA="strtsatfbzkem9"

OPERATOR_ID=$(az ad signed-in-user show --query id -o tsv)

TFSTATE_SCOPE=$(az storage account show \
  --subscription "$SHARED_SUB" \
  --resource-group "$TFSTATE_RG" \
  --name "$TFSTATE_SA" \
  --query id -o tsv)

az role assignment create \
  --assignee-object-id "$OPERATOR_ID" \
  --assignee-principal-type User \
  --role "Storage Blob Data Contributor" \
  --scope "$TFSTATE_SCOPE"
```

Allow several minutes for RBAC propagation, then test AAD access directly:

```bash
az storage blob list \
  --subscription "$SHARED_SUB" \
  --account-name "$TFSTATE_SA" \
  --container-name tfstate \
  --auth-mode login \
  --num-results 1 \
  -o table
```

Finally, remove the access-key fallback and prove Terraform works using your Entra token:

```bash
unset TFSTATE_ACCESS_KEY

export ARM_SUBSCRIPTION_ID="$NONPROD_SUB"

scripts/azure/preflight-environment-deploy.sh --env test
make -C infra/terraform env-plan ENV=test
```

When `TFSTATE_ACCESS_KEY` is unset, the Makefile automatically uses `use_azuread_auth=true`. This is preferable because no storage key is handled, copied, or exposed. GitHub deployer identities already receive the same data-plane role through identity.tf.
