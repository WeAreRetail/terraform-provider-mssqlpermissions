resource "mssqlpermissions_server_role" "role" {
  config = {
    server_fqdn   = "mssql-fixture"
    server_port   = 1433
    database_name = "master"

    sql_login = {
      username = "sa"
      password = "P@ssw0rd"
    }
  }

  name    = "my-second-tf-server-role"
  members = ["loginFixtureOne"]
}
