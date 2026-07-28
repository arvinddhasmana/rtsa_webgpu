# CLASSIFICATION: UNCLASSIFIED
# DNS/TLS module — optional A records in a shared DNS zone pointing at ingress IPs.
# Certificates are issued into Key Vault and mounted via CSI (see P2); this module
# only manages DNS records and is disabled by default.

variable "create_records" {
  type        = bool
  description = "Whether to create DNS A records."
  default     = false
}

variable "dns_zone_name" {
  type        = string
  description = "Name of the existing shared DNS zone."
  default     = ""
}

variable "dns_zone_resource_group" {
  type        = string
  description = "Resource group of the shared DNS zone."
  default     = ""
}

variable "records" {
  description = "A records to create: key = hostname label, value = { ip }."
  type = map(object({
    ip = string
  }))
  default = {}
}

variable "ttl" {
  type        = number
  description = "Record TTL (seconds)."
  default     = 300
}

resource "azurerm_dns_a_record" "this" {
  for_each = var.create_records ? var.records : {}

  name                = each.key
  zone_name           = var.dns_zone_name
  resource_group_name = var.dns_zone_resource_group
  ttl                 = var.ttl
  records             = [each.value.ip]
}

output "fqdns" {
  description = "Map of created record FQDNs."
  value       = { for k, r in azurerm_dns_a_record.this : k => r.fqdn }
}
