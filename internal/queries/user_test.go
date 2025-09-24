//go:build integration

package queries

import (
	"context"
	"fmt"
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

func TestConnector_CreateUser(t *testing.T) {
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		user             *model.User // Login information for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "create-on-LocalSQL-Contained",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:      "create-on-LocalSQL-Contained-NoPassword",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name: generateRandomString(10),
			},
			wantErr: true,
		},
		{
			name:      "create-on-LocalSQL-Contained-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultLanguage: "Français",
			},
			wantErr: false,
		},
		{
			name:      "create-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
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

			// Invoke the CreateUser method with the provided login model.
			err := tt.connector.CreateUser(ctx, db, tt.user)

			errCleanupUser := tt.connector.DeleteUser(ctx, db, tt.user)

			// Restore the original database state.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateUser() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errCleanupUser != nil {
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

func TestConnector_GetUser(t *testing.T) {
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		user             *model.User // Login information for the test case.
		preCreate        bool        // Do not create the user before getting it.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "get-on-LocalSQL-Contained",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-dbo-on-LocalSQL",
			connector: testConnectors.localSQL,
			user: &model.User{
				PrincipalID: 0,
			},
			preCreate: false,
			wantErr:   false,
		},
		{
			name:      "get-on-LocalSQL-Contained-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultLanguage: "Français",
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-on-dbo-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				PrincipalID: 0,
			},
			preCreate: false,
			wantErr:   false,
		},
	}

	// Iterate through each test case.
	for _, tt := range tests {
		// Run each test case in a sub-test to isolate and identify failures.
		t.Run(tt.name, func(t *testing.T) {
			var err error
			// Capture the original state of the connector's database for restoration.
			dbRestore := tt.connector.Database

			// Override the database if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Set up the context and connect to the database.
			ctx := context.Background()
			db, _ := tt.connector.Connect()

			// Invoke the CreateUser method with the provided login model.
			if tt.preCreate {
				err = tt.connector.CreateUser(ctx, db, tt.user)
				if err != nil {
					t.Errorf("Test case %s: Unable to create the user to get: %v", tt.name, err)
					// Restore the original database state.
					tt.connector.Database = dbRestore
					return
				}
			}

			_, err = tt.connector.GetUser(ctx, db, tt.user)

			var errCleanupUser error = nil

			if tt.preCreate {
				errCleanupUser = tt.connector.DeleteUser(ctx, db, tt.user)
			}

			// Restore the original database state.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateUser() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errCleanupUser != nil {
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

func TestConnector_UpdateUser(t *testing.T) {
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		user             *model.User // User information for the test case.
		updatedUser      *model.User // Updated user information for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "update-on-LocalSQL-Contained",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			updatedUser: &model.User{
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:      "update-on-LocalSQL-Contained-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultLanguage: "Français",
			},
			updatedUser: &model.User{
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultLanguage: "English",
			},
			wantErr: false,
		},
		{
			name:      "update-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			updatedUser: &model.User{
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

			tt.updatedUser.Name = tt.user.Name

			// Invoke the CreateUser method with the provided login model.
			err := tt.connector.CreateUser(ctx, db, tt.user)

			if err != nil {
				t.Errorf("Test case %s: Unable to create the user to get: %v", tt.name, err)

				// Restore the original database state.
				tt.connector.Database = dbRestore
				return
			}

			err = tt.connector.UpdateUser(ctx, db, tt.user)

			// Restore the original database state.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateUser() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				errCleanupUser := tt.connector.DeleteUser(ctx, db, tt.user)

				if errCleanupUser != nil {
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

func TestConnector_DeleteUser(t *testing.T) {
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		user             *model.User // Login information for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "delete-on-LocalSQL-Contained",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
			},
			wantErr: false,
		},
		{
			name:      "delete-on-LocalSQL-Contained-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultLanguage: "Français",
			},
			wantErr: false,
		},
		{
			name:      "delete-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
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

			// Invoke the CreateUser method with the provided login model.
			err := tt.connector.CreateUser(ctx, db, tt.user)

			if err != nil {
				t.Errorf("Test case %s: Unable to create the user to get: %v", tt.name, err)
				return
			}

			err = tt.connector.DeleteUser(ctx, db, tt.user)

			// Restore the original database state.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateUser() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
		})
	}
}
