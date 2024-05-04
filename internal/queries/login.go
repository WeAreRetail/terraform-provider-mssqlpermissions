package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"terraform-provider-mssqlpermissions/internal/queries/model"
)

// GetLogin retrieves login information based on the provided login name or principal ID.
// If the database is an Azure Database and the specified database is not "master", it returns an error.
// It returns the retrieved login information or an error if the login is not found or there is an error during the query execution.
func (c *Connector) GetLogin(ctx context.Context, db *sql.DB, login *model.Login) (*model.Login, error) {
	var err error

	// Check if it's an Azure Database and if the specified database is not "master".
	if c.isAzureDatabase && c.Database != "master" {
		return nil, errors.New("cannot get logins from non master database on Azure Database")
	}

	// Check if the database connection is nil.
	if db == nil {
		err = errors.New("database connection is nil")
		return nil, err
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to retrieve login information based on the login name.
	query := "SELECT name, principal_id, type, is_disabled, default_database_name, default_language_name FROM "
	if login.External {
		query = query + "[master].[sys].[server_principals]"
	} else {
		query = query + "[master].[sys].[sql_logins]"
	}
	if login.Name != "" {
		query = query + " WHERE [name] = @name"
	} else if login.PrincipalID != 0 {
		query = query + " WHERE [principal_id] = @principal_id"
	}

	// Execute the query and get a single row result.
	row := db.QueryRowContext(ctx, query, sql.Named("name", login.Name), sql.Named("principal_id", login.PrincipalID))

	// Check for any error during the query execution.
	if err = row.Err(); err != nil {
		return nil, fmt.Errorf("query execution error - cannot retrieve login: %v", err)
	}

	// Scan the result into the login model.
	err = row.Scan(&login.Name, &login.PrincipalID, &login.Type, &login.Is_Disabled, &login.DefaultDatabase, &login.DefaultLanguage)

	// Check if the login is not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("login not found")
	} else if err != nil {
		// Check for other scan errors.
		return nil, fmt.Errorf("scan error - cannot retrieve login: %v", err)
	}

	return login, nil
}

// CreateLogin creates a new login in the database.
// If the database is an Azure Database and the specified database is not "master", it returns an error.
// It returns an error if the database connection is nil, the database ping fails, or there is an error during the login creation.
func (c *Connector) CreateLogin(ctx context.Context, db *sql.DB, login *model.Login) error {
	//var login model.Login
	var err error

	// Check if it's an Azure Database and if the specified database is not "master".
	if c.isAzureDatabase && c.Database != "master" {
		return errors.New("cannot create logins on non master database on Azure Database")
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

	// SQL query to create a login
	// Note: CREATE LOGIN doesn't accept parameters. Working around by building the query string then executing it.
	query := "'CREATE LOGIN ' + QUOTENAME(@name)"

	if login.External {
		query = query + " + ' FROM EXTERNAL PROVIDER'"
	} else {
		query = query + " + ' WITH PASSWORD = ' + QUOTENAME(@password, '''')"
	}

	if !c.isAzureDatabase {
		if login.DefaultDatabase == "" {
			login.DefaultDatabase = "master"
		}
		query = query + " + ', DEFAULT_DATABASE = ' + QUOTENAME(@defaultDatabase)"

		if login.DefaultLanguage != "" && login.DefaultLanguage != c.defaultLanguage {
			query = query + "+ ', DEFAULT_LANGUAGE = ' + QUOTENAME(@defaultLanguage)"
		}
	}

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", login.Name),
		sql.Named("password", login.Password),
		sql.Named("defaultDatabase", login.DefaultDatabase),
		sql.Named("defaultLanguage", login.DefaultLanguage))

	if err != nil {
		return fmt.Errorf("cannot create login. Underlying sql error : %v", err)
	} else {
		return nil
	}
}

// UpdateLogin updates an existing login in the database.
// If the database is an Azure Database and the specified database is not "master", it returns an error.
// It returns an error if the database connection is nil, the database ping fails, or there is an error during the login update.
func (c *Connector) UpdateLogin(ctx context.Context, db *sql.DB, login *model.Login) error {
	var err error

	// Check if it's an Azure Database and if the specified database is not "master".
	if c.isAzureDatabase && c.Database != "master" {
		return errors.New("cannot get logins from non master database on Azure Database")
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

	// SQL query to create a login
	// Note: CREATE LOGIN doesn't accept parameters. Working around by building the query string then executing it.
	query := "'ALTER LOGIN ' + QUOTENAME(@name)"
	updateRequired := false

	if !login.External && login.Password != "" {
		query = query + " + ' WITH PASSWORD = ' + QUOTENAME(@password, '''')"
		updateRequired = true
	}

	if !c.isAzureDatabase {
		if login.DefaultDatabase == "" {
			login.DefaultDatabase = "master"
			query = query + " + ', DEFAULT_DATABASE = ' + QUOTENAME(@defaultDatabase)"
			updateRequired = true
		}
		if login.DefaultLanguage != "" && login.DefaultLanguage != c.defaultLanguage {
			query = query + "+ ', DEFAULT_LANGUAGE = ' + QUOTENAME(@defaultLanguage)"
			updateRequired = true
		}
	}

	// Nothing to update.
	if !updateRequired {
		return nil
	}

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", login.Name),
		sql.Named("password", login.Password),
		sql.Named("defaultDatabase", login.DefaultDatabase),
		sql.Named("defaultLanguage", login.DefaultLanguage))

	if err != nil {
		return fmt.Errorf("cannot update login. Underlying sql error : %v", err)
	} else {
		return nil
	}
}

// DeleteLogin deletes a login from the database.
// If the database is an Azure Database and the specified database is not "master", it returns an error.
// It returns an error if the database connection is nil, the database ping fails, or there is an error during the login deletion.
func (c *Connector) DeleteLogin(ctx context.Context, db *sql.DB, login *model.Login) error {
	var err error

	if err = c.killSessionsForLogin(ctx, db, login.Name); err != nil {
		return err
	}

	// Check if it's an Azure Database and if the specified database is not "master".
	if c.isAzureDatabase && c.Database != "master" {
		return errors.New("cannot get logins from non master database on Azure Database")
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

	// SQL query to delete a login
	query := "'DROP LOGIN ' + QUOTENAME(@name)"

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", login.Name),
	)

	if err != nil {
		return fmt.Errorf("cannot delete login. Underlying sql error : %v", err)
	}

	return nil
}

// killSessionsForLogin kills all sessions for a login in the database.
// If the database is an Azure Database and the specified database is not "master", it returns an error.
// It returns an error if the database connection is nil, the database ping fails, or there is an error during the session killing.
func (c *Connector) killSessionsForLogin(ctx context.Context, db *sql.DB, loginName string) error {
	var err error

	// Check if it's an Azure Database and if the specified database is not "master".
	if c.isAzureDatabase && c.Database != "master" {
		return errors.New("cannot get logins from non master database on Azure Database")
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

	// SQL query kill all sessions for a login
	// From https://stackoverflow.com/a/5178097/38055
	query := `DECLARE sessionsToKill CURSOR FAST_FORWARD FOR
			  SELECT session_id
			  FROM sys.dm_exec_sessions
			  WHERE login_name = @name
			OPEN sessionsToKill
			DECLARE @sessionId INT
			DECLARE @statement NVARCHAR(200)
			FETCH NEXT FROM sessionsToKill INTO @sessionId
			WHILE @@FETCH_STATUS = 0
			BEGIN
			  PRINT 'Killing session ' + CAST(@sessionId AS NVARCHAR(20)) + ' for login ' + @name
			  SET @statement = 'KILL ' + CAST(@sessionId AS NVARCHAR(20))
			  EXEC sp_executesql @statement
			  FETCH NEXT FROM sessionsToKill INTO @sessionId
			END
			CLOSE sessionsToKill
			DEALLOCATE sessionsToKill`

	// The full TSQL script.
	tsql := query

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", loginName))

	if err != nil {
		return fmt.Errorf("cannot kill login sessions. Underlying sql error : %v", err)
	} else {
		return nil
	}
}
