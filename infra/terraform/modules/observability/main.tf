# CLASSIFICATION: UNCLASSIFIED
# Observability module — optional Azure Monitor Workspace (managed Prometheus) +
# Managed Grafana. Disabled by default; enable for staging/prod. Container Insights
# is wired directly on the cluster (oms_agent), so dev needs nothing here.

variable "name_prefix" {
  type        = string
  description = "Resource name prefix, e.g. rtsa-prod."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "resource_group_name" {
  type        = string
  description = "Resource group name."
}

variable "enable_managed_prometheus" {
  type        = bool
  description = "Create an Azure Monitor Workspace for managed Prometheus."
  default     = false
}

variable "enable_grafana" {
  type        = bool
  description = "Create an Azure Managed Grafana instance."
  default     = false
}

variable "grafana_admin_object_ids" {
  type        = list(string)
  description = "Entra object IDs granted Grafana Admin."
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Tags."
  default     = {}
}

resource "azurerm_monitor_workspace" "prometheus" {
  count               = var.enable_managed_prometheus ? 1 : 0
  name                = "amw-${var.name_prefix}"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = var.tags
}

resource "azurerm_dashboard_grafana" "this" {
  count                             = var.enable_grafana ? 1 : 0
  name                              = "graf-${var.name_prefix}"
  location                          = var.location
  resource_group_name               = var.resource_group_name
  grafana_major_version             = "11"
  sku                               = "Standard"
  api_key_enabled                   = false
  deterministic_outbound_ip_enabled = false
  public_network_access_enabled     = true

  identity {
    type = "SystemAssigned"
  }

  dynamic "azure_monitor_workspace_integrations" {
    for_each = var.enable_managed_prometheus ? [1] : []
    content {
      resource_id = azurerm_monitor_workspace.prometheus[0].id
    }
  }

  tags = var.tags
}

resource "azurerm_role_assignment" "grafana_admin" {
  for_each = var.enable_grafana ? toset(var.grafana_admin_object_ids) : []

  scope                = azurerm_dashboard_grafana.this[0].id
  role_definition_name = "Grafana Admin"
  principal_id         = each.value
}

output "monitor_workspace_id" {
  description = "Azure Monitor Workspace ID (empty if disabled)."
  value       = try(azurerm_monitor_workspace.prometheus[0].id, "")
}

output "grafana_endpoint" {
  description = "Managed Grafana endpoint (empty if disabled)."
  value       = try(azurerm_dashboard_grafana.this[0].endpoint, "")
}
