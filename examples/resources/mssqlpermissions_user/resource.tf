resource "mssqlpermissions_user" "user_resource" {
  name     = "my-second-tf-user"
  password = "P@ssw0rd!"
  external = false
}
