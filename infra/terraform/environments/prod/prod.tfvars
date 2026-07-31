# CLASSIFICATION: UNCLASSIFIED
# Production environment values. No secrets here — subscription comes from ARM_SUBSCRIPTION_ID.
#
# Fill the three "shared foundation" names from `make bootstrap-output` after P0.

project     = "rtsa"
environment = "prod"
location    = "canadacentral"

# ── From P0 bootstrap outputs (replace the ACR placeholder) ───────────────────
shared_subscription_id = "11f614f9-a6d3-419b-9437-37a84c75f27a"
expected_subscription_id = "07832ea8-0aa8-4091-840c-184ac5e6306a"
shared_resource_group = "rg-rtsa-shared-cc"
acr_name              = "acrrtsabzkem9" # set from bootstrap output acr_name
law_name              = "log-rtsa-shared-cc"

# ── AKS (production profile) ──────────────────────────────────────────────────
# Production keeps ARM node compatibility but uses stronger baseline capacity.
aks_sku_tier          = "Standard"
system_node_vm_size   = "Standard_D2pds_v6"
system_node_min_count = 2
system_node_max_count = 4
system_node_zones     = [] # canadacentral: no AZ support for this SKU

# Bulkhead pools: all regular capacity in prod for deterministic behavior.
additional_node_pools = {
  ingest = {
    vm_size   = "Standard_D2pds_v6"
    min_count = 2
    max_count = 6
    priority  = "Regular"
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "ingestion" }
    taints    = ["workload=ingestion:NoSchedule"]
  }
  process = {
    vm_size   = "Standard_D2pds_v6"
    min_count = 2
    max_count = 6
    priority  = "Regular"
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "processing" }
    taints    = ["workload=processing:NoSchedule"]
  }
  stateful = {
    vm_size         = "Standard_D2pds_v6"
    min_count       = 2
    max_count       = 4
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
ttl_hours                  = 0
