package queries

import (
	"context"
	"fmt"
	model "queries/model"
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
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			wantErr: false,
		},
		{
			name:      "create-on-LocalSQL-Contained-NoPassword",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				Contained: true,
			},
			wantErr: true,
		},
		{
			name:      "create-on-LocalSQL-Contained-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained:       true,
				DefaultLanguage: "Français",
			},
			wantErr: false,
		},
		{
			name:      "create-on-LocalSQL-Instance",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				LoginName: generateRandomString(10),
				Contained: false,
			},
			wantErr: false,
		},
		{
			name:      "create-on-LocalSQL-Instance-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				LoginName:       generateRandomString(10),
				DefaultLanguage: "Français",
				Contained:       false,
			},
			wantErr: true,
		},
		{
			name:      "create-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
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

			// Create a login for the test if needed
			var login *model.Login = nil
			if tt.user.LoginName != "" {
				login = &model.Login{
					Name:     tt.user.LoginName,
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				}
				tt.connector.CreateLogin(ctx, db, login)
			}

			// Invoke the CreateUser method with the provided login model.
			err := tt.connector.CreateUser(ctx, db, tt.user)

			// Cleanup
			var errCleanupLogin error = nil
			if tt.user.LoginName != "" {
				errCleanupLogin = tt.connector.DeleteLogin(ctx, db, login)
			}

			errCleanupUser := tt.connector.DeleteUser(ctx, db, tt.user)

			// Restore the original database state.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateUser() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errCleanupUser != nil || errCleanupLogin != nil {
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
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-dbo-on-LocalSQL",
			connector: testConnectors.localSQL,
			user: &model.User{
				PrincipalID: 0,
				Contained:   true,
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
				Contained:       true,
				DefaultLanguage: "Français",
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-on-LocalSQL-Instance",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				LoginName: generateRandomString(10),
				Contained: false,
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			preCreate: true,
			wantErr:   false,
		},
		{
			name:      "get-on-dbo-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				PrincipalID: 0,
				Contained:   true,
			},
			preCreate: false,
			wantErr:   false,
		},
	}

	// Iterate through each test case.
	for _, tt := range tests {
		// Run each test case in a sub-test to isolate and identify failures.
		t.Run(tt.name, func(t *testing.T) {
			var err error = nil
			// Capture the original state of the connector's database for restoration.
			dbRestore := tt.connector.Database

			// Override the database if specified in the test case.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			// Set up the context and connect to the database.
			ctx := context.Background()
			db, _ := tt.connector.Connect()

			// Create a login for the test if needed
			var login *model.Login = nil
			if tt.user.LoginName != "" && tt.preCreate {
				login = &model.Login{
					Name:     tt.user.LoginName,
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				}
				err = tt.connector.CreateLogin(ctx, db, login)
				if err != nil {
					t.Errorf("Test case %s: Unable to create the login to get: %v", tt.name, err)
					return
				}
			}

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
			var errCleanupLogin error = nil

			if tt.preCreate {
				if tt.user.LoginName != "" {
					errCleanupLogin = tt.connector.DeleteLogin(ctx, db, login)
					t.Logf("Cleanup login: %s", login.Name)
				}
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
				if errCleanupUser != nil || errCleanupLogin != nil {
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
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
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
				Contained:       true,
				DefaultLanguage: "Français",
			},
			updatedUser: &model.User{
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				DefaultLanguage: "English",
			},
			wantErr: false,
		},
		{
			name:      "update-on-LocalSQL-Instance",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				LoginName: generateRandomString(10),
				Contained: false,
			},
			updatedUser: &model.User{
				LoginName: generateRandomString(10),
			},
			wantErr: false,
		},
		{
			name:      "update-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
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

			// Create a login for the test if needed
			var login *model.Login = nil
			if tt.user.LoginName != "" {
				login = &model.Login{
					Name:     tt.user.LoginName,
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				}
				tt.connector.CreateLogin(ctx, db, login)
			}

			// Create an updated login for the test if needed
			var updatedLogin *model.Login = nil
			if tt.updatedUser.LoginName != "" {
				updatedLogin = &model.Login{
					Name:     tt.updatedUser.LoginName,
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				}
				tt.connector.CreateLogin(ctx, db, updatedLogin)
			}

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

				// Cleanup login if needed

				var errCleanupLogin error = nil
				if tt.user.LoginName != "" {
					errCleanupLogin = tt.connector.DeleteLogin(ctx, db, login)
				}

				var errCleanupUpdatedLogin error = nil
				if tt.updatedUser.LoginName != "" {
					errCleanupUpdatedLogin = tt.connector.DeleteLogin(ctx, db, updatedLogin)
				}

				if (errCleanupUser != nil) || (errCleanupLogin != nil) || (errCleanupUpdatedLogin != nil) {
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
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			wantErr: false,
		},
		{
			name:      "delete-on-LocalSQL-Contained-with-DefaultLanguage",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:            generateRandomString(10),
				Password:        fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained:       true,
				DefaultLanguage: "Français",
			},
			wantErr: false,
		},
		{
			name:      "delete-on-LocalSQL-Instance",
			connector: testConnectors.localSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				LoginName: generateRandomString(10),
				Contained: false,
			},
			wantErr: false,
		},
		{
			name:      "delete-on-azureSQL-Contained",
			connector: testConnectors.azureSQL,
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
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

			// Create a login for the test if needed
			var login *model.Login = nil
			if tt.user.LoginName != "" {
				login = &model.Login{
					Name:     tt.user.LoginName,
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				}
				tt.connector.CreateLogin(ctx, db, login)
			}

			// Invoke the CreateUser method with the provided login model.
			err := tt.connector.CreateUser(ctx, db, tt.user)

			if err != nil {
				t.Errorf("Test case %s: Unable to create the user to get: %v", tt.name, err)
				return
			}

			err = tt.connector.DeleteUser(ctx, db, tt.user)

			// Cleanup login if needed
			if tt.user.LoginName != "" {
				errCleanupLogin := tt.connector.DeleteLogin(ctx, db, login)
				if errCleanupLogin != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}

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
