# CLASSIFICATION: UNCLASSIFIED
# Network module — VNet, subnets, and default-deny NSGs for an RTSA environment.

variable "name_prefix" {
  type        = string
  description = "Resource name prefix, e.g. rtsa-dev."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "resource_group_name" {
  type        = string
  description = "Resource group to create network resources in."
}

variable "vnet_address_space" {
  type        = list(string)
  description = "Address space for the VNet."
  default     = ["10.42.0.0/16"]
}

variable "subnet_aks_prefix" {
  type        = string
  description = "CIDR for the AKS node subnet (Azure CNI Overlay: nodes only)."
  default     = "10.42.0.0/20"
}

variable "subnet_pe_prefix" {
  type        = string
  description = "CIDR for the private-endpoints subnet."
  default     = "10.42.16.0/24"
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to all network resources."
  default     = {}
}
