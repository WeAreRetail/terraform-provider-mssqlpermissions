# `mssqlpermissions_schema_permissions` Data Source

This data source reads schema-level permissions assigned to a specific role on a specific schema in SQL Server / Azure SQL Database.

## Overview

The `mssqlpermissions_schema_permissions` data source retrieves all permissions (GRANT/DENY) assigned to a role for a particular schema. This is useful for:

- Auditing schema-level permission assignments
- Validating that schema permissions are correctly configured
- Creating dependent resources based on existing schema permissions
- Security compliance and documentation

## Example Usage

```hcl
# Read permissions for a role on a specific schema
data "mssqlpermissions_schema_permissions" "app_perms" {
  role_name   = "app_reader"
  schema_name = "sales"
}

# Output the permissions
output "schema_permissions" {
  value = data.mssqlpermissions_schema_permissions.app_perms.permissions
}

# Check if role has specific permission on schema
locals {
  has_select = contains(
    [for p in data.mssqlpermissions_schema_permissions.app_perms.permissions : p.permission_name],
    "SELECT"
  )
}
```

## Schema

### Required

- `role_name` (String) - The name of the database role to query.
- `schema_name` (String) - The name of the schema to query permissions for.

### Read-Only

- `permissions` (List of Object) - List of permissions assigned to this role on the schema.
  - `permission_name` (String) - The SQL permission name (e.g., "SELECT", "INSERT", "EXECUTE", "ALTER").
  - `class` (String) - Permission class code.
  - `class_desc` (String) - Permission class description (typically "SCHEMA").
  - `major_id` (Number) - Schema object ID.
  - `minor_id` (Number) - Permission minor ID.
  - `grantee_principal_id` (Number) - ID of the principal receiving the permission.
  - `grantor_principal_id` (Number) - ID of the principal granting the permission.
  - `type` (String) - Permission type code.
  - `state` (String) - Permission state (`G` for GRANT, `D` for DENY).
  - `state_desc` (String) - Permission state description.

## Use Cases

### 1. Schema Permission Auditing

```hcl
data "mssqlpermissions_schema_permissions" "audit" {
  role_name   = "app_role"
  schema_name = "dbo"
}

output "schema_audit" {
  value = {
    role   = data.mssqlpermissions_schema_permissions.audit.role_name
    schema = data.mssqlpermissions_schema_permissions.audit.schema_name
    count  = length(data.mssqlpermissions_schema_permissions.audit.permissions)
    perms = [
      for p in data.mssqlpermissions_schema_permissions.audit.permissions :
      "${p.state_desc} ${p.permission_name}"
    ]
  }
}
```

### 2. Multi-Schema Permission Check

```hcl
variable "schemas" {
  default = ["sales", "hr", "finance"]
}

data "mssqlpermissions_schema_permissions" "all_schemas" {
  for_each = toset(var.schemas)

  role_name   = "app_reader"
  schema_name = each.value
}

output "schema_permission_summary" {
  value = {
    for schema_name, data in data.mssqlpermissions_schema_permissions.all_schemas :
    schema_name => length(data.permissions)
  }
}
```

### 3. Conditional Resource Creation

```hcl
data "mssqlpermissions_schema_permissions" "check" {
  role_name   = "app_role"
  schema_name = "sales"
}

# Only grant SELECT if not already granted
resource "mssqlpermissions_schema_permissions" "ensure_select" {
  count = contains(
    [for p in data.mssqlpermissions_schema_permissions.check.permissions : p.permission_name],
    "SELECT"
  ) ? 0 : 1

  role_name   = "app_role"
  schema_name = "sales"

  permission {
    name  = "SELECT"
    state = "GRANT"
  }
}
```

### 4. Security Compliance

```hcl
data "mssqlpermissions_schema_permissions" "public_check" {
  role_name   = "public"
  schema_name = "dbo"
}

# Validate that public role has no dangerous permissions
check "public_schema_restricted" {
  assert {
    condition = alltrue([
      for p in data.mssqlpermissions_schema_permissions.public_check.permissions :
      p.state != "G" || !contains(["ALTER", "CONTROL", "TAKE OWNERSHIP"], p.permission_name)
    ])
    error_message = "Public role has dangerous schema permissions"
  }
}
```

## Notes

- This data source queries the current state from the database
- Permissions are scoped to a specific schema (use `mssqlpermissions_permissions_to_role` for database-level)
- Both GRANT and DENY permissions are returned
- Requires appropriate read permissions on system views (`sys.database_permissions`, `sys.schemas`)
- Schema must exist in the database
