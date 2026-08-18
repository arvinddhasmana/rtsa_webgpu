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
#   • Standard Dpdsv6 Family regular quota = 20 vCPU (plenty of headroom)
#   • Total Regional Low-priority (Spot) vCPUs is capped at 3 — a subscription-wide
#     default identical in every region (verified: eastus/eastus2/centralus/westus2/
#     canadaeast all show the same 3-vCPU cap). A single scale-to-zero Spot pool
#     already uses 2 of those 3 vCPUs, leaving no headroom for a second Spot pool
#     to scale up — this is what caused "FailedScheduling"/"backoff after failed
#     scale-up" for svc-fusion-engine/svc-track/svc-webtransport. Moving region
#     does NOT fix this; only a quota increase request or dropping Spot does.
#     Bulkhead pools below use Regular priority to sidestep the Spot cap entirely.
aks_sku_tier          = "Free"
system_node_vm_size   = "Standard_D2pds_v6" # ARM v6; quota: Dpdsv6 family = 6/20
system_node_min_count = 1
system_node_max_count = 3
system_node_zones     = [] # canadacentral: no AZ support for this SKU

# Bulkhead pools: ingestion/processing/stateful all on-demand (Regular priority).
# All pools use D2pds_v6; regular Dpdsv6 quota (20 vCPU) has ample headroom.
additional_node_pools = {
  ingest = {
    vm_size   = "Standard_D2pds_v6" # ARM v6; scale-to-zero (min=0 so no vCPU at idle)
    min_count = 0
    max_count = 1
    priority  = "Regular" # was Spot; the 3-vCPU low-priority cap blocked concurrent scale-up
    zones     = [] # canadacentral: no AZ support
    labels    = { workload = "ingestion" }
    taints    = ["workload=ingestion:NoSchedule"]
  }
  process = {
    vm_size   = "Standard_D2pds_v6" # ARM v6; scale-to-zero (min=0 so no vCPU at idle)
    min_count = 0
    max_count = 1
    priority  = "Regular" # was Spot; the 3-vCPU low-priority cap blocked concurrent scale-up
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
