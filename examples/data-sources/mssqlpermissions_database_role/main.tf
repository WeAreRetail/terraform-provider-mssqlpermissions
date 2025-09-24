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

output "example" {
  value = {
    name             = data.mssqlpermissions_database_role.example.name
    members          = data.mssqlpermissions_database_role.example.members
    principal_id     = data.mssqlpermissions_database_role.example.principal_id
    type             = data.mssqlpermissions_database_role.example.type
    type_description = data.mssqlpermissions_database_role.example.type_description
    owning_principal = data.mssqlpermissions_database_role.example.owning_principal
    is_fixed_role    = data.mssqlpermissions_database_role.example.is_fixed_role
  }
}
