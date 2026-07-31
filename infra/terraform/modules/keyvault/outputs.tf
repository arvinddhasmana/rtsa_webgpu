# CLASSIFICATION: UNCLASSIFIED
# Key Vault module outputs.

output "id" {
  description = "Resource ID of the Key Vault."
  value       = azurerm_key_vault.this.id
}

output "name" {
  description = "Name of the Key Vault."
  value       = azurerm_key_vault.this.name
}

output "uri" {
  description = "Vault URI."
  value       = azurerm_key_vault.this.vault_uri
}
