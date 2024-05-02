// Package queries provides functionality for handling database queries in different environments.
package queries

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"testing"
)

// runAzureTests and runLocalTests are flags to determine whether to run tests in Azure or local environment.
var runAzureTests bool
var runLocalTests bool

// localTestConfig represents the configuration for local testing.
var localTestConf localTestConfig

// azureTestConfig represents the configuration for Azure testing.
var azureTestConf azureTestConfig

// A structure with the different connections.
var testConnectors testConnection

// localTestConfig represents the configuration for local testing.
type localTestConfig struct {
	username string
	password string
	server   string
	database string
}

// azureTestConfig represents the configuration for Azure testing.
type azureTestConfig struct {
	tenantId          string
	adminClientId     string
	adminClientSecret string
	sqlPassword       string
	sqlUser           string
	userClientId      string
	userClientSecret  string
	database          string
	server            string
}

// testConnection represents the different connections available for testing.
type testConnection struct {
	localSQL      *Connector
	azureSQL      *Connector
	azureAadAdmin *Connector
}

// TestMain is the main function for running tests. It loads environment variables and exits with the test status.
func TestMain(m *testing.M) {
	// Call flag.Parse() here if TestMain uses flags.
	err := loadVars()

	initConnectors()

	if err != nil {
		log.Fatal(err.Error())
	}
	os.Exit(m.Run())
}

// checkEnvVars checks if the required environment variables are set for running tests.
func checkEnvVars() error {
	var keys []string
	_, localTest := os.LookupEnv("LOCAL_TEST")
	_, azureTest := os.LookupEnv("AZURE_TEST")

	if localTest {
		keys = append(keys,
			"LOCAL_MSSQL_DATABASE",
			"LOCAL_MSSQL_USERNAME",
			"LOCAL_MSSQL_PASSWORD",
			"LOCAL_MSSQL_SERVER")
	}
	if azureTest {
		keys = append(keys,
			"AZURE_TENANT_ID",
			"AZURE_MSSQL_DATABASE",
			"AZURE_MSSQL_ADMIN_CLIENT_ID",
			"AZURE_MSSQL_ADMIN_CLIENT_SECRET",
			"AZURE_MSSQL_PASSWORD",
			"AZURE_MSSQL_SERVER",
			"AZURE_MSSQL_USERNAME",
			"AZURE_MSSQL_USER_CLIENT_ID",
			"AZURE_MSSQL_USER_CLIENT_SECRET",
			"AZURE_MSSQL_USER_CLIENT_DISPLAY_NAME")
	}
	for _, key := range keys {
		if _, exist := os.LookupEnv(key); !exist {
			return fmt.Errorf("missing required environment variable %s for tests", key)
		}
	}
	return nil
}

// loadVars loads the environment variables for the respective testing environments.
func loadVars() error {
	if err := checkEnvVars(); err != nil {
		return fmt.Errorf("cannot load variable: %v", err)
	}
	_, runLocalTests = os.LookupEnv("LOCAL_TEST")
	_, runAzureTests = os.LookupEnv("AZURE_TEST")

	if runLocalTests {
		localTestConf.password = os.Getenv("LOCAL_MSSQL_PASSWORD")
		localTestConf.server = os.Getenv("LOCAL_MSSQL_SERVER")
		localTestConf.username = os.Getenv("LOCAL_MSSQL_USERNAME")
		localTestConf.database = os.Getenv("LOCAL_MSSQL_DATABASE")

	}

	if runAzureTests {
		azureTestConf.database = os.Getenv("AZURE_MSSQL_DATABASE")
		azureTestConf.server = os.Getenv("AZURE_MSSQL_SERVER")
		azureTestConf.adminClientId = os.Getenv("AZURE_MSSQL_ADMIN_CLIENT_ID")
		azureTestConf.adminClientSecret = os.Getenv("AZURE_MSSQL_ADMIN_CLIENT_SECRET")
		azureTestConf.sqlPassword = os.Getenv("AZURE_MSSQL_PASSWORD")
		azureTestConf.sqlUser = os.Getenv("AZURE_MSSQL_USERNAME")
		azureTestConf.tenantId = os.Getenv("AZURE_TENANT_ID")
		azureTestConf.userClientId = os.Getenv("AZURE_MSSQL_USER_CLIENT_ID")
		azureTestConf.userClientSecret = os.Getenv("AZURE_MSSQL_USER_CLIENT_DISPLAY_NAME")
	}
	fmt.Printf("Local test configuration loaded")

	return nil
}

// initConnections initializes the connectors that will be used trough the tests.
func initConnectors() {

	localSQLConnector := &Connector{
		Host:     localTestConf.server,
		Database: localTestConf.database,
		LocalUserLogin: &LocalUserLogin{
			Username: localTestConf.username,
			Password: localTestConf.password,
		},
	}

	azureSQLConnector := &Connector{
		Host:     azureTestConf.server,
		Database: azureTestConf.database,
		LocalUserLogin: &LocalUserLogin{
			Username: azureTestConf.sqlUser,
			Password: azureTestConf.sqlPassword,
		},
	}

	azureAadAdminConnector := &Connector{
		Host:     azureTestConf.server,
		Database: azureTestConf.database,
		AzureApplicationLogin: &AzureApplicationLogin{
			ClientId:     azureTestConf.adminClientId,
			ClientSecret: azureTestConf.adminClientSecret,
		},
	}

	testConnectors = testConnection{
		localSQL:      localSQLConnector,
		azureSQL:      azureSQLConnector,
		azureAadAdmin: azureAadAdminConnector,
	}
}

func generateRandomString(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(randomBytes)[:length]
}
