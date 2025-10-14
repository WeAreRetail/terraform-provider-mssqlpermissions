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

# Create a database role
resource "mssqlpermissions_database_role" "example_role" {
  name = "schema_permissions_example_role"
}

# Grant SELECT permission on dbo schema to the role
resource "mssqlpermissions_schema_permissions" "dbo_select" {
  schema_name = "dbo"
  role_name   = mssqlpermissions_database_role.example_role.name

  permissions = [
    {
      permission_name = "SELECT"
      state           = "G"
    }
  ]
}

# Grant multiple permissions on a custom schema
resource "mssqlpermissions_schema_permissions" "sales_permissions" {
  schema_name = "sales"
  role_name   = mssqlpermissions_database_role.example_role.name

  permissions = [
    {
      permission_name = "SELECT"
      state           = "G"
    },
    {
      permission_name = "INSERT"
      state           = "G"
    },
    {
      permission_name = "UPDATE"
      state           = "G"
    }
  ]
}

# Deny DELETE permission on a sensitive schema
resource "mssqlpermissions_schema_permissions" "sensitive_deny" {
  schema_name = "audit"
  role_name   = mssqlpermissions_database_role.example_role.name

  permissions = [
    {
      permission_name = "SELECT"
      state           = "G" # Allow reading audit data
    },
    {
      permission_name = "DELETE"
      state           = "D" # Explicitly deny deletion
    },
    {
      permission_name = "UPDATE"
      state           = "D" # Explicitly deny updates
    }
  ]
}
