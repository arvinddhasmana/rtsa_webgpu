# CLASSIFICATION: UNCLASSIFIED
# Dev environment values. No secrets here — subscription comes from ARM_SUBSCRIPTION_ID.
#
# Fill the three "shared foundation" names from `make bootstrap-output` after P0.

project     = "rtsa"
environment = "dev"
location    = "canadacentral"

# ── From P0 bootstrap outputs (replace the ACR placeholder) ───────────────────
shared_resource_group = "rg-rtsa-shared-cc"
acr_name              = "acrrtsaCHANGME" # e.g. acrrtsa3f9a2b (see bootstrap output acr_name)
law_name              = "log-rtsa-shared-cc"

# ── AKS (lean dev) ────────────────────────────────────────────────────────────
aks_sku_tier          = "Free"
system_node_vm_size   = "Standard_D2s_v5"
system_node_min_count = 1
system_node_max_count = 2

# Bulkhead pools: ingestion/processing on Spot (scale-to-zero), stateful on-demand.
additional_node_pools = {
  ingest = {
    vm_size   = "Standard_D2s_v5"
    min_count = 0
    max_count = 3
    priority  = "Spot"
    labels    = { workload = "ingestion" }
    taints    = ["workload=ingestion:NoSchedule"]
  }
  process = {
    vm_size   = "Standard_D2s_v5"
    min_count = 0
    max_count = 3
    priority  = "Spot"
    labels    = { workload = "processing" }
    taints    = ["workload=processing:NoSchedule"]
  }
  stateful = {
    vm_size         = "Standard_D4s_v5"
    min_count       = 1
    max_count       = 3
    priority        = "Regular"
    os_disk_size_gb = 128
    labels          = { workload = "stateful" }
    taints          = ["workload=stateful:NoSchedule"]
  }
}

# ── Workload identities for the walking skeleton (P2) ─────────────────────────
workloads = {
  webtransport = { namespace = "rtsa", service_account = "svc-webtransport" }
  track        = { namespace = "rtsa", service_account = "svc-track" }
  query        = { namespace = "rtsa", service_account = "svc-query" }
}

# ── Cost/lifecycle ────────────────────────────────────────────────────────────
key_vault_purge_protection = false
enable_managed_prometheus  = false
enable_grafana             = false
ttl_hours                  = 8
