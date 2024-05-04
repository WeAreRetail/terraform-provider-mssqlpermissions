package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"terraform-provider-mssqlpermissions/internal/queries/model"
)

// validateUser validates the given user object.
// It checks if the user has a name, and performs additional validations based on the user's properties.
// If any validation fails, it returns an error indicating the reason.
func (c *Connector) validateUser(user *model.User) error {

	if user.Name == "" {
		return errors.New("a user must have a name")
	}

	if user.Contained {
		if user.LoginName != "" {
			return errors.New("a contained user cannot have a login name")
		}
		if user.Password == "" && !user.External {
			return errors.New("a contained user must have a password if it's not external")
		}
	}

	if !user.Contained {
		if user.LoginName == "" && !user.External {
			return errors.New("a not contained and not external user must have a login")
		}

		if user.External && (user.LoginName != "" || user.Password != "") {
			return errors.New("a not contained external user cannot have a password or a login")
		}

		if user.DefaultLanguage != "" {
			return errors.New("a not contained user cannot have a default language")

		}
	}

	if user.External {
		if user.Password != "" {
			return errors.New("an external user cannot have a password")
		}
	}

	if user.ObjectID != "" {
		if !user.External {
			return errors.New("only external user can specify an ObjectID")
		}
	}

	if user.DefaultLanguage != "" {
		if c.isAzureDatabase {
			return errors.New("a user cannot have a default language in an Azure Database")
		}
	}

	return nil
}

// CreateUser creates a user on the specified database.
// It returns a user object with the login name populated.
// If any error occurs, it returns an error object with the reason.
func (c *Connector) CreateUser(ctx context.Context, db *sql.DB, user *model.User) error {

	var err error

	// Validate the user object.
	err = c.validateUser(user)
	if err != nil {
		return fmt.Errorf("cannot create user. validation failed : %v", err)
	}

	// Set the default schema to dbo if it's not specified.
	if user.DefaultSchema == "" {
		user.DefaultSchema = "dbo"
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

	// If a login name is specified, check if the login exists.
	if user.LoginName != "" {
		_, err = c.GetLogin(ctx, db, &model.Login{Name: user.LoginName})
		if err != nil {
			return fmt.Errorf("issue with the provided login: %v", err)
		}
	}

	// SQL query to create a user
	// Note: CREATE USER doesn't accept parameters. Working around by building the query string then executing it.
	query := "'CREATE USER ' + QUOTENAME(@name)"

	// The authentication type is Azure Active Directory.
	if user.External {
		if c.isAzureDatabase || user.Contained { // The database is an Azure Database or the user is contained.
			query = query + " + ' FROM EXTERNAL PROVIDER'"

			if c.isAzureDatabase && user.ObjectID != "" {
				// Link to a SPN in Azure Active Directory.
				// Note: this is an undocumented, unsupported option. See https://github.com/MicrosoftDocs/sql-docs/issues/2323
				query = query + " + ' WITH OBJECT_ID= ' + QuoteName(@objectID)"
			}

		} else { // The database is not an Azure Database and the user is not contained.
			query = query + " + ' FOR LOGIN ' + QuoteName(@loginName) + ' FROM EXTERNAL PROVIDER WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)"
		}
	} else { // The authentication type is SQL Server authentication.
		if user.Contained {
			if !c.isContainedDatabase {
				return errors.New("cannot create a user with a password in a non-contained database")
			}
			query = query + " + ' WITH PASSWORD = ' + QUOTENAME(@password, '''') + ', DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)"

			if !c.isAzureDatabase {
				// Set the default language to NONE if it's not specified.
				if user.DefaultLanguage == "" {
					query = query + " + ', DEFAULT_LANGUAGE = NONE'"
				} else {
					query = query + " + ', DEFAULT_LANGUAGE = ' + QuoteName(@defaultLanguage)"
				}
			}

		} else { // The user is not contained.
			query = query + " + ' FOR LOGIN ' + QuoteName(@loginName) + ' WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)"
		}
	}

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", user.Name),
		sql.Named("password", user.Password),
		sql.Named("loginName", user.LoginName),
		sql.Named("objectID", user.ObjectID),
		sql.Named("defaultSchema", user.DefaultSchema),
		sql.Named("defaultLanguage", user.DefaultLanguage))

	if err != nil {
		return fmt.Errorf("cannot create user. Underlying sql error : %v", err)
	}

	return nil
}

// GetUser retrieves a user from the database based on the provided user name.
// It takes a context, a database connection, and a user object as input.
// It returns the retrieved user object and an error if any.
func (c *Connector) GetUser(ctx context.Context, db *sql.DB, user *model.User) (*model.User, error) {
	var err error

	type DatabasePrincipals struct {
		Name                   string
		PrincipalID            int64
		Type                   string
		TypeDesc               string
		DefaultSchemaName      sql.NullString
		SID                    string
		AuthenticationType     int
		AuthenticationTypeDesc string
		DefaultLanguageName    sql.NullString
	}

	var result DatabasePrincipals

	// Check if the database connection is nil.
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to retrieve a user
	query := "SELECT [name], [principal_id], [type], [type_desc], [default_schema_name], CONVERT(varchar(max), [sid], 1) as [sid], [authentication_type], [authentication_type_desc], [default_language_name] FROM sys.database_principals"

	if user.Name != "" {
		query = query + " WHERE [name] = @name"
	} else if user.PrincipalID != 0 {
		query = query + " WHERE [principal_id] = @principal_id"
	}
	// Execute query
	row := db.QueryRowContext(ctx, query, sql.Named("name", user.Name), sql.Named("principal_id", user.PrincipalID))

	// Populate the result object with the result of the query.
	err = row.Scan(
		&result.Name,
		&result.PrincipalID,
		&result.Type,
		&result.TypeDesc,
		&result.DefaultSchemaName,
		&result.SID,
		&result.AuthenticationType,
		&result.AuthenticationTypeDesc,
		&result.DefaultLanguageName)

	// Check if the user was not found.
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("cannot retrieve user: %v", err)
	}

	// Populate the user object with the result.
	user.Name = result.Name
	if result.AuthenticationTypeDesc == "EXTERNAL" {
		user.External = true
	} else {
		user.External = false
	}
	if result.AuthenticationTypeDesc == "DATABASE" {
		user.Contained = true
	} else {
		user.Contained = false
	}
	user.DefaultSchema = result.DefaultSchemaName.String
	user.DefaultLanguage = result.DefaultLanguageName.String
	user.SID = result.SID
	user.PrincipalID = result.PrincipalID

	return user, nil
}

// UpdateUser updates a user on the specified database.
// If any error occurs, it returns an error object with the reason.
func (c *Connector) UpdateUser(ctx context.Context, db *sql.DB, user *model.User) error {
	var err error

	// Get the original user
	originalUser, err := c.GetUser(ctx, db, user)
	if err != nil {
		return fmt.Errorf("cannot retrieve the user to update. Underlying sql error : %v", err)
	}

	// if originalUser.External {
	// 	return errors.New("cannot update an external user")
	// }

	// Check if the database connection is nil.
	if db == nil {
		return errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	var altered = false // Flag to indicate if the user has been altered.

	// SQL query to update a user
	query := "'ALTER USER ' + QUOTENAME(@name) + ' WITH '"

	if user.DefaultSchema != "" && user.DefaultSchema != originalUser.DefaultSchema {
		altered = true
		query = query + " + 'DEFAULT_SCHEMA = ' + QuoteName(user.DefaultSchema) + ', '"
	}

	if !c.isAzureDatabase && user.DefaultLanguage != originalUser.DefaultLanguage {
		altered = true
		// Set the default language to NONE if it's not specified.
		if user.DefaultLanguage == "" {
			query = query + " + 'DEFAULT_LANGUAGE = NONE' + ', '"
		} else {
			query = query + " + 'DEFAULT_LANGUAGE = ' + QuoteName(@defaultLanguage) + ', '"
		}
	}

	if user.Password != "" {
		altered = true
		query = query + " + 'PASSWORD = ' + QUOTENAME(@password, '''') + ', '"
	}

	if user.LoginName != originalUser.LoginName {
		altered = true
		query = query + " + 'LOGIN = ' + QuoteName(@loginName) + ', '"
	}

	if !altered {
		return nil
	}

	// Trim the trailing comma and space from the query
	query = strings.TrimSuffix(query, " + ', '")

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", user.Name),
		sql.Named("password", user.Password),
		sql.Named("loginName", user.LoginName),
		sql.Named("defaultSchema", user.DefaultSchema),
		sql.Named("defaultLanguage", user.DefaultLanguage))

	if err != nil {
		return fmt.Errorf("cannot update user. Underlying sql error : %v", err)
	}

	return nil

}

// DeleteUser deletes a user
// If any error occurs, it returns an error object with the reason.
func (c *Connector) DeleteUser(ctx context.Context, db *sql.DB, user *model.User) error {
	var err error

	// Get the original user
	_, err = c.GetUser(ctx, db, user)
	if err != nil {
		return fmt.Errorf("cannot retrieve the user to delete. Underlying sql error : %v", err)
	}

	// if originalUser.External {
	// 	return errors.New("cannot delete an external user")
	// }

	// Check if the database connection is nil.
	if db == nil {
		return errors.New("database connection is nil")
	}

	// Check if the database is alive by pinging it.
	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	// SQL query to delete a user
	query := "'DROP USER ' + QUOTENAME(@name)"

	// The full TSQL script.
	tsql := fmt.Sprintf("DECLARE @sql NVARCHAR(MAX)\nSET @sql = %s;\nEXEC (@sql)", query)

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("name", user.Name),
	)

	if err != nil {
		return fmt.Errorf("cannot delete user. Underlying sql error : %v", err)
	}

	return nil
}
