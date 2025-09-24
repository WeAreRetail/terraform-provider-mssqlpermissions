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
    name             = data.mssqlpermissions_user.example.name
    external         = data.mssqlpermissions_user.example.external
    principal_id     = data.mssqlpermissions_user.example.principal_id
    default_schema   = data.mssqlpermissions_user.example.default_schema
    default_language = data.mssqlpermissions_user.example.default_language
    sid              = data.mssqlpermissions_user.example.sid
  }
}
