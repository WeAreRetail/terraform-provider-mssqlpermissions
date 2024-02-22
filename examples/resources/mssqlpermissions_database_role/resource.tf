resource "mssqlpermissions_database_role" "role" {
  config = {
    server_fqdn   = "mssql-fixture"
    server_port   = 1433
    database_name = "ApplicationDB"

    sql_login = {
      username = "sa"
      password = "P@ssw0rd"
    }
  }

  name = "my-database-role"
  members = [
    "fixtureOne",
    "fixtureTwo",
  ]
}
