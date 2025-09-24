resource "random_string" "admin_password" {
  length = 32

  lower   = true
  numeric = true
  special = true
  upper   = true

  min_lower   = 1
  min_numeric = 1
  min_special = 1
  min_upper   = 1
}

resource "azurecaf_name" "sql_server" {
  name          = ""
  resource_type = "azurerm_sql_server"
  prefixes      = module.naming.resource_prefixes
  random_length = 4
  suffixes      = []
  use_slug      = true
  clean_input   = true
  separator     = ""
}

resource "azurerm_mssql_server" "sql_server" {
  name                = azurecaf_name.sql_server.result
  resource_group_name = module.group.name
  location            = module.group.location

  version                      = "12.0"
  administrator_login          = "the-admin"
  administrator_login_password = random_string.admin_password.result

  azuread_administrator {
    login_username              = azuread_group.admin_group.display_name
    object_id                   = azuread_group.admin_group.object_id
    tenant_id                   = var.tenant_id
    azuread_authentication_only = false
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_mssql_firewall_rule" "sql_server_fw_rule" {
  for_each         = local.allowed_public_ips
  name             = "Allow IP ${each.key}"
  server_id        = azurerm_mssql_server.sql_server.id
  start_ip_address = each.key
  end_ip_address   = each.key
}

resource "azurerm_mssql_database" "db" {
  name      = "ApplicationDB"
  server_id = azurerm_mssql_server.sql_server.id
  sku_name  = "Basic"

  provisioner "local-exec" {
    command = "/usr/local/bin/sqlcmd -S '${azurerm_mssql_server.sql_server.fully_qualified_domain_name}' --authentication-method ActiveDirectoryServicePrincipal -U '${azuread_service_principal.sql_admin.client_id}' -P '${azuread_application_password.sql_admin.value}' -d ${self.name} -i ${path.root}/database_schema.sql"
  }
}


#
# Export the configuration to a file that will be sourced by Task.
#
resource "local_sensitive_file" "local_env" {
  filename             = "${path.root}/../../.azure.env"
  directory_permission = "0755"
  file_permission      = "0600"
  content              = <<-EOT
                        export TF_ACC=1
                        export AZURE_TEST=1
                        export AZURE_TENANT_ID='${var.tenant_id}'
                        export AZURE_MSSQL_DATABASE='${azurerm_mssql_database.db.name}'
                        export AZURE_MSSQL_ADMIN_CLIENT_ID='${azuread_service_principal.sql_admin.client_id}'
                        export AZURE_MSSQL_ADMIN_CLIENT_SECRET='${azuread_application_password.sql_admin.value}'
                        export AZURE_MSSQL_PASSWORD='${azurerm_mssql_server.sql_server.administrator_login_password}'
                        export AZURE_MSSQL_SERVER='${azurerm_mssql_server.sql_server.fully_qualified_domain_name}'
                        export AZURE_MSSQL_USERNAME='${azurerm_mssql_server.sql_server.administrator_login}'
                        export AZURE_MSSQL_USER_CLIENT_ID='${azuread_service_principal.sql_user.client_id}'
                        export AZURE_MSSQL_USER_CLIENT_SECRET='${azuread_application_password.sql_user.value}'
                        export AZURE_MSSQL_USER_CLIENT_DISPLAY_NAME='${azuread_service_principal.sql_user.display_name}'
                        # Configuration for fedauth which uses env vars via DefaultAzureCredential
                        export AZURE_TENANT_ID='${var.tenant_id}'
                        export AZURE_CLIENT_ID='${azuread_service_principal.sql_admin.client_id}'
                        export AZURE_CLIENT_SECRET='${azuread_application_password.sql_admin.value}'
                        # Configuration for provider
                        export ARM_CLIENT_ID='${azuread_service_principal.sql_admin.client_id}'
                        export ARM_CLIENT_SECRET='${azuread_application_password.sql_admin.value}'
                        export ARM_TENANT_ID='${var.tenant_id}'
                        export MSSQL_PORT='1433'
                         EOT
}
