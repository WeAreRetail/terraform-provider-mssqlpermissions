resource "mssqlpermissions_database_role" "role" {
  name = "another-database-role"
}

resource "mssqlpermissions_permissions_to_role" "permissions" {
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
