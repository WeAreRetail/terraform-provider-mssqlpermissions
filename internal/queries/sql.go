package queries

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	azuread "github.com/microsoft/go-mssqldb/azuread"
)

type FedAuth string

func (f FedAuth) String() string {
	return string(f)
}

// Constants.
const (
	defaultTimeout = 30 * time.Second

	// Represents different authentication options.
	ActiveDirectoryServicePrincipal FedAuth = "ActiveDirectoryServicePrincipal"
	ActiveDirectoryApplication      FedAuth = "ActiveDirectoryApplication"
	ActiveDirectoryPassword         FedAuth = "ActiveDirectoryPassword"
	ActiveDirectoryDefault          FedAuth = "ActiveDirectoryDefault"
	ActiveDirectoryManagedIdentity  FedAuth = "ActiveDirectoryManagedIdentity"
	ActiveDirectoryMSI              FedAuth = "ActiveDirectoryMSI"
	ActiveDirectoryInteractive      FedAuth = "ActiveDirectoryInteractive"
	ActiveDirectoryDeviceCode       FedAuth = "ActiveDirectoryDeviceCode"
	ActiveDirectoryAzCli            FedAuth = "ActiveDirectoryAzCli"
)

// Connector holds the configuration and authentication details required to connect to a SQL Server instance.
// It supports multiple authentication methods including local user login, Azure application login, and managed identity login.
// The struct also includes metadata about the target database, such as whether it is an Azure or contained database,
// as well as connection parameters like host, port, database name, timeout, and default language.
type Connector struct {
	Host                  string
	Port                  int
	Database              string
	Timeout               time.Duration
	LocalUserLogin        *LocalUserLogin
	AzureApplicationLogin *AzureApplicationLogin
	ManagedIdentityLogin  *ManagedIdentityLogin
	isAzureDatabase       bool
	defaultLanguage       string
}

// AzureApplicationLogin represents Azure Active Directory application login details.
type AzureApplicationLogin struct {
	ClientCertificatePath     string // TODO: implement certificate
	ClientCertificatePassword string // TODO: implement certificate
	ClientId                  string
	ClientSecret              string
	TenantId                  string
}

// ManagedIdentityLogin represents Managed Identity login details.
type ManagedIdentityLogin struct {
	UserIdentity bool
	UserId       string
	ResourceId   string
}

// LocalUserLogin represents local user login details.
type LocalUserLogin struct {
	Username string
	Password string
}

// connector returns a driver.Connector based on the specified authentication method.
func (c *Connector) connector() (driver.Connector, error) {
	// Validate the Connector fields before creating the driver.Connector
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	connectionString := &url.URL{
		Scheme: "sqlserver",
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
	}

	query := url.Values{}
	query.Add("database", c.Database)
	query.Add("app name", "terraform-sql-provider")

	// Determine the authentication method and construct the connection string accordingly
	switch {
	case c.LocalUserLogin != nil:
		connectionString.User = url.UserPassword(c.LocalUserLogin.Username, c.LocalUserLogin.Password)
		connectionString.RawQuery = query.Encode()
		return mssql.NewConnector(connectionString.String())
	case c.AzureApplicationLogin != nil:
		return c.configureAzureADConnector(connectionString, query)
	case c.ManagedIdentityLogin != nil:
		return c.configureManagedIdentityConnector(connectionString, query)
	default:
		query.Add("fedauth", ActiveDirectoryDefault.String())
		connectionString.RawQuery = query.Encode()
		return azuread.NewConnector(connectionString.String())
	}
}

// validate checks if required fields in Connector are provided.
func (c *Connector) validate() error {
	if c.Host == "" {
		return errors.New("missing host name")
	}
	if c.Database == "" {
		return errors.New("missing database name")
	}
	if c.Port == 0 {
		c.Port = 1433
	}
	return nil
}

// configureAzureADConnector configures the connection string for Azure AD authentication.
func (c *Connector) configureAzureADConnector(connectionString *url.URL, query url.Values) (driver.Connector, error) {
	userId := c.AzureApplicationLogin.ClientId
	if c.AzureApplicationLogin.TenantId != "" {
		userId = fmt.Sprintf("%s@%s", c.AzureApplicationLogin.ClientId, c.AzureApplicationLogin.TenantId)
	}

	query.Add("fedauth", ActiveDirectoryServicePrincipal.String())
	query.Add("user id", userId)
	query.Add("password", c.AzureApplicationLogin.ClientSecret)
	connectionString.RawQuery = query.Encode()
	return azuread.NewConnector(connectionString.String())
}

// configureManagedIdentityConnector configures the connection string for Managed Identity authentication.
func (c *Connector) configureManagedIdentityConnector(connectionString *url.URL, query url.Values) (driver.Connector, error) {
	query.Add("fedauth", ActiveDirectoryManagedIdentity.String())

	if c.ManagedIdentityLogin.UserIdentity && (c.ManagedIdentityLogin.UserId != "" || c.ManagedIdentityLogin.ResourceId != "") {
		if c.ManagedIdentityLogin.UserId != "" {
			query.Add("user id", c.ManagedIdentityLogin.UserId)
		}
		if c.ManagedIdentityLogin.ResourceId != "" {
			query.Add("resource id", c.ManagedIdentityLogin.ResourceId)
		}
	}

	connectionString.RawQuery = query.Encode()
	return azuread.NewConnector(connectionString.String())
}

// getVersion retrieves the version of the connected SQL Server.
func (c *Connector) getVersion(ctx context.Context, db *sql.DB) (string, error) {
	var version string
	var err error

	if db == nil {
		err = errors.New("connection is null")
		return "", err
	}

	// Check if database is alive.
	err = db.PingContext(ctx)
	if err != nil {
		return "", err
	}

	query := "SELECT @@VERSION"

	row := db.QueryRowContext(ctx, query)
	if err = row.Err(); err != nil {
		return "", fmt.Errorf("cannot retrieve version: %w", err)
	}

	err = row.Scan(&version)

	return version, err
}

// getDefaultLanguage retrieves the default language of the connected SQL Server.
func (c *Connector) getDefaultLanguage(ctx context.Context, db *sql.DB) (string, error) {
	var defaultLanguage string
	var err error

	if db == nil {
		err = errors.New("connection is null")
		return "", err
	}

	// Check if database is alive.
	err = db.PingContext(ctx)
	if err != nil {
		return "", err
	}

	query := "SELECT lang.name FROM [sys].[configurations] config INNER JOIN [sys].[syslanguages] lang ON config.[value] = lang.langid WHERE config.name = 'default language'"

	row := db.QueryRowContext(ctx, query)
	if err = row.Err(); err != nil {
		return "", fmt.Errorf("cannot retrieve default language: %w", err)
	}

	err = row.Scan(&defaultLanguage)

	return defaultLanguage, err
}

// getContainedStatus confirms if the connected database is a contained database.
func (c *Connector) containedEnabled(ctx context.Context, db *sql.DB) (bool, error) {
	var containedEnabled bool
	var err error

	if db == nil {
		err = errors.New("connection is null")
		return false, err
	}

	// Check if database is alive.
	err = db.PingContext(ctx)
	if err != nil {
		return false, err
	}

	query := "SELECT value_in_use FROM [sys].[configurations] WHERE name = 'contained database authentication'"

	row := db.QueryRowContext(ctx, query)
	if err = row.Err(); err != nil {
		return false, fmt.Errorf("cannot retrieve contained database status: %w", err)
	}

	err = row.Scan(&containedEnabled)

	return containedEnabled, err
}

// Connect establishes a connection to the SQL Server using the configured authentication method.
// It validates the connection, retrieves server metadata, and ensures database compatibility.
// Returns a *sql.DB connection and an error if the connection fails or database is incompatible.
func (c *Connector) Connect() (*sql.DB, error) {
	if c == nil {
		return nil, errors.New("no connector provided")
	}

	// Set default timeout if not provided
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}

	driverConnector, err := c.connector()
	if err != nil {
		return nil, fmt.Errorf("error creating driver connector: %w", err)
	}

	db := sql.OpenDB(driverConnector)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %s", err)
	}

	// Get the Server version and update isAzureDatabase accordingly
	version, err := c.getVersion(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("error retrieving the server version: %s", err)
	}

	if strings.Contains(version, "Microsoft SQL Azure") {
		c.isAzureDatabase = true
	}

	// Get the Server version and update isAzureDatabase accordingly
	defaultLanguage, err := c.getDefaultLanguage(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("error retrieving the server default language: %s", err)
	}

	c.defaultLanguage = defaultLanguage

	// Get the Database Contained status and update isContainedDatabase accordingly
	isContainedDatabase, err := c.containedEnabled(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("error retrieving the contained status: %s", err)
	}

	// This provider only supports contained databases. Return an error if the database is not contained.
	if !isContainedDatabase && !c.isAzureDatabase {
		return nil, fmt.Errorf("the target database is not a contained database. This provider only supports contained databases")
	}

	return db, nil
}

// validateDatabaseConnection validates that the database connection is not nil and is alive.
// This is a common validation pattern used across all database operations.
func (c *Connector) validateDatabaseConnection(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("database connection is nil")
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// validateDatabaseConnectionWithRetry validates database connection with retry logic.
// This should be used for query operations that might benefit from retrying on transient failures.
func (c *Connector) validateDatabaseConnectionWithRetry(ctx context.Context, db *sql.DB, maxRetries int) error {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for i := 0; i < maxRetries; i++ {
		if err := c.validateDatabaseConnection(ctx, db); err == nil {
			return nil
		}
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond * 100 * time.Duration(i+1)):
				// Continue to next retry
			}
		}
	}
	return c.validateDatabaseConnection(ctx, db) // Return the final error
}
