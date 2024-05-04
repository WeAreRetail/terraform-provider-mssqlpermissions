# Queries

## Purpose

Contain the high-level functions for managing logins, users and roles.

## Useful SQL Queries

### View permissions for connected user

```sql
SELECT * FROM fn_my_permissions(NULL, 'DATABASE');
```

### View permissions role

```sql
SELECT DB_NAME() AS 'DBName'
      ,p.[name] AS 'PrincipalName'
      ,p.[type_desc] AS 'PrincipalType'
      ,p2.[name] AS 'GrantedBy'
      ,dbp.[permission_name]
      ,dbp.[state_desc]
      ,so.[Name] AS 'ObjectName'
      ,so.[type_desc] AS 'ObjectType'
  FROM [sys].[database_permissions] dbp LEFT JOIN [sys].[objects] so
    ON dbp.[major_id] = so.[object_id] LEFT JOIN [sys].[database_principals] p
    ON dbp.[grantee_principal_id] = p.[principal_id] LEFT JOIN [sys].[database_principals] p2
    ON dbp.[grantor_principal_id] = p2.[principal_id]

WHERE p.[name] = 'aware_db_state'
```

### View roles membership

```sql
SELECT
    DP1.name AS DatabaseRoleName
    ,ISNULL(DP2.name, 'No members') AS DatabaseUserName
    ,DP2.principal_id
    ,DP2.create_date
FROM sys.database_role_members AS DRM
RIGHT OUTER JOIN sys.database_principals AS DP1 ON DRM.role_principal_id = DP1.principal_id
LEFT OUTER JOIN sys.database_principals AS DP2  ON DRM.member_principal_id = DP2.principal_id
WHERE DP1.type = 'R'
ORDER BY DP1.name, ISNULL(DP2.name, 'No members');
```

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
