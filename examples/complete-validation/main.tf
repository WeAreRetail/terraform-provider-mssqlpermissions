terraform {
  required_version = ">= 1.0"

  required_providers {
    mssqlpermissions = {
      source  = "WeAreRetail/mssqlpermissions"
      version = ">= 0.0.5"
    }

    local = {
      source  = "hashicorp/local"
      version = ">= 2.1.0"
    }
  }
}

# Provider configuration - this will use environment variables
provider "mssqlpermissions" {
  server_fqdn   = var.server_fqdn
  server_port   = var.server_port
  database_name = var.database_name

  # Use SQL authentication for local testing
  sql_login = var.use_sql_auth ? {
    username = var.sql_username
    password = var.sql_password
  } : null

  # Use Azure AD authentication for cloud testing (Service Principal)
  spn_login = var.use_azure_auth ? {
    tenant_id     = var.azure_tenant_id
    client_id     = var.azure_client_id
    client_secret = var.azure_client_secret
  } : null
}

data "local_file" "permissions" {
  filename = "${path.module}/permissions.yml"
}

locals {
  permissions_deny_list_of_maps = [for permission in yamldecode(data.local_file.permissions.content).database.DENY : {
    permission_name = permission
    state           = "D"
  }]
}

# Create test users
resource "mssqlpermissions_user" "test_user_1" {
  name     = "test_user_validation_1"
  external = var.create_azure_ad_user
  password = var.create_azure_ad_user ? null : var.test_user_password
}

resource "mssqlpermissions_user" "test_user_2" {
  name     = "test_user_validation_2"
  external = var.create_azure_ad_user
  password = var.create_azure_ad_user ? null : var.test_user_password
}

# Create test database role
resource "mssqlpermissions_database_role" "test_role" {
  name = "test_role_validation"
}

# Create role members management
resource "mssqlpermissions_database_role_members" "test_role_members" {
  name = mssqlpermissions_database_role.test_role.name
  members = [
    mssqlpermissions_user.test_user_1.name,
    mssqlpermissions_user.test_user_2.name
  ]
}

# Grant permissions to role
resource "mssqlpermissions_permissions_to_role" "test_permissions" {
  role_name = mssqlpermissions_database_role.test_role.name
  permissions = [
    {
      permission_name = "SELECT"
    },
    {
      permission_name = "CONNECT"
    }
  ]
}

# Grant permissions to role
resource "mssqlpermissions_permissions_to_role" "test_permissions_deny" {
  role_name   = mssqlpermissions_database_role.test_role.name
  permissions = local.permissions_deny_list_of_maps
}

# Grant schema-level permissions to role
resource "mssqlpermissions_schema_permissions" "test_schema_permissions" {
  schema_name = "dbo"
  role_name   = mssqlpermissions_database_role.test_role.name
  permissions = [
    {
      permission_name = "SELECT"
      state           = "G"
    },
    {
      permission_name = "INSERT"
      state           = "G"
    }
  ]
}


# Data sources for validation
data "mssqlpermissions_user" "test_user_1_data" {
  name       = mssqlpermissions_user.test_user_1.name
  depends_on = [mssqlpermissions_user.test_user_1]
}

data "mssqlpermissions_database_role" "test_role_data" {
  name       = mssqlpermissions_database_role.test_role.name
  depends_on = [mssqlpermissions_database_role.test_role]
}
