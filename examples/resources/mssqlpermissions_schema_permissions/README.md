# Schema Permissions Resource Example

This example demonstrates how to grant permissions on a specific database schema to a role.

## Basic Usage

```terraform
# Grant SELECT permission on the dbo schema to a role
resource "mssqlpermissions_schema_permissions" "example" {
  schema_name = "dbo"
  role_name   = "db_reader"

  permissions = [
    {
      permission_name = "SELECT"
      state           = "G" # GRANT
    }
  ]
}
```

## Multiple Permissions

```terraform
# Grant multiple permissions on a schema
resource "mssqlpermissions_schema_permissions" "multiple" {
  schema_name = "sales"
  role_name   = "sales_role"

  permissions = [
    {
      permission_name = "SELECT"
      state           = "G"
    },
    {
      permission_name = "INSERT"
      state           = "G"
    },
    {
      permission_name = "UPDATE"
      state           = "G"
    },
    {
      permission_name = "DELETE"
      state           = "G"
    }
  ]
}
```

## Deny Permissions

```terraform
# Deny specific permissions on a schema
resource "mssqlpermissions_schema_permissions" "deny_example" {
  schema_name = "restricted"
  role_name   = "limited_role"

  permissions = [
    {
      permission_name = "DELETE"
      state           = "D" # DENY
    },
    {
      permission_name = "ALTER"
      state           = "D"
    }
  ]
}
```

## Mixed Grant and Deny

```terraform
# Mix GRANT and DENY permissions
resource "mssqlpermissions_schema_permissions" "mixed" {
  schema_name = "mixed_schema"
  role_name   = "mixed_role"

  permissions = [
    {
      permission_name = "SELECT"
      state           = "G" # Grant SELECT
    },
    {
      permission_name = "INSERT"
      state           = "G" # Grant INSERT
    },
    {
      permission_name = "DELETE"
      state           = "D" # Deny DELETE
    }
  ]
}
```

## Common Schema Permissions

Available permission names for schemas include:

- `SELECT` - Query data in schema objects
- `INSERT` - Insert data into schema tables
- `UPDATE` - Update data in schema tables
- `DELETE` - Delete data from schema tables
- `EXECUTE` - Execute stored procedures/functions in schema
- `ALTER` - Alter schema objects
- `CONTROL` - Full control over the schema
- `REFERENCES` - Create foreign key references

## Notes

- The `state` field accepts:
  - `"G"` for GRANT (default if not specified)
  - `"D"` for DENY
- Changing `schema_name` or `role_name` requires resource replacement
- Permissions are managed as a set; removing a permission from the list will revoke it
- The role must exist before permissions can be assigned
