# CLASSIFICATION: UNCLASSIFIED
# ACR access module — grants the AKS kubelet identity pull rights on the shared ACR.

variable "acr_id" {
  type        = string
  description = "Resource ID of the shared Azure Container Registry."
}

variable "kubelet_object_id" {
  type        = string
  description = "Object (principal) ID of the AKS kubelet managed identity."
}

resource "azurerm_role_assignment" "acr_pull" {
  scope                            = var.acr_id
  role_definition_name             = "AcrPull"
  principal_id                     = var.kubelet_object_id
  principal_type                   = "ServicePrincipal"
  skip_service_principal_aad_check = true
}
