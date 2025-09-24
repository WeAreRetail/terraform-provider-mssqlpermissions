resource "mssqlpermissions_database_role" "role" {
  name = "my-database-role"
}

# Manage role members separately using the dedicated resource
resource "mssqlpermissions_database_role_members" "role_members" {
  name = mssqlpermissions_database_role.role.name
  members = [
    "fixtureOne",
    "fixtureTwo",
  ]
}
