output "sql_database_name" {
  description = "The SQL database name"
  value       = azurerm_mssql_database.db.name
}

output "sql_server_fqdn" {
  description = "The fully qualified name of the SQL Server"
  value       = azurerm_mssql_server.sql_server.fully_qualified_domain_name
}
