resource "mssqlpermissions_login" "login_resource" {
  config = {
    server_fqdn   = "mssql-fixture"
    server_port   = 1433
    database_name = "master"

    sql_login = {
      username = "sa"
      password = "P@ssw0rd"
    }
  }

  name     = "my-first-tf-login"
  password = "P@ssw0rd!"
}
