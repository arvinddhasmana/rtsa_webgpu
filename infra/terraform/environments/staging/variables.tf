# CLASSIFICATION: UNCLASSIFIED
# Input variables for the staging environment.

variable "project" {
  type        = string
  description = "Short project token."
  default     = "rtsa"
}

variable "environment" {
  type        = string
  description = "Environment name."
  default     = "staging"
}

variable "location" {
  type        = string
  description = "Azure region."
  default     = "canadacentral"
}

# ── Shared foundation (from P0 bootstrap outputs) ─────────────────────────────
variable "shared_resource_group" {
  type        = string
  description = "Name of the shared resource group (rg-rtsa-shared-<loc>)."
}

variable "shared_subscription_id" {
  type        = string
  description = "Subscription ID that hosts the shared foundation (ACR, tfstate, Log Analytics)."
}

variable "expected_subscription_id" {
  type        = string
  description = "Subscription ID this environment root is allowed to deploy into. Used as a safety guard against accidental cross-subscription applies."
}

variable "acr_name" {
  type        = string
  description = "Name of the shared Azure Container Registry."
}

variable "law_name" {
  type        = string
  description = "Name of the shared Log Analytics workspace."
}

# ── Networking ────────────────────────────────────────────────────────────────
variable "vnet_address_space" {
  type    = list(string)
  default = ["10.42.0.0/16"]
}

variable "subnet_aks_prefix" {
  type    = string
  default = "10.42.0.0/20"
}

variable "subnet_pe_prefix" {
  type    = string
  default = "10.42.16.0/24"
}

# ── AKS ───────────────────────────────────────────────────────────────────────
variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version (null = AKS default)."
  default     = null
}

variable "aks_sku_tier" {
  type    = string
  default = "Free"
}

variable "system_node_vm_size" {
  type    = string
  default = "Standard_D2pds_v6" # ARM v6; only Dpdsv6 family has quota in this subscription (canadacentral)
}

variable "system_node_min_count" {
  type    = number
  default = 1
}

variable "system_node_max_count" {
  type    = number
  default = 2
}

variable "system_node_zones" {
  type        = list(string)
  description = "Availability zones for the system node pool. Set to [] for regions that do not support AKS zones (e.g. canadacentral)."
  default     = []
}

variable "additional_node_pools" {
  type = map(object({
    vm_size         = string
    min_count       = number
    max_count       = number
    mode            = optional(string, "User")
    priority        = optional(string, "Regular")
    os_disk_size_gb = optional(number, 64)
    os_disk_type    = optional(string, "Managed")
    zones           = optional(list(string), ["1", "2", "3"])
    labels          = optional(map(string), {})
    taints          = optional(list(string), [])
  }))
  default = {}
}

variable "enable_istio" {
  type    = bool
  default = true
}

variable "istio_revisions" {
  type    = list(string)
  default = ["asm-1-29"] # asm-1-29: compatible with K8s 1.31-1.36; asm-1-23 was removed
}

variable "enable_keda" {
  type    = bool
  default = true
}

variable "admin_group_object_ids" {
  type    = list(string)
  default = []
}

# ── Key Vault ─────────────────────────────────────────────────────────────────
variable "key_vault_purge_protection" {
  type    = bool
  default = false
}

# ── Workload identities ───────────────────────────────────────────────────────
variable "workloads" {
  type = map(object({
    namespace              = string
    service_account        = string
    key_vault_secrets_user = optional(bool, true)
  }))
  default = {}
}

# ── Observability (managed Prometheus/Grafana — off in dev) ────────────────────
variable "enable_managed_prometheus" {
  type    = bool
  default = false
}

variable "enable_grafana" {
  type    = bool
  default = false
}

variable "grafana_admin_object_ids" {
  type    = list(string)
  default = []
}

# ── DNS (optional) ────────────────────────────────────────────────────────────
variable "create_dns_records" {
  type    = bool
  default = false
}

variable "dns_zone_name" {
  type    = string
  default = ""
}

variable "dns_zone_resource_group" {
  type    = string
  default = ""
}

variable "dns_records" {
  type = map(object({
    ip = string
  }))
  default = {}
}

# ── Lifecycle / tagging ───────────────────────────────────────────────────────
variable "ttl_hours" {
  type        = number
  description = "Intended lifetime hint for orphan sweeps / cost governance."
  default     = 8
}

variable "tags" {
  type    = map(string)
  default = {}
}
