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

output "database_role" {
  value = {
    name             = mssqlpermissions_database_role.role.name
    members          = mssqlpermissions_database_role_members.role_members.members
    principal_id     = mssqlpermissions_database_role.role.principal_id
    type             = mssqlpermissions_database_role.role.type
    type_description = mssqlpermissions_database_role.role.type_description
    owning_principal = mssqlpermissions_database_role.role.owning_principal
    is_fixed_role    = mssqlpermissions_database_role.role.is_fixed_role
  }
}
