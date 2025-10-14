package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"terraform-provider-mssqlpermissions/internal/queries/model"
)

// #region Constants and Variables
// ============================================================================
// CONSTANTS AND VARIABLES
// ============================================================================

// SQL Query Constants
const (
	// Server permission queries
	QueryServerPermissionsForRole = `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
		FROM [sys].[server_permissions]
		WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[server_principals] WHERE name = @name)`

	QueryServerPermissionForRole = `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
		FROM [sys].[server_permissions]
		WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[server_principals] WHERE name = @name)
			AND [permission_name] = @permissionName`

	// Database permission queries
	QueryDatabasePermissionsForRole = `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
		FROM [sys].[database_permissions]
		WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[database_principals] WHERE name = @name)`

	QueryDatabasePermissionForRole = `SELECT [class], [class_desc], [major_id], [minor_id], [grantee_principal_id], [grantor_principal_id], [type], [permission_name], [state], [state_desc]
		FROM [sys].[database_permissions]
		WHERE grantee_principal_id = (SELECT principal_id FROM [sys].[database_principals] WHERE name = @name)
			AND [permission_name] = @permissionName`

	// Schema permission queries
	QuerySchemaPermissionsForRole = `SELECT dp.[class], dp.[class_desc], dp.[major_id], dp.[minor_id], dp.[grantee_principal_id], dp.[grantor_principal_id], dp.[type], dp.[permission_name], dp.[state], dp.[state_desc]
		FROM [sys].[database_permissions] dp
		INNER JOIN [sys].[schemas] s ON dp.[major_id] = s.[schema_id]
		WHERE dp.[grantee_principal_id] = (SELECT principal_id FROM [sys].[database_principals] WHERE name = @roleName)
			AND s.[name] = @schemaName
			AND dp.[class] = 3`

	QuerySchemaPermissionForRole = `SELECT dp.[class], dp.[class_desc], dp.[major_id], dp.[minor_id], dp.[grantee_principal_id], dp.[grantor_principal_id], dp.[type], dp.[permission_name], dp.[state], dp.[state_desc]
		FROM [sys].[database_permissions] dp
		INNER JOIN [sys].[schemas] s ON dp.[major_id] = s.[schema_id]
		WHERE dp.[grantee_principal_id] = (SELECT principal_id FROM [sys].[database_principals] WHERE name = @roleName)
			AND s.[name] = @schemaName
			AND dp.[permission_name] = @permissionName
			AND dp.[class] = 3`

	// SQL identifier validation
	MaxSQLIdentifierLength = 128
)

// Regular expression for valid SQL identifiers
var sqlIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Regular expression for valid SQL permission names (allows spaces and more flexible casing)
// SQL Server permission names can be uppercase, may contain spaces, numbers, and underscores
// They typically start with a letter but can have various formats
var permissionNameRegex = regexp.MustCompile(`^[A-Z][A-Z ]*[A-Z]$`)

// Notes:
// MS SQL allows to grant permissions on specific objects only. These functions do not support that.
// The difficulty is that the object is stored in [database_permissions] and [server_permissions] by its ID in the major_id column.
// With additional sub-object ID, like the column, in the minor_id column.
// It means we need to query multiple views based on the object type to retrieve the full definition of the permission.
// Schemas would be in sys.schemas, tables in sys.tables, columns in sys.columns, etc.
// #endregion

// #region Helper and Utility Functions
// ============================================================================
// HELPER AND UTILITY FUNCTIONS
// ============================================================================

// scanPermissionRow scans a SQL row into a Permission model
func scanPermissionRow(rows *sql.Rows) (*model.Permission, error) {
	var permission model.Permission
	err := rows.Scan(
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
	if err != nil {
		return nil, fmt.Errorf("failed to scan permission row: %w", err)
	}
	return &permission, nil
}

// scanPermissionRowFromSingleRow scans a SQL row from QueryRow into a Permission model
func scanPermissionRowFromSingleRow(row *sql.Row, permission *model.Permission) error {
	err := row.Scan(
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
	if err != nil {
		return fmt.Errorf("failed to scan permission row: %w", err)
	}
	return nil
}

// executePermissionsInTransaction executes a slice of permission operations within a transaction
func (c *Connector) executePermissionsInTransaction(ctx context.Context, db *sql.DB, operations []func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error, but don't override the original error
				fmt.Printf("error rolling back transaction: %v\n", rollbackErr)
			}
		}
	}()

	for _, operation := range operations {
		if err = operation(tx); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// #endregion

// #region Validation Functions
// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// Note: Database connection validation is now handled by sql.validateDatabaseConnection()
// and sql.validateDatabaseConnectionWithRetry() for consistency across all packages.

// validateRoleName validates that the role is not nil and has a valid name.
func validateRoleName(role *model.Role) error {
	if role == nil || role.Name == "" {
		return errors.New("role name cannot be empty")
	}
	return validateSQLIdentifier(role.Name)
}

// validatePermissionName validates that the permission is not nil and has a valid permission name.
func validatePermissionName(permission *model.Permission) error {
	if permission == nil || permission.Name == "" {
		return errors.New("permission name cannot be empty")
	}
	return validateSQLPermissionName(permission.Name)
}

// ValidatePermissionName is the exported version of validatePermissionName for testing
func ValidatePermissionName(permission *model.Permission) error {
	return validatePermissionName(permission)
}

// validatePermissionState validates the permission state and returns the appropriate SQL verb.
func validatePermissionState(permission *model.Permission) (string, error) {
	if (permission.State != "G" && permission.State != "D" && permission.State != "") ||
		(permission.StateDesc != "GRANT" && permission.StateDesc != "DENY" && permission.StateDesc != "") {
		return "", fmt.Errorf("invalid state value, must be 'G', 'D', 'GRANT', or 'DENY'")
	}

	stateVerb := "GRANT"
	if permission.State == "D" || permission.StateDesc == "DENY" {
		stateVerb = "DENY"
	}
	return stateVerb, nil
}

// validateSQLIdentifier validates that a string is a valid SQL identifier
func validateSQLIdentifier(name string) error {
	if name == "" {
		return errors.New("SQL identifier cannot be empty")
	}
	if len(name) > MaxSQLIdentifierLength {
		return fmt.Errorf("SQL identifier too long (max %d characters)", MaxSQLIdentifierLength)
	}
	if !sqlIdentifierRegex.MatchString(name) {
		return errors.New("invalid SQL identifier format")
	}
	return nil
}

// validateSQLPermissionName validates that a string is a valid SQL permission name
// Permission names can contain spaces and follow different rules than regular SQL identifiers
func validateSQLPermissionName(name string) error {
	if name == "" {
		return errors.New("permission name cannot be empty")
	}
	if len(name) > MaxSQLIdentifierLength {
		return fmt.Errorf("permission name too long (max %d characters)", MaxSQLIdentifierLength)
	}
	if !permissionNameRegex.MatchString(name) {
		return errors.New("invalid permission name format, must be uppercase letters, may contain spaces")
	}
	return nil
}

// validateSchemaName validates that a schema name is valid
func validateSchemaName(schema string) error {
	if schema == "" {
		return errors.New("schema name cannot be empty")
	}
	return validateSQLIdentifier(schema)
}

// #endregion

// #region Core Permission Operations - Database Level
// ============================================================================
// CORE PERMISSION OPERATIONS - DATABASE LEVEL
// ============================================================================

// AssignPermissionToRole assigns the specified permission, grant or deny, to a role in the database.
//
// This function provides a unified interface for both granting and denying permissions to database roles.
// The permission state is determined by the permission.State ("G" for GRANT, "D" for DENY) or
// permission.StateDesc ("GRANT" or "DENY") fields.
//
// Example usage:
//
//	role := &model.Role{Name: "db_reader"}
//	permission := &model.Permission{Name: "SELECT", State: "G"}
//	err := connector.AssignPermissionToRole(ctx, db, role, permission)
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection (must be valid and connected)
//   - role: Target role to assign permission to (name will be validated as SQL identifier)
//   - permission: Permission to assign with valid state and name
//
// Returns:
//   - nil if the permission is successfully assigned
//   - error if validation fails, connection is invalid, or database operation fails
//
// The function validates all inputs including SQL identifier format and executes the
// permission assignment within the current transaction context.
func (c *Connector) AssignPermissionToRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return err
	}
	if err := validatePermissionName(permission); err != nil {
		return err
	}

	// Validate the permission state and get the SQL verb
	stateVerb, err := validatePermissionState(permission)
	if err != nil {
		return err
	}

	// Validate database connection with retry logic
	if err := c.validateDatabaseConnectionWithRetry(ctx, db, 3); err != nil {
		return err
	}

	// SQL query to assign permissions to a role.
	query := fmt.Sprintf("'%s %s TO ' + QUOTENAME(@roleName)", stateVerb, permission.Name)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	// Execute the query.
	_, err = db.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return fmt.Errorf("query execution error - cannot assign permissions to role: %w", err)
	}

	// Return nil error.
	return nil
}

// GrantPermissionToRole grants the specified permission to a role in the database.
// It takes a context, a database connection, a role, and a permission as parameters.
// Returns nil if the permission is successfully granted to the role, otherwise returns an error.
func (c *Connector) GrantPermissionToRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	// Create a copy of the permission to avoid mutating the input parameter
	permCopy := *permission
	permCopy.State = "G"
	return c.AssignPermissionToRole(ctx, db, role, &permCopy)
}

// DenyPermissionToRole denies the specified permission to a role in the database.
// It takes a context, a database connection, a role, and a permission as parameters.
// Returns nil if the permission is successfully denied to the role, otherwise returns an error.
func (c *Connector) DenyPermissionToRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	// Create a copy of the permission to avoid mutating the input parameter
	permCopy := *permission
	permCopy.State = "D"
	return c.AssignPermissionToRole(ctx, db, role, &permCopy)
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

// RevokePermissionFromRole revokes the specified database permissions from a role.
// It takes a context, a database connection, a role, and a permission as parameters.
// If there is an error during the query execution, it returns an error.
// Otherwise, it returns nil.
func (c *Connector) RevokePermissionFromRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) error {
	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return err
	}
	if err := validatePermissionName(permission); err != nil {
		return err
	}

	// Validate database connection
	if err := c.validateDatabaseConnection(ctx, db); err != nil {
		return err
	}

	// SQL query to revoke permissions from a role.
	query := fmt.Sprintf("'REVOKE %s FROM ' + QUOTENAME(@roleName)", permission.Name)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	// Execute the query.
	_, err := db.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return fmt.Errorf("query execution error - cannot revoke permissions from role: %w", err)
	}

	// Return nil error.
	return nil
}

// RevokePermissionsFromRole revokes multiple database permissions from a role.
// It takes a context, a database connection, a role, and a slice of permissions as parameters.
// Returns nil if all permissions are successfully revoked, otherwise returns an error.
func (c *Connector) RevokePermissionsFromRole(ctx context.Context, db *sql.DB, role *model.Role, permissions []*model.Permission) error {
	for _, permission := range permissions {
		if err := c.RevokePermissionFromRole(ctx, db, role, permission); err != nil {
			return fmt.Errorf("failed to revoke permission %s: %w", permission.Name, err)
		}
	}
	return nil
}

// #endregion

// #region Schema-Level Permission Operations
// ============================================================================
// SCHEMA-LEVEL PERMISSION OPERATIONS
// ============================================================================

// AssignPermissionOnSchemaToRole assigns the specified permission, grant or deny, to a role on a specific schema in the database.
func (c *Connector) AssignPermissionOnSchemaToRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permission *model.Permission) error {
	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return err
	}
	if err := validatePermissionName(permission); err != nil {
		return err
	}

	if err := validateSchemaName(schema); err != nil {
		return err
	}

	// Validate the permission state and get the SQL verb
	stateVerb, err := validatePermissionState(permission)
	if err != nil {
		return err
	}

	// Validate database connection
	if err := c.validateDatabaseConnection(ctx, db); err != nil {
		return err
	}

	// SQL query to assign permissions to a role on a schema.
	query := fmt.Sprintf("'%s %s ON SCHEMA::%s TO ' + QUOTENAME(@roleName)", stateVerb, permission.Name, schema)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	// Execute the query.
	_, err = db.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return fmt.Errorf("query execution error - cannot assign permissions to role: %w", err)
	}

	// Return nil error.
	return nil
}

// GrantPermissionOnSchemaToRole grants the specified permission to a role on a specific schema in the database.
// It takes a context, a database connection, a role, schema name, and a permission as parameters.
// Returns nil if the permission is successfully granted to the role, otherwise returns an error.
func (c *Connector) GrantPermissionOnSchemaToRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permission *model.Permission) error {
	// Create a copy of the permission to avoid mutating the input parameter
	permCopy := *permission
	permCopy.State = "G"
	return c.AssignPermissionOnSchemaToRole(ctx, db, role, schema, &permCopy)
}

// DenyPermissionOnSchemaToRole denies the specified permission to a role on a specific schema in the database.
// It takes a context, a database connection, a role, schema name, and a permission as parameters.
// Returns nil if the permission is successfully denied to the role, otherwise returns an error.
func (c *Connector) DenyPermissionOnSchemaToRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permission *model.Permission) error {
	// Create a copy of the permission to avoid mutating the input parameter
	permCopy := *permission
	permCopy.State = "D"
	return c.AssignPermissionOnSchemaToRole(ctx, db, role, schema, &permCopy)
}

// GrantPermissionsOnSchemaToRole grants the specified permissions to a role on a specific schema in the database.
// It takes a context, a database connection, a role, schema name, and a slice of permissions as parameters.
// Returns nil if the permissions are successfully granted to the role, otherwise returns an error.
func (c *Connector) GrantPermissionsOnSchemaToRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permissions []*model.Permission) error {
	for _, permission := range permissions {
		err := c.GrantPermissionOnSchemaToRole(ctx, db, role, schema, permission)
		if err != nil {
			return err
		}
	}
	return nil
}

// DenyPermissionsOnSchemaToRole denies the specified permissions to a role on a specific schema in the database.
// It takes a context, a database connection, a role, schema name, and a slice of permissions as parameters.
// Returns nil if the permissions are successfully denied to the role, otherwise returns an error.
func (c *Connector) DenyPermissionsOnSchemaToRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permissions []*model.Permission) error {
	for _, permission := range permissions {
		err := c.DenyPermissionOnSchemaToRole(ctx, db, role, schema, permission)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevokePermissionOnSchemaFromRole revokes the specified schema permissions from a role.
// It takes a context, a database connection, a role, schema name, and a permission as parameters.
// If there is an error during the query execution, it returns an error.
// Otherwise, it returns nil.
func (c *Connector) RevokePermissionOnSchemaFromRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permission *model.Permission) error {
	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return err
	}
	if err := validatePermissionName(permission); err != nil {
		return err
	}

	if err := validateSchemaName(schema); err != nil {
		return err
	}

	// Validate database connection
	if err := c.validateDatabaseConnection(ctx, db); err != nil {
		return err
	}

	// SQL query to revoke permissions from a role on a schema.
	query := fmt.Sprintf("'REVOKE %s ON SCHEMA::%s FROM ' + QUOTENAME(@roleName)", permission.Name, schema)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	// Execute the query.
	_, err := db.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return fmt.Errorf("query execution error - cannot revoke schema permissions from role: %w", err)
	}

	// Return nil error.
	return nil
}

// RevokePermissionsOnSchemaFromRole revokes multiple schema permissions from a role.
// It takes a context, a database connection, a role, schema name, and a slice of permissions as parameters.
// Returns nil if all permissions are successfully revoked, otherwise returns an error.
func (c *Connector) RevokePermissionsOnSchemaFromRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permissions []*model.Permission) error {
	for _, permission := range permissions {
		if err := c.RevokePermissionOnSchemaFromRole(ctx, db, role, schema, permission); err != nil {
			return fmt.Errorf("failed to revoke schema permission %s: %w", permission.Name, err)
		}
	}
	return nil
}

// #endregion

// #region Transaction-Enabled Batch Operations
// ============================================================================
// TRANSACTION-ENABLED BATCH OPERATIONS
// ============================================================================

// GrantPermissionsToRoleWithTransaction grants the specified permissions to a role within a transaction.
// This ensures atomicity - either all permissions are granted or none are.
func (c *Connector) GrantPermissionsToRoleWithTransaction(ctx context.Context, db *sql.DB, role *model.Role, permissions []*model.Permission) error {
	operations := make([]func(*sql.Tx) error, len(permissions))
	for i, permission := range permissions {
		perm := permission // capture loop variable
		operations[i] = func(tx *sql.Tx) error {
			return c.assignPermissionToRoleInTx(ctx, tx, role, perm, "GRANT")
		}
	}
	return c.executePermissionsInTransaction(ctx, db, operations)
}

// DenyPermissionsToRoleWithTransaction denies the specified permissions to a role within a transaction.
//
// This function provides atomic permission denial - either all permissions are successfully denied
// or none are, ensuring database consistency. If any permission fails, the entire operation is
// rolled back automatically.
//
// Example usage:
//
//	permissions := []*model.Permission{
//	    {Name: "SELECT"}, {Name: "INSERT"}, {Name: "UPDATE"},
//	}
//	err := connector.DenyPermissionsToRoleWithTransaction(ctx, db, role, permissions)
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection (must support transactions)
//   - role: Target role to deny permissions from
//   - permissions: Slice of permissions to deny atomically
//
// Returns:
//   - nil if all permissions are successfully denied
//   - error if any validation fails or any permission denial fails (with rollback)
func (c *Connector) DenyPermissionsToRoleWithTransaction(ctx context.Context, db *sql.DB, role *model.Role, permissions []*model.Permission) error {
	operations := make([]func(*sql.Tx) error, len(permissions))
	for i, permission := range permissions {
		perm := permission // capture loop variable
		operations[i] = func(tx *sql.Tx) error {
			return c.assignPermissionToRoleInTx(ctx, tx, role, perm, "DENY")
		}
	}
	return c.executePermissionsInTransaction(ctx, db, operations)
}

// RevokePermissionsFromRoleWithTransaction revokes multiple database permissions from a role within a transaction.
func (c *Connector) RevokePermissionsFromRoleWithTransaction(ctx context.Context, db *sql.DB, role *model.Role, permissions []*model.Permission) error {
	operations := make([]func(*sql.Tx) error, len(permissions))
	for i, permission := range permissions {
		perm := permission // capture loop variable
		operations[i] = func(tx *sql.Tx) error {
			return c.revokePermissionFromRoleInTx(ctx, tx, role, perm)
		}
	}
	return c.executePermissionsInTransaction(ctx, db, operations)
}

// RevokePermissionsOnSchemaFromRoleWithTransaction revokes multiple schema permissions from a role within a transaction.
func (c *Connector) RevokePermissionsOnSchemaFromRoleWithTransaction(ctx context.Context, db *sql.DB, role *model.Role, schema string, permissions []*model.Permission) error {
	operations := make([]func(*sql.Tx) error, len(permissions))
	for i, permission := range permissions {
		perm := permission // capture loop variable
		operations[i] = func(tx *sql.Tx) error {
			return c.revokePermissionOnSchemaFromRoleInTx(ctx, tx, role, schema, perm)
		}
	}
	return c.executePermissionsInTransaction(ctx, db, operations)
}

// #endregion

// #region Private Transaction Helper Functions
// ============================================================================
// PRIVATE TRANSACTION HELPER FUNCTIONS
// ============================================================================

// assignPermissionToRoleInTx assigns a permission to a role within a transaction
func (c *Connector) assignPermissionToRoleInTx(ctx context.Context, tx *sql.Tx, role *model.Role, permission *model.Permission, verb string) error {
	query := fmt.Sprintf("'%s %s TO ' + QUOTENAME(@roleName)", verb, permission.Name)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err := tx.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))
	if err != nil {
		return fmt.Errorf("failed to %s permission %s to role %s: %w", verb, permission.Name, role.Name, err)
	}
	return nil
}

// revokePermissionFromRoleInTx revokes a permission from a role within a transaction
func (c *Connector) revokePermissionFromRoleInTx(ctx context.Context, tx *sql.Tx, role *model.Role, permission *model.Permission) error {
	query := fmt.Sprintf("'REVOKE %s FROM ' + QUOTENAME(@roleName)", permission.Name)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err := tx.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))
	if err != nil {
		return fmt.Errorf("failed to revoke permission %s from role %s: %w", permission.Name, role.Name, err)
	}
	return nil
}

// revokePermissionOnSchemaFromRoleInTx revokes a schema permission from a role within a transaction
func (c *Connector) revokePermissionOnSchemaFromRoleInTx(ctx context.Context, tx *sql.Tx, role *model.Role, schema string, permission *model.Permission) error {
	query := fmt.Sprintf("'REVOKE %s ON SCHEMA::%s FROM ' + QUOTENAME(@roleName)", permission.Name, schema)
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err := tx.ExecContext(ctx, tsql, sql.Named("roleName", role.Name))
	if err != nil {
		return fmt.Errorf("failed to revoke schema permission %s from role %s: %w", permission.Name, role.Name, err)
	}
	return nil
}

// #endregion

// #region Query/Retrieval Functions
// ============================================================================
// QUERY/RETRIEVAL FUNCTIONS
// ============================================================================

// GetDatabasePermissionsForRole retrieves the permissions for a given role from the database.
// It takes a context.Context, *sql.DB, and *model.Role as input parameters.
// It returns a slice of model.Permission and an error.
func (c *Connector) GetDatabasePermissionsForRole(ctx context.Context, db *sql.DB, role *model.Role) ([]model.Permission, error) {
	var permissions []model.Permission

	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return nil, err
	}

	// Validate database connection with retry logic
	if err := c.validateDatabaseConnectionWithRetry(ctx, db, 3); err != nil {
		return nil, err
	}

	// Execute the query using the predefined constant.
	rows, err := db.QueryContext(ctx, QueryDatabasePermissionsForRole, sql.Named("name", role.Name))

	// Check for any error during the query execution.
	if err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve permissions for role: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("error closing rows: %v\n", closeErr)
		}
	}()

	// Iterate through the resultset.
	for rows.Next() {
		// Scan the result into the Permission model using helper function.
		permission, err := scanPermissionRow(rows)
		if err != nil {
			return nil, err
		}

		// Append the permission to the permissions slice.
		permissions = append(permissions, *permission)
	}

	// Return the retrieved permissions.
	return permissions, nil
}

// GetDatabasePermissionForRole retrieves the permission details for a given role from the database.
// It takes a context.Context, *sql.DB, *model.Role, and a *model.Permission as input parameters.
// It returns a *model.Permission and an error.
func (c *Connector) GetDatabasePermissionForRole(ctx context.Context, db *sql.DB, role *model.Role, permission *model.Permission) (*model.Permission, error) {
	var err error

	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return nil, err
	}
	if err := validatePermissionName(permission); err != nil {
		return nil, err
	}

	// Validate database connection
	if err := c.validateDatabaseConnection(ctx, db); err != nil {
		return nil, err
	}

	// Execute the query using the predefined constant.
	row := db.QueryRowContext(
		ctx,
		QueryDatabasePermissionForRole,
		sql.Named("name", role.Name),
		sql.Named("permissionName", permission.Name))

	// Check for any error during the query execution.
	if row.Err() != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve permission for role: %w", row.Err())
	}

	// Scan the result into the Permission model using helper function.
	err = scanPermissionRowFromSingleRow(row, permission)

	// Check if the permission is not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("permissions not found")
	} else if err != nil {
		// Check for other scan errors.
		return nil, err
	}

	// Return the retrieved permissions.
	return permission, nil
}

// GetSchemaPermissionsForRole retrieves the permissions for a role on a specific schema in the database.
func (c *Connector) GetSchemaPermissionsForRole(ctx context.Context, db *sql.DB, role *model.Role, schema string) ([]model.Permission, error) {
	var permissions []model.Permission

	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return nil, err
	}

	if err := validateSchemaName(schema); err != nil {
		return nil, err
	}

	// Validate database connection
	if err := c.validateDatabaseConnection(ctx, db); err != nil {
		return nil, err
	}

	// Execute the query using the predefined constant.
	rows, err := db.QueryContext(ctx, QuerySchemaPermissionsForRole, sql.Named("roleName", role.Name), sql.Named("schemaName", schema))
	if err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve schema permissions for role: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("error closing rows: %v\n", closeErr)
		}
	}()

	// Iterate through the resultset.
	for rows.Next() {
		// Scan the result into the Permission model using helper function.
		permission, err := scanPermissionRow(rows)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, *permission)
	}

	return permissions, nil
}

// GetSchemaPermissionForRole retrieves a specific permission for a role on a specific schema in the database.
func (c *Connector) GetSchemaPermissionForRole(ctx context.Context, db *sql.DB, role *model.Role, schema string, permission *model.Permission) (*model.Permission, error) {
	var err error

	// Validate inputs
	if err := validateRoleName(role); err != nil {
		return nil, err
	}
	if err := validatePermissionName(permission); err != nil {
		return nil, err
	}

	if err := validateSchemaName(schema); err != nil {
		return nil, err
	}

	// Validate database connection
	if err := c.validateDatabaseConnection(ctx, db); err != nil {
		return nil, err
	}

	// Execute the query using the predefined constant.
	row := db.QueryRowContext(
		ctx,
		QuerySchemaPermissionForRole,
		sql.Named("roleName", role.Name),
		sql.Named("schemaName", schema),
		sql.Named("permissionName", permission.Name))

	// Check for any error during the query execution.
	if row.Err() != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve schema permission for role: %w", row.Err())
	}

	// Scan the result into the Permission model using helper function.
	err = scanPermissionRowFromSingleRow(row, permission)

	// Check if the permission is not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("permission not found")
	} else if err != nil {
		// Check for other scan errors.
		return nil, err
	}

	// Return the retrieved permission.
	return permission, nil
}

// #endregion

// #region Test Helper Functions
// ============================================================================
// TEST HELPER FUNCTIONS
// ============================================================================

// CreateTestPermission creates a test permission with the given name and state
func CreateTestPermission(name, state string) *model.Permission {
	return &model.Permission{
		Name:  name,
		State: state,
	}
}

// CreateTestRole creates a test role with the given name
func CreateTestRole(name string) *model.Role {
	return &model.Role{
		Name: name,
	}
}

// CreateTestPermissionWithStateDesc creates a test permission with name and state description
func CreateTestPermissionWithStateDesc(name, stateDesc string) *model.Permission {
	return &model.Permission{
		Name:      name,
		StateDesc: stateDesc,
	}
}

// #endregion
