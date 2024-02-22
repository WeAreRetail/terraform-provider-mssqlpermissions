resource "mssqlpermissions_user" "user_resource" {
  config = {
    server_fqdn   = "mssql-fixture"
    server_port   = 1433
    database_name = "ApplicationDB"

    sql_login = {
      username = "sa"
      password = "P@ssw0rd"
    }
  }
  name      = "my-second-tf-user"
  password  = "P@ssw0rd!"
  contained = true
}
