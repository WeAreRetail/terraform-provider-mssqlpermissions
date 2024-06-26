package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"terraform-provider-mssqlpermissions/internal/queries/model"
)

// GetServerRole retrieves the server role information from the master database.
// It takes a context, a database connection, and a server role model as input.
// It returns the retrieved server role model and an error if any.
func (c *Connector) GetServerRole(ctx context.Context, db *sql.DB, serverRole *model.Role) (*model.Role, error) {
	var err error

	if c.Database != "master" {
		return nil, errors.New("cannot get server role from non master database")
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

	// SQL query to get a server role.
	query := `SELECT name, principal_id, type, type_desc, owning_principal_id, is_fixed_role
				FROM [master].[sys].[server_principals]
				WHERE [name] = @name AND type_desc = 'SERVER_ROLE'`

	// Execute the query and get a single row result.
	row := db.QueryRowContext(ctx, query, sql.Named("name", serverRole.Name))

	// Check for any error during the query execution.
	if err = row.Err(); err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve server role: %v", err)
	}

	// Scan the result into the ServerRole model.
	err = row.Scan(&serverRole.Name, &serverRole.PrincipalID, &serverRole.Type, &serverRole.TypeDescription, &serverRole.OwningPrincipal, &serverRole.IsFixedRole)

	// Check if the server role is not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("server role not found")
	} else if err != nil {
		// Check for other scan errors.
		return nil, fmt.Errorf("scan error - cannot retrieve server role. Underlying sql error : %v", err)
	}

	return serverRole, nil
}

// CreateServerRole creates a server role in the specified database. Not available on Azure Database.
// It takes a context, a database connection, and a server role model as input.
// The function returns an error if the server role creation fails.
func (c *Connector) CreateServerRole(ctx context.Context, db *sql.DB, serverRole *model.Role) error {
	var err error

	if c.isAzureDatabase {
		return errors.New("cannot create server role on Azure Database")
	}

	if c.Database != "master" {
		return errors.New("cannot create server role from non master database")
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

	// Check the provided PrincipalID. Defaulting to 1 if not provided.
	if serverRole.PrincipalID == 0 {
		serverRole.PrincipalID = 1
	}

	// Retrieve the login with the provided ID.
	login, err := c.GetLogin(ctx, db, &model.Login{PrincipalID: serverRole.PrincipalID})
	if err != nil {
		return fmt.Errorf("cannot get login with PrincipalID equals to %d. Underlying error : %v", serverRole.PrincipalID, err)
	}

	// SQL query to get a server role.
	query := "'CREATE SERVER ROLE ' + QUOTENAME(@server_role_name) + ' AUTHORIZATION ' + QUOTENAME(@login_name)"

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("server_role_name", serverRole.Name),
		sql.Named("login_name", login.Name))

	if err != nil {
		return fmt.Errorf("cannot create server role. Underlying sql error : %v", err)
	} else {
		return nil
	}
}

// DeleteServerRole deletes a server role from the database.
// It takes a context, a database connection, and a server role as parameters.
// It returns an error if the deletion fails, otherwise it returns nil.
func (c *Connector) DeleteServerRole(ctx context.Context, db *sql.DB, serverRole *model.Role) error {
	var err error

	if c.isAzureDatabase {
		return errors.New("cannot delete server role on Azure Database")
	}

	if c.Database != "master" {
		return errors.New("cannot delete server role from non master database")
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

	// The full TSQL script.
	// Adapted from the script generated by SSMS.
	tsql := `
	IF @RoleName <> N'public' and (select is_fixed_role from sys.server_principals where name = @RoleName) = 0
	BEGIN
		DECLARE @RoleMemberName sysname
		DECLARE Member_Cursor CURSOR FOR
		select [name]
		from sys.server_principals
		where principal_id in (
			select member_principal_id
			from sys.server_role_members
			where role_principal_id in (
				select principal_id
				FROM sys.server_principals where [name] = @RoleName  AND type = 'R' ))

		OPEN Member_Cursor;

		FETCH NEXT FROM Member_Cursor
		into @RoleMemberName

		DECLARE @SQL NVARCHAR(4000)

		WHILE @@FETCH_STATUS = 0
		BEGIN

			SET @SQL = 'ALTER SERVER ROLE '+ QUOTENAME(@RoleName,'[') +' DROP MEMBER '+ QUOTENAME(@RoleMemberName,'[')
			EXEC(@SQL)

			FETCH NEXT FROM Member_Cursor
			into @RoleMemberName
		END;

		CLOSE Member_Cursor;
		DEALLOCATE Member_Cursor;
	END;

	SET @SQL = 'DROP SERVER ROLE ' + QUOTENAME(@RoleName)
	EXEC(@SQL);
	`

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("RoleName", serverRole.Name))

	if err != nil {
		return fmt.Errorf("cannot delete server role. Underlying sql error : %v", err)
	} else {
		return nil
	}
}

// AddServerRoleMember adds a member to a server role.
// It takes a context, a database connection, a server role, and a login as parameters.
// It returns an error if the addition fails, otherwise it returns nil.
func (c *Connector) AddServerRoleMember(ctx context.Context, db *sql.DB, serverRole *model.Role, login *model.Login) error {
	var err error

	if c.isAzureDatabase {
		return errors.New("cannot add server role member on Azure Database")
	}

	if c.Database != "master" {
		return errors.New("cannot add server role member from non master database")
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

	// Validate the provided login.
	login, err = c.GetLogin(ctx, db, login)
	if err != nil {
		return fmt.Errorf("cannot get login. Underlying error : %v", err)
	}

	// Validate the provided server role.
	serverRole, err = c.GetServerRole(ctx, db, serverRole)
	if err != nil {
		return fmt.Errorf("cannot get server role. Underlying error : %v", err)
	}

	// SQL query to add a member to a server role.
	query := "'ALTER SERVER ROLE ' + QUOTENAME(@server_role_name) + ' ADD MEMBER ' + QUOTENAME(@login_name)"

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("server_role_name", serverRole.Name),
		sql.Named("login_name", login.Name))

	if err != nil {
		return fmt.Errorf("cannot add server role member. Underlying sql error : %v", err)
	} else {
		return nil
	}
}

// AddServerRoleMembers adds members to a server role.
// It takes a context, a database connection,a server role and a list of logins as parameters.
// It returns an error if the addition fails, otherwise it returns nil.
func (c *Connector) AddServerRoleMembers(ctx context.Context, db *sql.DB, serverRole *model.Role, logins []*model.Login) error {
	for _, login := range logins {
		err := c.AddServerRoleMember(ctx, db, serverRole, login)
		if err != nil {
			return fmt.Errorf("cannot add member to server role. Underlying error : %v", err)
		}
	}
	return nil
}

// RemoveServerRoleMember removes a member from a server role.
// It takes a context, a database connection, a login, and a server role as parameters.
// It returns an error if the removal fails, otherwise it returns nil.
func (c *Connector) RemoveServerRoleMember(ctx context.Context, db *sql.DB, serverRole *model.Role, login *model.Login) error {
	var err error

	if c.isAzureDatabase {
		return errors.New("cannot remove server role member on Azure Database")
	}

	if c.Database != "master" {
		return errors.New("cannot remove server role member from non master database")
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

	// Validate the provided login.
	login, err = c.GetLogin(ctx, db, login)
	if err != nil {
		return fmt.Errorf("cannot get login. Underlying error : %v", err)
	}

	// Validate the provided server role.
	serverRole, err = c.GetServerRole(ctx, db, serverRole)
	if err != nil {
		return fmt.Errorf("cannot get server role. Underlying error : %v", err)
	}

	// SQL query to remove a member from a server role.
	query := "'ALTER SERVER ROLE ' + QUOTENAME(@server_role_name) + ' DROP MEMBER ' + QUOTENAME(@login_name)"

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("server_role_name", serverRole.Name),
		sql.Named("login_name", login.Name))

	if err != nil {
		return fmt.Errorf("cannot remove server role member. Underlying sql error : %v", err)
	} else {
		return nil
	}
}

// RemoveServerRoleMembers removes members to a server role.
// It takes a context, a database connection,a server role and a list of logins as parameters.
// It returns an error if the remove fails, otherwise it returns nil.
func (c *Connector) RemoveServerRoleMembers(ctx context.Context, db *sql.DB, serverRole *model.Role, logins []*model.Login) error {
	for _, login := range logins {
		err := c.RemoveServerRoleMember(ctx, db, serverRole, login)
		if err != nil {
			return fmt.Errorf("cannot remove member to server role. Underlying error : %v", err)
		}
	}
	return nil
}

// GetServerRoleMembers retrieves the members of a server role.
// It takes a context, a database connection, and a server role as parameters.
// It returns a list of logins and an error if any.
func (c *Connector) GetServerRoleMembers(ctx context.Context, db *sql.DB, serverRole *model.Role) ([]*model.Login, error) {
	var err error
	var logins []*model.Login

	type ServerPrincipals struct {
		Name                string
		PrincipalID         int64
		SID                 string
		Type                string
		TypeDesc            string
		IsDisabled          bool
		DefaultDatabaseName sql.NullString
		DefaultLanguageName sql.NullString
	}

	// Check if the database connection is nil.
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Check if the database role is nil.
	if serverRole == nil {
		return nil, errors.New("server role is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to get the members of a server role.
	query := `SELECT [name], principal_id, sid, type, type_desc, is_disabled, default_database_name, default_language_name
				FROM [master].[sys].[server_principals]
				WHERE principal_id IN (
					SELECT member_principal_id
					FROM [master].[sys].[server_role_members]
					WHERE role_principal_id = (
						SELECT principal_id
						FROM [master].[sys].[server_principals]
						WHERE [name] = @server_role_name AND type_desc = 'SERVER_ROLE'))`

	// Execute the query
	rows, err := db.QueryContext(ctx, query, sql.Named("server_role_name", serverRole.Name))
	if err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve server role members: %v", err)
	}

	// Check for any error during the query execution.
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve server role members: %v", err)
	}

	// Scan the result into the ServerPrincipals model.
	for rows.Next() {
		var result ServerPrincipals

		err = rows.Scan(
			&result.Name,
			&result.PrincipalID,
			&result.SID,
			&result.Type,
			&result.TypeDesc,
			&result.IsDisabled,
			&result.DefaultDatabaseName,
			&result.DefaultLanguageName)
		if err != nil {
			return nil, fmt.Errorf("scan error - cannot retrieve server role members: %v", err)
		}

		login := &model.Login{
			Name: result.Name,
		}

		login, err = c.GetLogin(ctx, db, login)
		if err != nil {
			return nil, fmt.Errorf("cannot get login. Underlying error : %v", err)
		}

		logins = append(logins, login)
	}

	// Return the retrieved server role information.
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve server role members. Underlying sql error : %v", err)
	} else {
		return logins, nil
	}
}
