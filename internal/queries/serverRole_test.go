package queries

import (
	"context"
	"fmt"
	"reflect"
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

// TestConnector_GetServerRole is a test function for the GetServerRole method of the Connector struct.
func TestConnector_GetServerRole(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		serverRole       *model.Role // Server role for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:             "sysadmin-exists-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: "sysadmin",
			},
			wantErr: false,
		},
		{
			name:             "##MS_LoginManager##-exists-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: "##MS_LoginManager##",
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
			gotServerRole, err := tt.connector.GetServerRole(ctx, db, tt.serverRole)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetServerRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			// Check the type of PrincipalID in the returned serverRole.
			principalIDType := reflect.TypeOf(gotServerRole.PrincipalID)
			if principalIDType != reflect.TypeOf(int64(0)) {
				t.Errorf("Test case %s: Invalid type for PrincipalID is %s while expecting int64", tt.name, principalIDType)
			}
			// Uncomment the following line to log the type of PrincipalID.
			// t.Logf("principalIDType: %s", principalIDType)
		})
	}
}

func TestConnector_CreateServerRole(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		serverRole       *model.Role // Server role for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:             "testers-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: false,
		},
		{
			name:             "testers-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: true,
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
			err := tt.connector.CreateServerRole(ctx, db, tt.serverRole)
			errCleanup := tt.connector.DeleteServerRole(ctx, db, tt.serverRole)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.CreateServerRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
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

func TestConnector_DeleteServerRole(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		serverRole       *model.Role // Server role for the test case.
		wantErr          bool        // Expected error condition.
	}{
		{
			name:             "testers-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: "HZllg7MWHm",
			},
			wantErr: false,
		},
		{
			name:             "testers-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: generateRandomString(10),
			},
			wantErr: true,
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

			// Create the server role to delete
			err := tt.connector.CreateServerRole(ctx, db, tt.serverRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			}

			err = tt.connector.DeleteServerRole(ctx, db, tt.serverRole)

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.DeleteServerRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
		})
	}
}

func TestConnector_AddServerRoleMember(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		serverRole       *model.Role // Server role for the test case.
		login            *model.Login
		wantErr          bool // Expected error condition.
	}{
		{
			name:             "add-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: generateRandomString(10),
			},
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
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

			// Setup the server role
			errRoleCreate := tt.connector.CreateServerRole(ctx, db, tt.serverRole)
			if errRoleCreate != nil {
				t.Errorf("Test case %s: error during setup, cannot create role = %v", tt.name, errRoleCreate)
			}

			// Setup the login
			errLoginCreate := tt.connector.CreateLogin(ctx, db, tt.login)
			if errLoginCreate != nil {
				t.Errorf("Test case %s: error during setup, cannot create login = %v", tt.name, errLoginCreate)
			}

			// Call the function to test
			err := tt.connector.AddServerRoleMember(ctx, db, tt.serverRole, tt.login)

			// Cleanup
			errRoleDelete := tt.connector.DeleteServerRole(ctx, db, tt.serverRole)
			errLoginDelete := tt.connector.DeleteLogin(ctx, db, tt.login)

			if errRoleDelete != nil {
				t.Errorf("Test case %s: error during cleanup, cannot delete role = %v", tt.name, errRoleDelete)
			}

			if errLoginDelete != nil {
				t.Errorf("Test case %s: error during cleanup, cannot delete login = %v", tt.name, errLoginDelete)
			}

			// Restore the original database value.
			tt.connector.Database = dbRestore

			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.AddServerRoleMember() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errRoleDelete != nil || errLoginDelete != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}
		})
	}
}

func TestConnector_RemoveServerRoleMember(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string      // Name of the test case.
		connector        *Connector  // Connector instance to be tested.
		databaseOverride string      // Override for the database, if any.
		serverRole       *model.Role // Server role for the test case.
		login            *model.Login
		wantErr          bool // Expected error condition.
	}{
		{
			name:             "add-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: generateRandomString(10),
			},
			login: &model.Login{
				Name:     generateRandomString(10),
				Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
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

			// Setup the server role
			errRoleCreate := tt.connector.CreateServerRole(ctx, db, tt.serverRole)
			if errRoleCreate != nil {
				t.Errorf("Test case %s: error during setup, cannot create role = %v", tt.name, errRoleCreate)
			}

			// Setup the login
			errLoginCreate := tt.connector.CreateLogin(ctx, db, tt.login)
			if errLoginCreate != nil {
				t.Errorf("Test case %s: error during setup, cannot create login = %v", tt.name, errLoginCreate)
			}

			// Setup the server role member
			errRoleMemberAdd := tt.connector.AddServerRoleMember(ctx, db, tt.serverRole, tt.login)
			if errRoleMemberAdd != nil {
				t.Errorf("Test case %s: error during setup, cannot add server role member = %v", tt.name, errRoleMemberAdd)
			}

			// Call the function to test
			err := tt.connector.RemoveServerRoleMember(ctx, db, tt.serverRole, tt.login)

			// Cleanup
			errRoleDelete := tt.connector.DeleteServerRole(ctx, db, tt.serverRole)
			errLoginDelete := tt.connector.DeleteLogin(ctx, db, tt.login)

			if errRoleDelete != nil {
				t.Errorf("Test case %s: error during cleanup, cannot delete role = %v", tt.name, errRoleDelete)
			}

			if errLoginDelete != nil {
				t.Errorf("Test case %s: error during cleanup, cannot delete login = %v", tt.name, errLoginDelete)
			}

			// Restore the original database value.
			tt.connector.Database = dbRestore

			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.AddServerRoleMember() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful.
				if errRoleDelete != nil || errLoginDelete != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
					return
				}
			}
		})
	}
}

// TestConnector_GetServerRoleMembers is a test function for the GetServerRoleMembers method of the Connector struct.
func TestConnector_GetServerRoleMembers(t *testing.T) {
	// Define test cases with different scenarios.
	tests := []struct {
		name             string         // Name of the test case.
		connector        *Connector     // Connector instance to be tested.
		databaseOverride string         // Override for the database, if any.
		serverRole       *model.Role    // Database role for the test case.
		logins           []*model.Login // Logins to add to the database role.
		wantErr          bool           // Expected error condition.
	}{
		{
			name:             "db-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			serverRole: &model.Role{
				Name: generateRandomString(10),
			},
			logins: []*model.Login{
				{
					Name:     generateRandomString(10),
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				},
				{
					Name:     generateRandomString(10),
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
				},
				{
					Name:     generateRandomString(10),
					Password: fmt.Sprintf("%s1aA!", generateRandomString(16)),
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

			// Setup the role and logins
			errRoleCreate := tt.connector.CreateServerRole(ctx, db, tt.serverRole)
			if errRoleCreate != nil {
				t.Errorf("Test case %s: error during role setup = %v", tt.name, errRoleCreate)
				return
			}

			for _, login := range tt.logins {
				errLoginCreate := tt.connector.CreateLogin(ctx, db, login)
				if errLoginCreate != nil {
					t.Errorf("Test case %s: error during login setup = %v", tt.name, errLoginCreate)
					return
				}
			}

			errAddLogins := tt.connector.AddServerRoleMembers(ctx, db, tt.serverRole, tt.logins)
			if errAddLogins != nil {
				t.Errorf("Test case %s: error during adding logins to role setup = %v", tt.name, errAddLogins)
				return
			}

			// Call the function to test.
			gotServerRoleMembers, err := tt.connector.GetServerRoleMembers(ctx, db, tt.serverRole)

			// Cleanup the database role and logins
			errRemoveLogins := tt.connector.RemoveServerRoleMembers(ctx, db, tt.serverRole, tt.logins)
			if errRemoveLogins != nil {
				t.Errorf("Test case %s: error during removing logins from role setup = %v", tt.name, errRemoveLogins)
				return
			}

			for _, login := range tt.logins {
				errLoginCleanup := tt.connector.DeleteLogin(ctx, db, login)
				if errLoginCleanup != nil {
					t.Errorf("Test case %s: error during login cleanup = %v", tt.name, errLoginCleanup)
					return
				}
			}

			errRoleCleanup := tt.connector.DeleteServerRole(ctx, db, tt.serverRole)
			if errRoleCleanup != nil {
				t.Errorf("Test case %s: error during role cleanup = %v", tt.name, errRoleCleanup)
				return
			}

			// Restore the original database value.
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: Connector.GetServerRoleMembers() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			// Check the logins returned by the function is the same as the logins added to the role.
			for _, login := range tt.logins {
				found := false
				for _, gotLogin := range gotServerRoleMembers {
					if login.Name == gotLogin.Name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Test case %s: login %s not found in the role members", tt.name, login.Name)
				}
			}
		})
	}
}
