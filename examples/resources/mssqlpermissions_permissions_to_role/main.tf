terraform {

  required_version = ">= 1.0"

  required_providers {
    mssqlpermissions = {
      source  = "WeAreRetail/mssqlpermissions"
      version = ">= 0.0.5"
    }
  }
}

provider "mssqlpermissions" {
  server_fqdn   = "mssql-fixture"
  server_port   = 1433
  database_name = "ApplicationDB"

  sql_login = {
    username = "sa"
    password = "P@ssw0rd"
  }
}

output "permissions" {
  value = {
    role_name   = mssqlpermissions_permissions_to_role.permissions.role_name
    permissions = mssqlpermissions_permissions_to_role.permissions.permissions
  }
}
