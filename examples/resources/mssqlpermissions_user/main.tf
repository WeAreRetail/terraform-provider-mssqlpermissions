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

output "user" {
  value = {
    name             = mssqlpermissions_user.user_resource.name
    external         = mssqlpermissions_user.user_resource.external
    principal_id     = mssqlpermissions_user.user_resource.principal_id
    default_schema   = mssqlpermissions_user.user_resource.default_schema
    default_language = mssqlpermissions_user.user_resource.default_language
    sid              = mssqlpermissions_user.user_resource.sid
  }
  sensitive = true # Because it references a resource with sensitive attributes
}
