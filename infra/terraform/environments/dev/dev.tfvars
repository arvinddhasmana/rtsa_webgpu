# CLASSIFICATION: UNCLASSIFIED
# Dev environment values. No secrets here — subscription comes from ARM_SUBSCRIPTION_ID.
#
# Fill the three "shared foundation" names from `make bootstrap-output` after P0.

project     = "rtsa"
environment = "dev"
location    = "canadacentral"

# ── From P0 bootstrap outputs (replace the ACR placeholder) ───────────────────
shared_subscription_id = "11f614f9-a6d3-419b-9437-37a84c75f27a"
expected_subscription_id = "bb2b8549-9693-40f2-9287-3bd5afcc6633"
shared_resource_group = "rg-rtsa-shared-cc"
acr_name              = "acrrtsabzkem9" # set from bootstrap output acr_name
law_name              = "log-rtsa-shared-cc"

# ── AKS (lean dev) ────────────────────────────────────────────────────────────
# canadacentral constraints:
#   • No AZ support for these VM SKUs → zones = []
#   • Subscription restricts to ARM-architecture VMs (Ampere / p-series)
#   • Only "Standard Dpdsv6 Family" has quota (0/10 vCPU) → use Standard_D2pds_v6
#   • Total Regional vCPUs = 10; min cluster = system(1×2) + stateful(1×2) = 4 vCPU
aks_sku_tier          = "Free"
system_node_vm_size   = "Standard_D2pds_v6" # ARM v6; quota: Dpdsv6 family = 0/10
system_node_min_count = 1
system_node_max_count = 2
system_node_zones     = [] # canadacentral: no AZ support for this SKU

# Bulkhead pools: ingestion/processing on Spot (scale-to-zero), stateful on-demand.
# All pools use D2pds_v6 to stay within the 10 total regional vCPU limit.
additional_node_pools = {
  ingest = {
    vm_size   = "Standard_D2pds_v6" # ARM v6; scale-to-zero (min=0 so no vCPU at idle)
    min_count = 0
    max_count = 1
    priority  = "Spot"
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "ingestion" }
    taints    = ["workload=ingestion:NoSchedule"]
  }
  process = {
    vm_size   = "Standard_D2pds_v6" # ARM v6; scale-to-zero (min=0 so no vCPU at idle)
    min_count = 0
    max_count = 1
    priority  = "Spot"
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "processing" }
    taints    = ["workload=processing:NoSchedule"]
  }
  stateful = {
    vm_size         = "Standard_D2pds_v6" # ARM v6; D2 (not D4) to stay under 10-vCPU regional limit
    min_count       = 1
    max_count       = 1
    priority        = "Regular"
    os_disk_size_gb = 128
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
enable_managed_prometheus  = false
enable_grafana             = false
ttl_hours                  = 8
