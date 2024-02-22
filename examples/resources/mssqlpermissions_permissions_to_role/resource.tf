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

  name = "another-database-role"
}

resource "mssqlpermissions_permissions_to_role" "permissions" {
  config = {
    server_fqdn   = "mssql-fixture"
    server_port   = 1433
    database_name = "ApplicationDB"

    sql_login = {
      username = "sa"
      password = "P@ssw0rd"
    }
  }

  role_name = mssqlpermissions_database_role.role.name
  permissions = [
    {
      permission_name = "SELECT"
    },
    {
      permission_name = "INSERT"
    },
    {
      permission_name = "UPDATE"
    },
    {
      permission_name = "DELETE"
    },
    {
      permission_name = "EXECUTE"
    }
  ]
}
