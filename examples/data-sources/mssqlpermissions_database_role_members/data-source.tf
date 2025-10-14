terraform {
  required_version = ">= 1.0"

  required_providers {
    mssqlpermissions = {
      source  = "WeAreRetail/mssqlpermissions"
      version = ">= 0.0.5"
    }
  }
}

# Configure the provider
provider "mssqlpermissions" {
  server_fqdn   = "localhost"
  server_port   = 1433
  database_name = "testdb"

  sql_login = {
    username = "sa"
    password = "YourStrong@Passw0rd"
  }
}

# Read members of an existing database role
data "mssqlpermissions_database_role_members" "db_datareader_members" {
  name = "db_datareader"
}

# Output the list of members
output "datareader_members" {
  value       = data.mssqlpermissions_database_role_members.db_datareader_members.members
  description = "Members of the db_datareader role"
}

# Read members of a custom role
data "mssqlpermissions_database_role_members" "custom_role_members" {
  name = "custom_application_role"
}

output "custom_role_members" {
  value = data.mssqlpermissions_database_role_members.custom_role_members.members
}

# Use the data source in conjunction with a resource
resource "mssqlpermissions_database_role" "new_role" {
  name = "new_role"
}

# Read the members (initially empty for a new role)
data "mssqlpermissions_database_role_members" "new_role_members" {
  name       = mssqlpermissions_database_role.new_role.name
  depends_on = [mssqlpermissions_database_role.new_role]
}
