# `mssqlpermissions_permissions_to_role` Data Source

This data source reads database-level permissions assigned to a specific role in SQL Server / Azure SQL Database.

## Overview

The `mssqlpermissions_permissions_to_role` data source retrieves all database-level permissions (GRANT/DENY) assigned to a role. This is useful for:

- Auditing current permission assignments
- Validating that permissions are correctly configured
- Creating dependent resources based on existing permissions
- Documentation and compliance reporting

## Example Usage

```hcl
# Read permissions for a specific role
data "mssqlpermissions_permissions_to_role" "app_reader_perms" {
  role_name = "app_reader"
}

# Output the permissions
output "reader_permissions" {
  value = data.mssqlpermissions_permissions_to_role.app_reader_perms.permissions
}

# Use permissions in conditional logic
locals {
  has_select = contains(
    [for p in data.mssqlpermissions_permissions_to_role.app_reader_perms.permissions : p.permission_name],
    "SELECT"
  )
}
```

## Schema

### Required

- `role_name` (String) - The name of the database role to query.

### Read-Only

- `permissions` (List of Object) - List of permissions assigned to this role.
  - `permission_name` (String) - The SQL permission name (e.g., "SELECT", "INSERT", "EXECUTE").
  - `class` (String) - Permission class code.
  - `class_desc` (String) - Permission class description.
  - `major_id` (Number) - Permission major ID.
  - `minor_id` (Number) - Permission minor ID.
  - `grantee_principal_id` (Number) - ID of the principal receiving the permission.
  - `grantor_principal_id` (Number) - ID of the principal granting the permission.
  - `type` (String) - Permission type code.
  - `state` (String) - Permission state (`G` for GRANT, `D` for DENY).
  - `state_desc` (String) - Permission state description.

## Use Cases

### 1. Permission Auditing

```hcl
data "mssqlpermissions_permissions_to_role" "audit" {
  role_name = "db_datareader"
}

output "permission_audit" {
  value = {
    role  = data.mssqlpermissions_permissions_to_role.audit.role_name
    count = length(data.mssqlpermissions_permissions_to_role.audit.permissions)
    perms = [
      for p in data.mssqlpermissions_permissions_to_role.audit.permissions :
      "${p.state_desc} ${p.permission_name}"
    ]
  }
}
```

### 2. Conditional Resource Creation

```hcl
data "mssqlpermissions_permissions_to_role" "check" {
  role_name = "app_role"
}

# Only create additional permissions if SELECT is missing
resource "mssqlpermissions_permissions_to_role" "add_select" {
  count     = contains([for p in data.mssqlpermissions_permissions_to_role.check.permissions : p.permission_name], "SELECT") ? 0 : 1
  role_name = "app_role"

  permission {
    name  = "SELECT"
    state = "GRANT"
  }
}
```

### 3. Compliance Validation

```hcl
data "mssqlpermissions_permissions_to_role" "security_check" {
  role_name = "public"
}

# Check that public role has no dangerous permissions
check "public_role_restricted" {
  assert {
    condition = alltrue([
      for p in data.mssqlpermissions_permissions_to_role.security_check.permissions :
      !contains(["DELETE", "INSERT", "UPDATE"], p.permission_name)
    ])
    error_message = "Public role has dangerous permissions assigned"
  }
}
```

## Notes

- This data source queries the current state from the database
- Permissions include both GRANT and DENY states
- Database-level permissions only (use `mssqlpermissions_schema_permissions` for schema-level)
- Requires appropriate read permissions on system views
