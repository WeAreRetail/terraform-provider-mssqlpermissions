<!-- markdownlint-disable-file MD024 MD041 -->

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
