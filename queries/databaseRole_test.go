package queries

import (
	"context"
	"fmt"
	model "queries/model"
	"reflect"
	"testing"
)

// TestConnector_GetDatabaseRole is a test function for the GetDatabaseRole method of the Connector struct.
func TestConnector_GetDatabaseRole(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		databaseRole     *model.Role // Database role for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:             "db_owner-exists-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			databaseRole: &model.Role{
				Name: "db_owner",
			},
			wantErr: false,
		},
		{
			name:             "db_owner-exists-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			databaseRole: &model.Role{
				Name: "db_owner",
			},
			wantErr: false,
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
			db, _ := tt.connector.Connect()

			// Call the function to test.
			gotDatabaseRole, err := tt.connector.GetDatabaseRole(ctx, db, tt.databaseRole)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetDatabaseRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			// Check the type of PrincipalID in the returned databaseRole.
			principalIDType := reflect.TypeOf(gotDatabaseRole.PrincipalID)
			if principalIDType != reflect.TypeOf(int64(0)) {
				t.Errorf("Test case %s: Invalid type for PrincipalID is %s while expecting int64", tt.name, principalIDType)
			}
			// Uncomment the following line to log the type of PrincipalID.
			// t.Logf("principalIDType: %s", principalIDType)
		})
	}
}

func TestConnector_CreateDatabaseRole(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		databaseRole     *model.Role // Database role for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "testers-on-LocalSQL",
			connector: testConnectors.localSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: false,
		},
		{
			name:      "testers-on-azureSQL",
			connector: testConnectors.azureSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: false,
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
			db, _ := tt.connector.Connect()

			// Call the function to test.
			err := tt.connector.CreateDatabaseRole(ctx, db, tt.databaseRole)

			// Cleanup the database role.
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.databaseRole)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateDatabaseRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
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

func TestConnector_DeleteDatabaseRole(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		databaseRole     *model.Role // Database role for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "testers-on-LocalSQL",
			connector: testConnectors.localSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: false,
		},
		{
			name:      "testers-on-azureSQL",
			connector: testConnectors.azureSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: false,
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
			db, _ := tt.connector.Connect()

			// Create the database role to delete
			err := tt.connector.CreateDatabaseRole(ctx, db, tt.databaseRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			}

			// Call the function to test.
			err = tt.connector.DeleteDatabaseRole(ctx, db, tt.databaseRole)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.DeleteDatabaseRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
		})
	}
}

func TestConnector_AddDatabaseRoleMember(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		databaseRole     *model.Role // Database role for the test case.
		user             *model.User // User to add to the database role.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "add-user-on-LocalSQL",
			connector: testConnectors.localSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			wantErr: false,
		},
		{
			name:      "add-user-on-azureSQL",
			connector: testConnectors.azureSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			wantErr: false,
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
			db, _ := tt.connector.Connect()

			// Setup the role
			errRoleCreate := tt.connector.CreateDatabaseRole(ctx, db, tt.databaseRole)
			if errRoleCreate != nil {
				t.Errorf("Test case %s: error during role setup = %v", tt.name, errRoleCreate)
				return
			}

			// Setup the user
			errUserCreate := tt.connector.CreateUser(ctx, db, tt.user)
			if errUserCreate != nil {
				t.Errorf("Test case %s: error during user setup = %v", tt.name, errUserCreate)
				return
			}

			// Call the function to test.
			err := tt.connector.AddDatabaseRoleMember(ctx, db, tt.databaseRole, tt.user)

			errRoleCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.databaseRole)
			errUserCleanup := tt.connector.DeleteUser(ctx, db, tt.user)

			if errRoleCleanup != nil {
				t.Errorf("Test case %s: error during role cleanup = %v", tt.name, errRoleCleanup)
			}

			if errUserCleanup != nil {
				t.Errorf("Test case %s: error during role cleanup = %v", tt.name, errUserCleanup)
			}

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.AddDatabaseRoleMember() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errRoleCleanup != nil || errUserCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}
		})
	}
}

func TestConnector_RemoveDatabaseRoleMember(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		databaseRole     *model.Role // Database role for the test case.
		user             *model.User // User to add to the database role.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:      "remove-user-on-LocalSQL",
			connector: testConnectors.localSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			wantErr: false,
		},
		{
			name:      "remove-user-on-azureSQL",
			connector: testConnectors.azureSQL,
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			user: &model.User{
				Name:      generateRandomString(10),
				Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
				Contained: true,
			},
			wantErr: false,
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
			db, _ := tt.connector.Connect()

			// Setup the role
			errRoleCreate := tt.connector.CreateDatabaseRole(ctx, db, tt.databaseRole)
			if errRoleCreate != nil {
				t.Errorf("Test case %s: error during role setup = %v", tt.name, errRoleCreate)
				return
			}

			// Setup the user
			errUserCreate := tt.connector.CreateUser(ctx, db, tt.user)
			if errUserCreate != nil {
				t.Errorf("Test case %s: error during user setup = %v", tt.name, errUserCreate)
				return
			}

			// Setup the user in the role
			errAddUser := tt.connector.AddDatabaseRoleMember(ctx, db, tt.databaseRole, tt.user)
			if errAddUser != nil {
				t.Errorf("Test case %s: error during adding user to role setup = %v", tt.name, errAddUser)
				return
			}

			// Call the function to test.
			err := tt.connector.RemoveDatabaseRoleMember(ctx, db, tt.databaseRole, tt.user)

			errRoleCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.databaseRole)
			errUserCleanup := tt.connector.DeleteUser(ctx, db, tt.user)

			if errRoleCleanup != nil {
				t.Errorf("Test case %s: error during role cleanup = %v", tt.name, errRoleCleanup)
			}

			if errUserCleanup != nil {
				t.Errorf("Test case %s: error during role cleanup = %v", tt.name, errUserCleanup)
			}

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.RemoveDatabaseRoleMember() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errRoleCleanup != nil || errUserCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}
		})
	}
}

// TestConnector_GetDatabaseRoleMembers is a test function for the GetDatabaseRoleMembers method of the Connector struct.
func TestConnector_GetDatabaseRoleMembers(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string        // Name of the test case.
		connector        *Connector    // Connector instance to be tested.
		databaseOverride string        // Override for the database, if any.
		databaseRole     *model.Role   // Database role for the test case.
		users            []*model.User // Users to add to the database role.
		wantErr          bool          // Expected error condition.
	}{
		{
			name:             "db-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			databaseRole: &model.Role{
				Name: generateRandomString(10),
			},
			users: []*model.User{
				{
					Name:      generateRandomString(10),
					Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
					Contained: true,
				},
				{
					Name:      generateRandomString(10),
					Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
					Contained: true,
				},
				{
					Name:      generateRandomString(10),
					Password:  fmt.Sprintf("%s1aA!", generateRandomString(16)),
					Contained: true,
				},
			},
			wantErr: false,
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
			db, _ := tt.connector.Connect()

			// Setup the role and users
			errRoleCreate := tt.connector.CreateDatabaseRole(ctx, db, tt.databaseRole)
			if errRoleCreate != nil {
				t.Errorf("Test case %s: error during role setup = %v", tt.name, errRoleCreate)
				return
			}

			for _, user := range tt.users {
				errUserCreate := tt.connector.CreateUser(ctx, db, user)
				if errUserCreate != nil {
					t.Errorf("Test case %s: error during user setup = %v", tt.name, errUserCreate)
					return
				}
			}

			errAddUsers := tt.connector.AddDatabaseRoleMembers(ctx, db, tt.databaseRole, tt.users)
			if errAddUsers != nil {
				t.Errorf("Test case %s: error during adding users to role setup = %v", tt.name, errAddUsers)
				return
			}

			// Call the function to test.
			gotDatabaseRoleMembers, err := tt.connector.GetDatabaseRoleMembers(ctx, db, tt.databaseRole)

			// Cleanup the database role and users
			errRemoveUsers := tt.connector.RemoveDatabaseRoleMembers(ctx, db, tt.databaseRole, tt.users)
			if errRemoveUsers != nil {
				t.Errorf("Test case %s: error during removing users from role setup = %v", tt.name, errRemoveUsers)
				return
			}

			for _, user := range tt.users {
				errUserCleanup := tt.connector.DeleteUser(ctx, db, user)
				if errUserCleanup != nil {
					t.Errorf("Test case %s: error during user cleanup = %v", tt.name, errUserCleanup)
					return
				}
			}

			errRoleCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.databaseRole)
			if errRoleCleanup != nil {
				t.Errorf("Test case %s: error during role cleanup = %v", tt.name, errRoleCleanup)
				return
			}

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetDatabaseRoleMembers() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			// Check the users returned by the function is the same as the users added to the role.
			for _, user := range tt.users {
				found := false
				for _, gotUser := range gotDatabaseRoleMembers {
					if user.Name == gotUser.Name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Test case %s: user %s not found in the role members", tt.name, user.Name)
				}
			}
		})
	}
}
