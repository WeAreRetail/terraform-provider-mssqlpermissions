# Example provider configuration for Azure SQL Database using Service Principal authentication
provider "mssqlpermissions" {
  server_fqdn   = "myserver.database.windows.net"
  server_port   = 1433
  database_name = "ApplicationDB"

  spn_login = {
    client_id     = var.azure_client_id
    client_secret = var.azure_client_secret
    tenant_id     = var.azure_tenant_id
  }
}

variable "azure_client_id" {
  description = "Azure AD application client ID"
  type        = string
}

variable "azure_client_secret" {
  description = "Azure AD application client secret"
  type        = string
  sensitive   = true
}

variable "azure_tenant_id" {
  description = "Azure AD tenant ID"
  type        = string
}
