resource "mssqlpermissions_database_role_members" "role" {
  name = "my-database-role"
  members = [
    "fixtureOne",
    "fixtureTwo",
  ]
}
