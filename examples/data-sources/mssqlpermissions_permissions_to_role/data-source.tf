terraform {
  required_version = ">= 1.0"

  required_providers {
    mssqlpermissions = {
      source  = "WeAreRetail/mssqlpermissions"
      version = ">= 0.0.5"
    }
  }
}

# Read database-level permissions for a role
data "mssqlpermissions_permissions_to_role" "example" {
  role_name = "app_reader"
}

# Output the permissions
output "permissions" {
  description = "All database-level permissions assigned to the role"
  value       = data.mssqlpermissions_permissions_to_role.example.permissions
}

# Filter for specific permission types
output "grant_permissions" {
  description = "Only GRANT permissions"
  value = [
    for p in data.mssqlpermissions_permissions_to_role.example.permissions :
    p if p.state == "G"
  ]
}

output "deny_permissions" {
  description = "Only DENY permissions"
  value = [
    for p in data.mssqlpermissions_permissions_to_role.example.permissions :
    p if p.state == "D"
  ]
}

# List permission names only
output "permission_names" {
  description = "Simple list of permission names"
  value = [
    for p in data.mssqlpermissions_permissions_to_role.example.permissions :
    p.permission_name
  ]
}
