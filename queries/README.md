# Queries

## Purpose

Contain the high-level functions for managing logins, users and roles.

## Notes

### Contained Database

Azure SQL Databases are contained database. They can handle the user authentication at the database level rather than the server level.

By default, databases created on an on-premises server will not have the contained option enable.
Use the following snippet to enable it:

```sql
EXEC sp_configure 'CONTAINED DATABASE AUTHENTICATION', 1
GO
RECONFIGURE
GO
ALTER DATABASE ApplicationDB SET containment=PARTIAL
GO
```

Validate with:

```sql
EXEC sp_configure 'CONTAINED DATABASE AUTHENTICATION'
```
