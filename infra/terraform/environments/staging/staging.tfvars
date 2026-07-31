# CLASSIFICATION: UNCLASSIFIED
# Staging environment values. No secrets here — subscription comes from ARM_SUBSCRIPTION_ID.
#
# Fill the three "shared foundation" names from `make bootstrap-output` after P0.

project     = "rtsa"
environment = "staging"
location    = "canadacentral"

# ── From P0 bootstrap outputs (replace the ACR placeholder) ───────────────────
shared_subscription_id = "11f614f9-a6d3-419b-9437-37a84c75f27a"
expected_subscription_id = "bb2b8549-9693-40f2-9287-3bd5afcc6633"
shared_resource_group = "rg-rtsa-shared-cc"
acr_name              = "acrrtsabzkem9" # set from bootstrap output acr_name
law_name              = "log-rtsa-shared-cc"

# ── AKS (staging profile) ─────────────────────────────────────────────────────
# Staging is production-like but still cost-aware.
aks_sku_tier          = "Standard"
system_node_vm_size   = "Standard_D2pds_v6" # keep ARM compatibility with current service images
system_node_min_count = 2
system_node_max_count = 3
system_node_zones     = [] # canadacentral: no AZ support for this SKU

# Bulkhead pools: run on regular capacity in staging to reduce pre-prod drift.
additional_node_pools = {
  ingest = {
    vm_size   = "Standard_D2pds_v6"
    min_count = 1
    max_count = 4
    priority  = "Regular"
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "ingestion" }
    taints    = ["workload=ingestion:NoSchedule"]
  }
  process = {
    vm_size   = "Standard_D2pds_v6"
    min_count = 1
    max_count = 4
    priority  = "Regular"
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "processing" }
    taints    = ["workload=processing:NoSchedule"]
  }
  stateful = {
    vm_size         = "Standard_D2pds_v6"
    min_count       = 2
    max_count       = 3
    priority        = "Regular"
    os_disk_size_gb = 256
    zones           = [] # canadacentral: no AZ support
    labels          = { workload = "stateful" }
    taints          = ["workload=stateful:NoSchedule"]
  }
}

# ── Workload identities for the walking skeleton (P2) ─────────────────────────
istio_revisions = ["asm-1-29"] # supported on K8s 1.31-1.36; asm-1-23 was removed from canadacentral
workloads = {
  webtransport = { namespace = "rtsa", service_account = "svc-webtransport" }
  track        = { namespace = "rtsa", service_account = "svc-track" }
  query        = { namespace = "rtsa", service_account = "svc-query" }
}

# ── Cost/lifecycle ────────────────────────────────────────────────────────────
key_vault_purge_protection = false
enable_managed_prometheus  = true
enable_grafana             = true
ttl_hours                  = 72
