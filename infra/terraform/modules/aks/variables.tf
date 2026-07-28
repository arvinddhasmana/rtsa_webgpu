# CLASSIFICATION: UNCLASSIFIED
# AKS module — production-baseline cluster with bulkhead node pools, managed Istio,
# KEDA, Workload Identity (OIDC issuer), Azure CNI Overlay + Cilium, and Key Vault CSI.

variable "name" {
  type        = string
  description = "AKS cluster name."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "resource_group_name" {
  type        = string
  description = "Resource group for the cluster."
}

variable "dns_prefix" {
  type        = string
  description = "DNS prefix for the cluster API server."
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version. null = AKS default."
  default     = null
}

variable "sku_tier" {
  type        = string
  description = "Control-plane SKU tier (Free for non-prod, Standard for prod SLA)."
  default     = "Free"

  validation {
    condition     = contains(["Free", "Standard", "Premium"], var.sku_tier)
    error_message = "sku_tier must be Free, Standard or Premium."
  }
}

variable "aks_subnet_id" {
  type        = string
  description = "Subnet ID for cluster nodes."
}

variable "log_analytics_workspace_id" {
  type        = string
  description = "Log Analytics workspace ID for Container Insights."
}

variable "system_node_vm_size" {
  type        = string
  description = "VM size for the system node pool."
  default     = "Standard_D2s_v5"
}

variable "system_node_min_count" {
  type        = number
  description = "Minimum system node count."
  default     = 1
}

variable "system_node_max_count" {
  type        = number
  description = "Maximum system node count."
  default     = 3
}

variable "system_node_zones" {
  type        = list(string)
  description = "Availability zones for the system node pool."
  default     = ["1", "2", "3"]
}

variable "additional_node_pools" {
  description = "User node pools keyed by name (max 12 lowercase alphanumeric chars)."
  type = map(object({
    vm_size         = string
    min_count       = number
    max_count       = number
    mode            = optional(string, "User")
    priority        = optional(string, "Regular") # Regular | Spot
    os_disk_size_gb = optional(number, 64)
    os_disk_type    = optional(string, "Managed") # Managed | Ephemeral
    zones           = optional(list(string), ["1", "2", "3"])
    labels          = optional(map(string), {})
    taints          = optional(list(string), [])
  }))
  default = {}
}

variable "pod_cidr" {
  type        = string
  description = "Pod CIDR for Azure CNI Overlay."
  default     = "10.244.0.0/16"
}

variable "service_cidr" {
  type        = string
  description = "Service CIDR."
  default     = "10.0.0.0/16"
}

variable "dns_service_ip" {
  type        = string
  description = "Cluster DNS service IP (inside service_cidr)."
  default     = "10.0.0.10"
}

variable "enable_istio" {
  type        = bool
  description = "Enable the managed Istio service mesh add-on."
  default     = true
}

variable "istio_revisions" {
  type        = list(string)
  description = "Istio add-on revision(s), e.g. [\"asm-1-23\"]."
  default     = ["asm-1-23"]
}

variable "istio_external_ingress_enabled" {
  type        = bool
  description = "Deploy the Istio external ingress gateway (cold-path edge)."
  default     = true
}

variable "enable_keda" {
  type        = bool
  description = "Enable the KEDA workload autoscaler add-on."
  default     = true
}

variable "admin_group_object_ids" {
  type        = list(string)
  description = "Entra group object IDs granted cluster-admin via Azure RBAC."
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Tags."
  default     = {}
}
