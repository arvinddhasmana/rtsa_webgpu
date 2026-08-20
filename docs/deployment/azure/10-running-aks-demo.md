<!-- CLASSIFICATION: UNCLASSIFIED -->

# Running A Finite AKS Demo

This runbook runs the synthetic RTSA simulator against services already deployed
in the Azure Dev environment. It is deliberately separate from GitHub Actions:
demo data generation is an operator activity, not a build or deployment trigger.

## Prerequisites

- `az`, `kubectl`, and `git` installed in the Azure-connected operator environment.
- An authenticated Azure CLI session with permission to read the Dev AKS cluster,
  create/delete Jobs in namespace `rtsa`, and build/push to the shared ACR.
- AKS is already connected to ACR and the RTSA Dev services are healthy.
- Only synthetic, UNCLASSIFIED data is used. Never point this tool at real feeds.

Authenticate and select the cluster using the values from the Dev GitHub Environment:

```bash
az login
az account set --subscription "$AZURE_SUBSCRIPTION_ID"
az aks get-credentials --resource-group "$AKS_RESOURCE_GROUP" \
  --name "$AKS_CLUSTER_NAME" --overwrite-existing
kubelogin convert-kubeconfig -l azurecli
kubectl get pods -n rtsa
```

## Run A Demo

The script builds the simulator image in ACR with the current commit SHA, then
creates a Kubernetes Job. The Job runs for a finite duration and receives service
endpoints through cluster DNS, for example `svc-ew-ingestion:50051`.

```bash
bash scripts/demo/run-azure-demo.sh \
  --acr "$ACR_NAME" \
  --scenario maritime-demo.yaml \
  --duration 20 \
  --cleanup
```

Use `--image` with an existing immutable image to skip the ACR build. Use a
different scenario such as `multi-domain-demo.yaml` or `fusion-dashboard-demo.yaml`
for another synthetic showcase. Job logs are printed when the run completes.

## Cleanup And Cost Control

The Job is configured with `backoffLimit: 0`, an active deadline, and a one-hour
`ttlSecondsAfterFinished` value. `--cleanup` deletes it immediately
after success or failure. If the terminal disconnects, remove abandoned Jobs with:

```bash
bash scripts/demo/cleanup-azure-demo.sh
```

To remove one run only:

```bash
bash scripts/demo/cleanup-azure-demo.sh --job-name rtsa-demo-YYYYMMDDHHMMSS
```

This cleanup only removes simulator Jobs. It does not tear down the Dev cluster,
RTSA services, Redpanda, ClickHouse, or their volumes. Use the existing
environment lifecycle runbook for infrastructure teardown.
