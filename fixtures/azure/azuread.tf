resource "azuread_directory_role" "directory_reader" {
  display_name = "Directory readers"
}

resource "random_string" "unique_admin" {
  length  = 4
  upper   = false
  special = false
}

locals {
  spn_prefix = join("", module.naming.resource_prefixes, ["-", random_string.unique_admin.result])
}

# An Azure AD service principal used as Azure Administrator for the Azure SQL Server resource.
# The SPN must have "Directory readers" role
resource "azuread_application" "sql_admin" {
  display_name = "terraform-provider-${local.spn_prefix}-sql-admin"
  web {
    homepage_url = "https://sql.example.com"
  }
}

resource "azuread_service_principal" "sql_admin" {
  client_id = azuread_application.sql_admin.client_id
}

resource "azuread_service_principal_password" "sql_admin" {
  service_principal_id = azuread_service_principal.sql_admin.id
}

resource "azuread_directory_role_assignment" "spn_directory_reader" {
  role_id             = azuread_directory_role.directory_reader.template_id
  principal_object_id = azuread_service_principal.sql_admin.id
}

resource "azuread_directory_role_assignment" "server_directory_reader" {
  role_id             = azuread_directory_role.directory_reader.template_id
  principal_object_id = azurerm_mssql_server.sql_server.identity.0.principal_id
}

resource "azuread_group" "admin_group" {
  display_name            = "terraform-provider-${local.spn_prefix}-sql-administrators"
  members                 = [data.azuread_group.admin_team.object_id, azuread_service_principal.sql_admin.object_id]
  security_enabled        = true
  prevent_duplicate_names = true
}


# An Azure AD service principal used to test grants
resource "azuread_application" "sql_user" {
  display_name = "terraform-provider-${local.spn_prefix}-sql-user"
  web {
    homepage_url = "https://sql.example.com"
  }
}

resource "azuread_service_principal" "sql_user" {
  client_id = azuread_application.sql_user.client_id
}

resource "azuread_service_principal_password" "sql_user" {
  service_principal_id = azuread_service_principal.sql_user.id
}
