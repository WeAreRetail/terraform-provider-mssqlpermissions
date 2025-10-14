<!-- markdownlint-disable-file MD024 MD041 -->

## 1.1.0

NEW FEATURES:

* New resource: `mssqlpermissions_schema_permissions` - Manage schema-level permissions on SQL Server databases
* New data source: `mssqlpermissions_database_role_members` - Query role membership
* New data source: `mssqlpermissions_permissions_to_role` - Query database-level permissions for a role
* New data source: `mssqlpermissions_schema_permissions` - Query schema-level permissions for a role

ENHANCEMENTS:

* Complete resource/data source parity - all 5 resources now have matching data sources
* Improved documentation with comprehensive examples for all data sources

## 1.0.1

BUG FIXES:

* Fix various permissions issues
* Improve permission handling and validation

## 1.0.0

⚠️ BREAKING CHANGES:

* Remove configuration block from resources (use provider-level config instead)
* Remove support for server-only objects (login and server role) - focus on contained databases only
* Remove role membership management from role resource (use `mssqlpermissions_database_role_members` instead)
* Remove `mssqlpermissions_server_role_members` resource

ENHANCEMENTS:

* Add comprehensive unit tests alongside acceptance tests
* Add queries to manage schema permissions (backend support)
* Update Go to 1.24
* Update all dependencies
* Update dev container
* Improve code organization and separation of concerns

## 0.1.5

BUG FIXES:

* Keep state order for roles' members.

## 0.1.4

BUG FIXES:

* Handle the "user not found" gracefully.

## 0.1.3

BUG FIXES:

* SID uses string representation.

## 0.1.2

BUG FIXES:

* Create DENY permissions created as GRANT

## 0.1.1

BUG FIXES:

* Ignore `dbo` in role membership.

## 0.1.0

NEW FEATURE:

* Add a "DENY" permission option.
* New resource: `mssqlpermissions_database_role_members`
* New resource: `mssqlpermissions_server_role_members`
* New data source: `mssqlpermissions_database_role`

## 0.0.7

ENHANCEMENT:

* Improved documentation

## 0.0.6

BUG FIXES:

* Permissions updates not applied.
