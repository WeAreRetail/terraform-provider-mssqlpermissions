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
}

output "example" {
  value = data.mssqlpermissions_database_role.example
}
