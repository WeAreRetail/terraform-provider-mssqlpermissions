package queries

import (
	"context"
	"fmt"
	model "queries/model"
	"reflect"
	"testing"
)

// TestConnector_GetLogin tests the GetLogin method of the Connector type.
func TestConnector_GetLogin(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string       // Name of the test case.
		connector        *Connector   // Connector instance to be tested.
		databaseOverride string       // Override for the database, if any.
		login            *model.Login // Login information for the test case.
		wantErr          bool         // Expected error condition.
	}{
		{
			name:      "sa-exists-on-LocalSQL",
			connector: testConnectors.localSQL,
			login: &model.Login{
				Name: "sa",
			},
			wantErr: false,
		},
		{
			name:      "1-exists-on-LocalSQL",
			connector: testConnectors.localSQL,
			login: &model.Login{
				PrincipalID: 1,
			},
			wantErr: false,
		},
		{
			name:             "the-admin-exists-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name: "the-admin",
			},
			wantErr: false, // GetLogin only works on master database for Azure
		},
		{
			name:             "1-exists-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				PrincipalID: 1,
			},
			wantErr: false, // GetLogin only works on master database for Azure
		},
		{
			name:             "the-admin-exists-on-azureAadAdmin",
			connector:        testConnectors.azureAadAdmin,
			databaseOverride: "master",
			login: &model.Login{
				Name: "the-admin",
			},
			wantErr: false, // GetLogin only works on master database for Azure
		},
		{
			name:      "cannot-use-GetLogin-on-non-master-azureSQL",
			connector: testConnectors.azureSQL,
			login: &model.Login{
				Name: "the-admin",
			},
			wantErr: true, // GetLogin only works on master database for Azure
		},
		{
			name:      "cannot-use-GetLogin-on-non-master-azureAadAdmin",
			connector: testConnectors.azureAadAdmin,
			login: &model.Login{
				Name: "the-admin",
			},
			wantErr: true, // GetLogin only works on master database for Azure
		},
	}

	// Iterate through the test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()

			// Check if the connection was successful.
			if err != nil {
				t.Errorf("Test case %s: Connector.Connect() error = %v", tt.name, err)

				// Restore the original database value.
				tt.connector.Database = dbRestore

				return
			}

			gotLogin, err := tt.connector.GetLogin(ctx, db, tt.login)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetLogin() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if (err != nil) == tt.wantErr {
				// Error expected, return.
				return
			}

			// Check the type of PrincipalID in the returned login.
			principalIDType := reflect.TypeOf(gotLogin.PrincipalID)
			if principalIDType != reflect.TypeOf(int64(0)) {
				t.Errorf("Test case %s: Invalid type for PrincipalID is %s while expecting int64", tt.name, principalIDType)
			}
			// Uncomment the following line to log the type of PrincipalID.
			// t.Logf("principalIDType: %s", principalIDType)
		})
	}
}

// TestConnector_CreateLogin tests the CreateLogin method of the Connector.
// This test suite includes various scenarios to ensure the proper functioning
// of creating logins with different configurations and databases.
func TestConnector_CreateLogin(t *testing.T) {
	// Define test cases with different inputs and expected outcomes.
	tests := []struct {
		name             string       // Test case name for identification.
		connector        *Connector   // Connector instance for the test.
		databaseOverride string       // Override for the database, if any.
		login            *model.Login // Input login model for the test case.
		wantErr          bool         // Expected error outcome.
	}{
		{
			name:             "create-on-master-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:      "create-on-db-LocalSQL",
			connector: testConnectors.localSQL,
			login: &model.Login{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultDatabase: "ApplicationDB",
				DefaultLanguage: "Français",
			},
			wantErr: false,
		},
		{
			name:             "create-does-not-meet-password-complexity-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: "a",
			},
			wantErr: true,
		},
		{
			name:             "create-on-master-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:      "create-on-db-azureSQL",
			connector: testConnectors.azureSQL,
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: true,
		},
		{
			name:             "create-does-not-meet-password-complexity-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: "a",
			},
			wantErr: true,
		},
		{
			name:             "create-external-unknown-azureAadAdmin",
			connector:        testConnectors.azureAadAdmin,
			databaseOverride: "master",
			login: &model.Login{
				Name:     "unknown@weareretail.ai",
				External: true,
			},
			wantErr: true,
		},
	}

	// Iterate through each test case.
	for _, tt := range tests {
		// Run each test case in a sub-test to isolate and identify failures.
		t.Run(tt.name, func(t *testing.T) {
			// Capture the original state of the connector's database for restoration.
			dbRestore := tt.connector.Database

			// Override the database if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Set up the context and connect to the database.
			ctx := context.Background()
			db, _ := tt.connector.Connect()

			// Invoke the CreateLogin method with the provided login model.
			err := tt.connector.CreateLogin(ctx, db, tt.login)
			errCleanup := tt.connector.DeleteLogin(ctx, db, tt.login)

			// Restore the original database state.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateLogin() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}
		})
	}
}

// TestConnector_UpdateLogin tests the UpdateLogin method of the Connector.
// This test suite includes various scenarios to ensure the proper functioning
// of updating logins with different configurations and databases.
func TestConnector_UpdateLogin(t *testing.T) {
	// Define test cases with different inputs and expected outcomes.
	tests := []struct {
		name             string       // Test case name for identification.
		connector        *Connector   // Connector instance for the test.
		databaseOverride string       // Override for the database, if any.
		login            *model.Login // Input login model for the test case.
		LoginUpdate      *model.Login // Input for updating the login.
		wantErr          bool         // Expected error outcome.
	}{
		{
			name:             "update-password-on-master-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			LoginUpdate: &model.Login{
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultDatabase: "ApplicationDB",
				DefaultLanguage: "Français",
			},
			wantErr: false,
		},
		{
			name:      "update-on-db-LocalSQL",
			connector: testConnectors.localSQL,
			login: &model.Login{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultDatabase: "master",
				DefaultLanguage: "Italiano",
			},
			LoginUpdate: &model.Login{
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultDatabase: "ApplicationDB",
				DefaultLanguage: "us_english",
			},
			wantErr: false,
		},
		{
			name:             "update-does-not-meet-password-complexity-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			LoginUpdate: &model.Login{
				Password: "a",
			},
			wantErr: true,
		},
		{
			name:             "update-on-master-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			LoginUpdate: &model.Login{
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:             "update-does-not-meet-password-complexity-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			LoginUpdate: &model.Login{
				Password: "a",
			},
			wantErr: true,
		},
	}

	// Iterate through each test case.
	for _, tt := range tests {
		// Run each test case in a sub-test to isolate and identify failures.
		t.Run(tt.name, func(t *testing.T) {
			// Capture the original state of the connector's database for restoration.
			dbRestore := tt.connector.Database

			// Override the database if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Set up the context and connect to the database.
			ctx := context.Background()
			db, _ := tt.connector.Connect()

			// Create a login before performing the update test.
			err := tt.connector.CreateLogin(ctx, db, tt.login)
			tt.connector.Database = dbRestore

			// Check for errors during login creation.
			if err != nil {
				t.Errorf("Test case %s: error on create login before update test: %v", tt.name, err)
				return
			}

			// Override the database again if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Update the login with the provided update model.
			tt.LoginUpdate.Name = tt.login.Name
			err = tt.connector.UpdateLogin(ctx, db, tt.LoginUpdate)
			errCleanup := tt.connector.DeleteLogin(ctx, db, tt.LoginUpdate)
			tt.connector.Database = dbRestore

			// Check if the error outcome matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetLogin() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}

			if (err != nil) == tt.wantErr {
				// Error expected, return.
				return
			}
		})
	}
}

// TestConnector_DeleteLogin tests the DeleteLogin method of the Connector.
// This test suite includes various scenarios to ensure the proper functioning
// of deleting logins with different configurations and databases.
func TestConnector_DeleteLogin(t *testing.T) {
	// Define test cases with different inputs and expected outcomes.
	tests := []struct {
		name             string       // Test case name for identification.
		connector        *Connector   // Connector instance for the test.
		databaseOverride string       // Override for the database, if any.
		login            *model.Login // Input login model for the test case.
		wantErr          bool         // Expected error outcome.
	}{
		{
			name:             "delete-password-on-master-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:      "delete-on-db-LocalSQL",
			connector: testConnectors.localSQL,
			login: &model.Login{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultDatabase: "master",
				DefaultLanguage: "Italiano",
			},
			wantErr: false,
		},
		{
			name:             "delete-on-master-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
	}

	// Iterate through each test case.
	for _, tt := range tests {
		// Run each test case in a sub-test to isolate and identify failures.
		t.Run(tt.name, func(t *testing.T) {
			// Capture the original state of the connector's database for restoration.
			dbRestore := tt.connector.Database

			// Override the database if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Set up the context and connect to the database.
			ctx := context.Background()
			db, _ := tt.connector.Connect()

			// Create a login before performing the delete test.
			err := tt.connector.CreateLogin(ctx, db, tt.login)
			tt.connector.Database = dbRestore

			// Check for errors during login creation.
			if err != nil {
				t.Errorf("Test case %s: error on create login before delete test: %v", tt.name, err)
				return
			}

			// Override the database again if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Delete the login
			err = tt.connector.DeleteLogin(ctx, db, tt.login)
			tt.connector.Database = dbRestore

			// Check if the error outcome matches the expectation.
			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetLogin() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if (err != nil) == tt.wantErr {
				// Error expected, return.
				return
			}
		})
	}
}
