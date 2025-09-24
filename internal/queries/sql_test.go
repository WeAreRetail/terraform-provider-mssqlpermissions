//go:build integration

package queries

import (
	"log"
	"testing"
	"time"
)

// TestConnector_validate is a test function for the validate method of the Connector struct.
func TestConnector_validate(t *testing.T) {

	// fields struct definition representing the fields of the Connector struct.
	type fields struct {
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
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "with hostname, database and port",
			fields: fields{
				Host:     "sql.example.com",
				Database: "app_db",
				Port:     1433,
			},
			wantErr: false,
		}, {
			name: "with hostname, database",
			fields: fields{
				Host:     "sql.example.com",
				Database: "app_db",
			},
			wantErr: false,
		}, {
			name: "with hostname",
			fields: fields{
				Host: "sql.example.com",
			},
			wantErr: true,
		}, {
			name: "with database",
			fields: fields{
				Database: "app_db",
			},
			wantErr: true,
		},
	}

	// Iterating through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Creating a Connector instance with the specified fields
			c := &Connector{
				Host:                  tt.fields.Host,
				Port:                  tt.fields.Port,
				Database:              tt.fields.Database,
				Timeout:               tt.fields.Timeout,
				LocalUserLogin:        tt.fields.LocalUserLogin,
				AzureApplicationLogin: tt.fields.AzureApplicationLogin,
				ManagedIdentityLogin:  tt.fields.ManagedIdentityLogin,
				isAzureDatabase:       tt.fields.isAzureDatabase,
				defaultLanguage:       tt.fields.defaultLanguage,
			}

			// Validating the Connector instance
			if err := c.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Connector.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConnector_connector_local_sql_authentication is a test function for connecting to a local server using SQL authentication.
func TestConnector_connector_local_sql_authentication(t *testing.T) {
	// Creating a Connector instance for local SQL authentication
	connector := &Connector{
		Host:     localTestConf.server,
		Database: localTestConf.database,
		LocalUserLogin: &LocalUserLogin{
			Username: localTestConf.username,
			Password: localTestConf.password,
		},
	}

	// Connecting to the database and checking if the connection is successful
	db, err := connector.Connect() // The actual function to test
	if err != nil {
		t.Errorf("cannot connect to local server using SQL Authentication: %s", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()
}

// TestConnector_connector_azure_sql_authentication is a test function for connecting to a local server using Azure SQL authentication.
func TestConnector_connector_azure_sql_authentication(t *testing.T) {
	connector := &Connector{
		Host:     azureTestConf.server,
		Database: azureTestConf.database,
		LocalUserLogin: &LocalUserLogin{
			Username: azureTestConf.sqlUser,
			Password: azureTestConf.sqlPassword,
		},
	}

	// Connecting to the database and checking if the connection is successful
	db, err := connector.Connect() // The actual function to test
	if err != nil {
		t.Errorf("cannot connect to local server using SQL Authentication: %s", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()
}

// TestConnector_connector_azure_aad_authentication is a test function for connecting to a local server using Azure AAD authentication.
func TestConnector_connector_azure_aad_authentication(t *testing.T) {
	connector := &Connector{
		Host:     azureTestConf.server,
		Database: azureTestConf.database,
		AzureApplicationLogin: &AzureApplicationLogin{
			ClientId:     azureTestConf.adminClientId,
			ClientSecret: azureTestConf.adminClientSecret,
		},
	}

	db, err := connector.Connect() // The actual function to test
	if err != nil {
		t.Errorf("cannot connect to local server using AAD Authentication with SPN: %s", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()
}
