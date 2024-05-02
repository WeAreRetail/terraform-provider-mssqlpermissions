data "mssqlpermissions_database_role" "example" {
  config = {
    server_fqdn   = "mssql-fixture"
    server_port   = 1433
    database_name = "ApplicationDB"

    sql_login = {
      username = "sa"
      password = "P@ssw0rd"
    }
  }

  name = "db_owner"
}
