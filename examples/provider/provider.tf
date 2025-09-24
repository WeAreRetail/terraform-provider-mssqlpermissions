# Example provider configuration for local testing with SQL authentication
provider "mssqlpermissions" {
  server_fqdn   = "mssql-fixture"
  server_port   = 1433
  database_name = "ApplicationDB"

  sql_login = {
    username = "sa"
    password = "P@ssw0rd"
  }
}

# For production environments, consider using one of these authentication methods:

# Azure Service Principal authentication (recommended for automation)
# provider "mssqlpermissions" {
#   server_fqdn   = "myserver.database.windows.net"
#   server_port   = 1433
#   database_name = "ApplicationDB"
#
#   spn_login = {
#     client_id     = var.azure_client_id
#     client_secret = var.azure_client_secret
#     tenant_id     = var.azure_tenant_id
#   }
# }

# Managed Service Identity authentication (recommended for Azure resources)
# provider "mssqlpermissions" {
#   server_fqdn   = "myserver.database.windows.net"
#   server_port   = 1433
#   database_name = "ApplicationDB"
#
#   msi_login = {
#     user_identity = true
#     user_id       = var.managed_identity_user_id
#   }
# }
