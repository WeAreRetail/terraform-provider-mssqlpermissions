package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"terraform-provider-mssqlpermissions/internal/queries/model"
)

// Notes:
// MS SQL allows to grant permissions on specific objects only. These functions do not support that.
// The difficulty is that the object is stored in [database_permissions] and [server_permissions] by its ID in the major_id column.
// With additional sub-object ID, like the column, in the minor_id column.
// It means we need to query multiple views based on the object type to retrieve the full definition of the permission.
// Schemas would be in sys.schemas, tables in sys.tables, columns in sys.columns, etc.

// AssignPermissionToRole assigns the specified permission, grant or deny, to a role in the database.
// It takes a context, a database connection, a role, and a permission as parameters.
// Returns nil if the permission is successfully denied to the role, otherwise returns an error.
func (c *Connector) AssignPermissionToRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	var err error
	var stateVerb = "GRANT"

	if (permission.State != "G" && permission.State != "D" && permission.State != "") || (permission.StateDesc != "GRANT" && permission.StateDesc != "DENY" && permission.StateDesc != "") {
		return fmt.Errorf("invalid state value, must be 'G', 'D', 'GRANT', or 'DENY'")
	} else if permission.State == "G" || permission.StateDesc == "GRANT" {
		stateVerb = "GRANT"
	} else if permission.State == "D" || permission.StateDesc == "Deny" {
		stateVerb = "DENY"
	}

	// Check if the database connection is nil.
	if db == nil {
		return errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to deny permissions to a role.
	query := fmt.Sprintf("'%s %s TO ' + QUOTENAME(@roleName)", stateVerb, permission.Name)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	// Execute the query.
	_, err = db.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return fmt.Errorf("query execution error - cannot deny permissions to role: %v", err)
	}

	// Return nil error.
	return nil
}

// DenyPermissionToRole denies the specified permission to a role in the database.
// It takes a context, a database connection, a role, and a permission as parameters.
// Returns nil if the permission is successfully denied to the role, otherwise returns an error.
func (c *Connector) DenyPermissionToRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	permission.State = "D"

	return c.AssignPermissionToRole(ctx, db, role, permission)
}

// DenyPermissionsToRole denies the specified permissions to a role in the database.
// It takes a context, a database connection, a role, and a slice of permissions as parameters.
// Returns nil if the permissions are successfully denied to the role, otherwise returns an error.
func (c *Connector) DenyPermissionsToRole(ctx context.Context, db *sql.DB, role *model.Role, permissions []*model.Permission) error {

	for _, permission := range permissions {
		err := c.DenyPermissionToRole(ctx, db, role, permission)
		if err != nil {
			return err
		}
	}
	return nil
}

// GrantPermissionToRole grants the specified permission to a role in the database.
// It takes a context, a database connection, a role, and a permission as parameters.
// Returns nil if the permission is successfully granted to the role, otherwise returns an error.
func (c *Connector) GrantPermissionToRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	permission.State = "G"

	return c.AssignPermissionToRole(ctx, db, role, permission)
}

// GrantPermissionsToRole grants the specified permissions to a role in the database.
// It takes a context, a database connection, a role, and a slice of permissions as parameters.
// Returns nil if the permissions are successfully granted to the role, otherwise returns an error.
func (c *Connector) GrantPermissionsToRole(ctx context.Context, db *sql.DB, role *model.Role, permissions []*model.Permission) error {

	for _, permission := range permissions {
		err := c.GrantPermissionToRole(ctx, db, role, permission)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetServerPermissionsForRole retrieves the server permissions for a given role.
// It takes a context, a database connection, and a role as input parameters.
// It returns a slice of model.Permission and an error.
func (c *Connector) GetServerPermissionsForRole(ctx context.Context, db *sql.DB, role *model.Role) ([]model.Permission, error) {
	var err error
	var permissions []model.Permission

	if c.Database != "master" {
		return nil, errors.New("cannot get server permissions from non master database")
	}

	// Check if the database connection is nil.
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to get permissions for a role.
	query := `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
				FROM [sys].[server_permissions]
				WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[server_principals] WHERE name = @name)`

	// Execute the query and get a single row result.
	rows, err := db.QueryContext(ctx, query, sql.Named("name", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve permission for role: %v", err)
	}

	// Iterate through the resultset.
	for rows.Next() {
		var permission model.Permission

		// Scan the result into the Permission model.
		err = rows.Scan(
			&permission.Class,
			&permission.ClassDesc,
			&permission.MajorID,
			&permission.MinorID,
			&permission.GranteePrincipalID,
			&permission.GrantorPrincipalID,
			&permission.Type,
			&permission.Name,
			&permission.State,
			&permission.StateDesc)

		// Check for any error during the query execution.
		if err != nil {
			return nil, fmt.Errorf("scan error - cannot retrieve permissions for role: %v", err)
		}

		// Append the permission to the permissions slice.
		permissions = append(permissions, permission)
	}

	// Return the retrieved permissions.
	return permissions, nil
}

// GetServerPermissionForRole retrieves the permission details for a given server role.
// It takes a context.Context, *sql.DB, *model.Role, and a *model.Permission as input parameters.
// It returns a model.Permission and an error.
func (c *Connector) GetServerPermissionForRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) (*model.Permission, error) {
	var err error

	if c.Database != "master" {
		return nil, errors.New("cannot get server permissions from non master database")
	}

	// Check if the database connection is nil.
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to get permissions for a role.
	query := `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
				FROM [sys].[server_permissions]
				WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[server_principals] WHERE name = @name)
					AND [permission_name] = @permissionName`

	// Execute the query and get a single row result.
	row := db.QueryRowContext(
		ctx,
		query,
		sql.Named("name", role.Name),
		sql.Named("permissionName", permission.Name))

	// Check for any error during the query execution.
	if row.Err() != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve permission for role: %v", err)
	}

	// Scan the result into the Permission model.
	err = row.Scan(
		&permission.Class,
		&permission.ClassDesc,
		&permission.MajorID,
		&permission.MinorID,
		&permission.GranteePrincipalID,
		&permission.GrantorPrincipalID,
		&permission.Type,
		&permission.Name,
		&permission.State,
		&permission.StateDesc)

	// Check if the permission is not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("permissions not found")
	} else if err != nil {
		// Check for other scan errors.
		return nil, fmt.Errorf("scan error - cannot retrieve permission for role: %v", err)
	}

	// Return the retrieved permissions.
	return permission, nil
}

// GetDatabasePermissionsForRole retrieves the permissions for a given role from the database.
// It takes a context.Context, *sql.DB, and *model.Role as input parameters.
// It returns a slice of model.Permission and an error.
func (c *Connector) GetDatabasePermissionsForRole(ctx context.Context, db *sql.DB, role *model.Role) ([]model.Permission, error) {
	var err error
	var permissions []model.Permission

	// Check if the database connection is nil.
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to get permissions for a role.
	query := `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
				FROM [sys].[database_permissions]
				WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[database_principals] WHERE name = @name)`

	// Execute the query and get a single row result.
	rows, err := db.QueryContext(ctx, query, sql.Named("name", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve permissions for role: %v", err)
	}

	// Iterate through the resultset.
	for rows.Next() {
		var permission model.Permission

		// Scan the result into the Permission model.
		err = rows.Scan(
			&permission.Class,
			&permission.ClassDesc,
			&permission.MajorID,
			&permission.MinorID,
			&permission.GranteePrincipalID,
			&permission.GrantorPrincipalID,
			&permission.Type,
			&permission.Name,
			&permission.State,
			&permission.StateDesc)

		// Check for any error during the query execution.
		if err != nil {
			return nil, fmt.Errorf("scan error - cannot retrieve permissions for role: %v", err)
		}

		// Append the permission to the permissions slice.
		permissions = append(permissions, permission)
	}

	// Return the retrieved permissions.
	return permissions, nil
}

// GetServerPermissionForRole retrieves the permission details for a given server role.
// It takes a context.Context, *sql.DB, *model.Role, and a *model.Permission as input parameters.
// It returns a model.Permission and an error.
func (c *Connector) GetDatabasePermissionForRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) (*model.Permission, error) {
	var err error

	// Check if the database connection is nil.
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to get permissions for a role.
	query := `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
				FROM [sys].[database_permissions]
				WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[database_principals] WHERE name = @name)
					AND [permission_name] = @permissionName`

	// Execute the query and get a single row result.
	row := db.QueryRowContext(
		ctx,
		query,
		sql.Named("name", role.Name),
		sql.Named("permissionName", permission.Name))

	// Check for any error during the query execution.
	if row.Err() != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve permission for role: %v", err)
	}

	// Scan the result into the Permission model.
	err = row.Scan(
		&permission.Class,
		&permission.ClassDesc,
		&permission.MajorID,
		&permission.MinorID,
		&permission.GranteePrincipalID,
		&permission.GrantorPrincipalID,
		&permission.Type,
		&permission.Name,
		&permission.State,
		&permission.StateDesc)

	// Check if the permission is not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("permissions not found")
	} else if err != nil {
		// Check for other scan errors.
		return nil, fmt.Errorf("scan error - cannot retrieve permission for role: %v", err)
	}

	// Return the retrieved permissions.
	return permission, nil
}

// RevokePermissionFromRole revokes the specified database permissions from a role.
// It takes a context, a database connection, a role, and a permission as parameters.
// If there is an error during the query execution, it returns an error.
// Otherwise, it returns nil.
func (c *Connector) RevokePermissionFromRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	var err error

	// Check if the database connection is nil.
	if db == nil {
		return errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to revoke permissions from a role.
	query := fmt.Sprintf("'REVOKE %s TO ' + QUOTENAME(@roleName)", permission.Name)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	// Execute the query.
	_, err = db.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return fmt.Errorf("query execution error - cannot revoke permissions from role: %v", err)
	}

	// Return nil error.
	return nil
}
