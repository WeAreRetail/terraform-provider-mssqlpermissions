# Example provider configuration for Azure SQL Database using Managed Service Identity
provider "mssqlpermissions" {
  server_fqdn   = "myserver.database.windows.net"
  server_port   = 1433
  database_name = "ApplicationDB"

  msi_login = {
    user_identity = true
    user_id       = var.managed_identity_user_id # Optional, required if user_identity is true
  }
}

variable "managed_identity_user_id" {
  description = "The user identity for MSI authentication"
  type        = string
  default     = null
}
