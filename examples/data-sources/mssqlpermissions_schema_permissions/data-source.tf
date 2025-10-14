terraform {
  required_version = ">= 1.0"

  required_providers {
    mssqlpermissions = {
      source  = "WeAreRetail/mssqlpermissions"
      version = ">= 0.0.5"
    }
  }
}

# Read schema-level permissions for a role
data "mssqlpermissions_schema_permissions" "example" {
  role_name   = "app_reader"
  schema_name = "sales"
}

# Output the permissions
output "permissions" {
  description = "All permissions assigned to the role on the schema"
  value       = data.mssqlpermissions_schema_permissions.example.permissions
}

# Filter for GRANT permissions only
output "granted_permissions" {
  description = "Only GRANT permissions"
  value = [
    for p in data.mssqlpermissions_schema_permissions.example.permissions :
    p if p.state == "G"
  ]
}

# Filter for DENY permissions only
output "denied_permissions" {
  description = "Only DENY permissions"
  value = [
    for p in data.mssqlpermissions_schema_permissions.example.permissions :
    p if p.state == "D"
  ]
}

# List permission names only
output "permission_names" {
  description = "Simple list of permission names"
  value = [
    for p in data.mssqlpermissions_schema_permissions.example.permissions :
    "${p.state_desc} ${p.permission_name}"
  ]
}

# Check for specific permissions
output "has_select" {
  description = "Whether the role has SELECT permission"
  value = contains(
    [for p in data.mssqlpermissions_schema_permissions.example.permissions : p.permission_name],
    "SELECT"
  )
}

output "has_execute" {
  description = "Whether the role has EXECUTE permission"
  value = contains(
    [for p in data.mssqlpermissions_schema_permissions.example.permissions : p.permission_name],
    "EXECUTE"
  )
}
