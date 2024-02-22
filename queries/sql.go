package queries

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	azuread "github.com/microsoft/go-mssqldb/azuread"
)

// The database connection
var db *sql.DB

type FedAuth string

func (f FedAuth) String() string {
	return string(f)
}

// Constants
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

// Connector represents a connection to a SQL Server, optionally with a database.
type Connector struct {
	Host                  string
	Port                  int
	Database              string
	Timeout               time.Duration
	LocalUserLogin        *LocalUserLogin
	AzureApplicationLogin *AzureApplicationLogin
	ManagedIdentityLogin  *ManagedIdentityLogin
	isAzureDatabase       bool
	isContainedDatabase   bool
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
		return nil, fmt.Errorf("validation error: %v", err)
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
		log.Fatal(fmt.Errorf("cannot retrieve version: %v", err))
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
		log.Fatal(fmt.Errorf("cannot retrieve version: %v", err))
	}

	err = row.Scan(&defaultLanguage)

	return defaultLanguage, err
}

// getContainedStatus retrieves the default language of the connected SQL Server.
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
		log.Fatal(fmt.Errorf("cannot retrieve version: %v", err))
	}

	err = row.Scan(&containedEnabled)

	return containedEnabled, err
}

// Connect establishes a connection to the SQL Server.
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
		return nil, fmt.Errorf("error creating driver connector: %v", err)
	}

	db = sql.OpenDB(driverConnector)

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

	// Azure SQL Database are always contained database even if the setting is not present
	c.isContainedDatabase = isContainedDatabase || c.isAzureDatabase

	return db, nil
}
