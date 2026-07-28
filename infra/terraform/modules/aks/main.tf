# CLASSIFICATION: UNCLASSIFIED
# AKS module — cluster + bulkhead user node pools.

resource "azurerm_kubernetes_cluster" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  dns_prefix          = var.dns_prefix
  kubernetes_version  = var.kubernetes_version
  sku_tier            = var.sku_tier

  # Identity & security posture.
  oidc_issuer_enabled       = true # required for Workload Identity federation
  workload_identity_enabled = true
  azure_policy_enabled      = true

  default_node_pool {
    name                         = "system"
    vm_size                      = var.system_node_vm_size
    vnet_subnet_id               = var.aks_subnet_id
    orchestrator_version         = var.kubernetes_version
    os_sku                       = "Ubuntu"
    zones                        = var.system_node_zones
    auto_scaling_enabled         = true
    min_count                    = var.system_node_min_count
    max_count                    = var.system_node_max_count
    only_critical_addons_enabled = true # keep app workloads off the system pool (bulkhead)
    temporary_name_for_rotation  = "systmp"

    upgrade_settings {
      max_surge = "33%"
    }
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    network_policy      = "cilium"
    network_data_plane  = "cilium" # Azure CNI powered by Cilium (eBPF)
    load_balancer_sku   = "standard"
    pod_cidr            = var.pod_cidr
    service_cidr        = var.service_cidr
    dns_service_ip      = var.dns_service_ip
  }

  # Secrets Store CSI driver (Key Vault) with auto-rotation.
  key_vault_secrets_provider {
    secret_rotation_enabled = true
  }

  # KEDA (+ leave VPA off for the baseline).
  workload_autoscaler_profile {
    keda_enabled                    = var.enable_keda
    vertical_pod_autoscaler_enabled = false
  }

  # Container Insights → shared Log Analytics.
  oms_agent {
    log_analytics_workspace_id = var.log_analytics_workspace_id
  }

  # Azure RBAC for Kubernetes authorization.
  azure_active_directory_role_based_access_control {
    azure_rbac_enabled     = true
    admin_group_object_ids = var.admin_group_object_ids
  }

  # Managed Istio service-mesh add-on (resiliency + mTLS plane).
  dynamic "service_mesh_profile" {
    for_each = var.enable_istio ? [1] : []
    content {
      mode                             = "Istio"
      revisions                        = var.istio_revisions
      external_ingress_gateway_enabled = var.istio_external_ingress_enabled
    }
  }

  tags = var.tags

  lifecycle {
    ignore_changes = [
      kubernetes_version,              # patch upgrades handled out-of-band
      default_node_pool[0].node_count, # managed by the cluster autoscaler
    ]
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "pools" {
  for_each = var.additional_node_pools

  name                  = each.key
  kubernetes_cluster_id = azurerm_kubernetes_cluster.this.id
  vm_size               = each.value.vm_size
  mode                  = each.value.mode
  vnet_subnet_id        = var.aks_subnet_id
  zones                 = each.value.zones
  os_disk_size_gb       = each.value.os_disk_size_gb
  os_disk_type          = each.value.os_disk_type

  auto_scaling_enabled = true
  min_count            = each.value.min_count
  max_count            = each.value.max_count

  priority        = each.value.priority
  eviction_policy = each.value.priority == "Spot" ? "Delete" : null
  spot_max_price  = each.value.priority == "Spot" ? -1 : null

  node_labels = each.value.labels
  node_taints = each.value.taints

  tags = var.tags

  lifecycle {
    ignore_changes = [node_count]
  }
}
