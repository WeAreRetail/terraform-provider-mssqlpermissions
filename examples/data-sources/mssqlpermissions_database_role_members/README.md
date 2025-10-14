# Database Role Members Data Source

This data source reads the current members of a database role.

## Basic Usage

```terraform
data "mssqlpermissions_database_role_members" "example" {
  name = "db_datareader"
}

output "members" {
  value = data.mssqlpermissions_database_role_members.example.members
}
```

## Use Cases

### 1. Audit Role Membership

Check who has access to specific roles:

```terraform
data "mssqlpermissions_database_role_members" "admin_members" {
  name = "db_owner"
}

output "database_admins" {
  value       = data.mssqlpermissions_database_role_members.admin_members.members
  description = "List of users with db_owner access"
}
```

### 2. Validate Role Setup

Verify a role has the expected members after creation:

```terraform
resource "mssqlpermissions_database_role" "app_role" {
  name = "application_role"
}

resource "mssqlpermissions_database_role_members" "app_members" {
  name    = mssqlpermissions_database_role.app_role.name
  members = ["app_user_1", "app_user_2"]
}

data "mssqlpermissions_database_role_members" "verify_members" {
  name       = mssqlpermissions_database_role.app_role.name
  depends_on = [mssqlpermissions_database_role_members.app_members]
}

# Verify the members were added
output "actual_members" {
  value = data.mssqlpermissions_database_role_members.verify_members.members
}
```

### 3. Conditional Logic Based on Membership

Use membership information for conditional logic:

```terraform
locals {
  has_members = length(data.mssqlpermissions_database_role_members.example.members) > 0
  member_count = length(data.mssqlpermissions_database_role_members.example.members)
}

output "role_status" {
  value = local.has_members ? "Role has ${local.member_count} member(s)" : "Role is empty"
}
```

### 4. Cross-Role Verification

Compare membership across multiple roles:

```terraform
data "mssqlpermissions_database_role_members" "readers" {
  name = "db_datareader"
}

data "mssqlpermissions_database_role_members" "writers" {
  name = "db_datawriter"
}

output "read_only_users" {
  value = setsubtract(
    toset(data.mssqlpermissions_database_role_members.readers.members),
    toset(data.mssqlpermissions_database_role_members.writers.members)
  )
  description = "Users with read access but not write access"
}
```

## Attributes

### Required

- `name` (String) - The name of the database role to query

### Computed

- `members` (List of String) - List of user names that are members of this role

## Notes

- The data source queries the current state from the database
- Returns an empty list if the role has no members
- Fixed database roles (like `db_datareader`) can be queried
- The `dbo` user is excluded from the members list (as per provider design)
- Use `depends_on` when querying a role that was just created in the same configuration

## Common Fixed Database Roles

You can query these built-in SQL Server roles:

- `db_owner` - Database owners
- `db_datareader` - Can read all data
- `db_datawriter` - Can write all data
- `db_ddladmin` - Can run DDL commands
- `db_accessadmin` - Can manage role membership
- `db_securityadmin` - Can manage permissions
- `db_backupoperator` - Can backup the database
- `db_denydatareader` - Cannot read data
- `db_denydatawriter` - Cannot write data
