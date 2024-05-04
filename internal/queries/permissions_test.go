package queries

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

// TestConnector_GrantPermissionToRole is a unit test function that tests the GrantPermissionToRole method of the Connector struct.
// It verifies the behavior of granting permissions to a role on a database or server.
func TestConnector_GrantPermissionToRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permission       *model.Permission
		wantErr          bool
	}{
		{
			name:             "grant-server-permission-to-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "VIEW SERVER STATE",
			},
			wantErr: false,
		},
		{
			name:             "grant-server-permission-to-role-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "VIEW SERVER STATE",
			},
			wantErr: true, // This test case will fail because we cannot create server roles on Azure SQL.
		},
	}

	for _, tt := range tests {

		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the server role to delete
		err := tt.connector.CreateServerRole(ctx, db, tt.role)

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Call the function to test.
		err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, tt.permission)

		// Cleanup the server role created for the test.
		errCleanup := tt.connector.DeleteServerRole(ctx, db, tt.role)

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.GrantPermissionToRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
				return
			}
		}
	}
} // TestConnector_GrantPermissionToRole is a unit test function that tests the GrantPermissionToRole method of the Connector struct.
// It verifies the behavior of granting permissions to a role on a database or server.
func TestConnector_DenyPermissionToRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permission       *model.Permission
		wantErr          bool
	}{
		{
			name:             "deny-server-permission-to-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "VIEW SERVER STATE",
			},
			wantErr: false,
		},
		{
			name:             "deny-server-permission-to-role-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "VIEW SERVER STATE",
			},
			wantErr: true, // This test case will fail because we cannot create server roles on Azure SQL.
		},
	}

	for _, tt := range tests {

		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the server role to delete
		err := tt.connector.CreateServerRole(ctx, db, tt.role)

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Call the function to test.
		err = tt.connector.DenyPermissionToRole(ctx, db, tt.role, tt.permission)

		// Cleanup the server role created for the test.
		errCleanup := tt.connector.DeleteServerRole(ctx, db, tt.role)

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.DenyPermissionToRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
				return
			}
		}
	}
}

// TestConnector_GetServerPermissionsForRole is a unit test function that tests the GetServerPermissionsForRole method of the Connector struct.
// It verifies the behavior of retrieving server permissions for a role.
func TestConnector_GetServerPermissionsForRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "get-server-permissions-for-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{
					Name: "VIEW SERVER STATE",
				},
				{
					Name: "VIEW ANY DEFINITION",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {

		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the server role to delete
		err := tt.connector.CreateServerRole(ctx, db, tt.role)

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Grant permissions to the role.
		for _, permission := range tt.permissions {
			err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, permission)
			if err != nil {
				t.Errorf("Test case %s: error during setup = %v", tt.name, err)
				return
			}
		}

		// Call the function to test.
		_, err = tt.connector.GetServerPermissionsForRole(ctx, db, tt.role)

		// Cleanup the server role created for the test.
		errCleanup := tt.connector.DeleteServerRole(ctx, db, tt.role)

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.GetServerPermissionsForRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
				return
			}
		}
	}
}

// TestConnector_GetServerPermissionsForRole is a unit test function that tests the GetServerPermissionsForRole method of the Connector struct.
// It verifies the behavior of retrieving server permissions for a role.
func TestConnector_GetServerPermissionForRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "get-server-permissions-for-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{
					Name: "VIEW SERVER STATE",
				},
				{
					Name: "VIEW ANY DEFINITION",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {

		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the server role to delete
		err := tt.connector.CreateServerRole(ctx, db, tt.role)

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Grant permissions to the role.
		for _, permission := range tt.permissions {
			err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, permission)
			if err != nil {
				t.Errorf("Test case %s: error during setup = %v", tt.name, err)
				return
			}

			// Call the function to test.
			_, err = tt.connector.GetServerPermissionForRole(ctx, db, tt.role, permission)
		}

		// Cleanup the server role created for the test.
		errCleanup := tt.connector.DeleteServerRole(ctx, db, tt.role)

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.GetServerPermissionsForRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
				return
			}
		}
	}
}

// TestConnector_GetDatabasePermissionsForRole is a unit test function that tests the GetDatabasePermissionsForRole method of the Connector struct.
// It verifies the behavior of retrieving database permissions for a role.
func TestConnector_GetDatabasePermissionsForRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "get-database-permissions-for-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{
					Name: "SELECT",
				},
				{
					Name: "INSERT",
				},
			},
			wantErr: false,
		},
		{
			name:             "get-database-permissions-for-role-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{
					Name: "SELECT",
				},
				{
					Name: "INSERT",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {

		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the database role to delete
		err := tt.connector.CreateDatabaseRole(ctx, db, tt.role)

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Grant permissions to the role.
		for _, permission := range tt.permissions {
			err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, permission)
			if err != nil {
				t.Errorf("Test case %s: error during setup = %v", tt.name, err)
				return
			}
		}

		// Call the function to test.
		_, err = tt.connector.GetDatabasePermissionsForRole(ctx, db, tt.role)

		// Cleanup the database role created for the test.
		errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.GetDatabasePermissionsForRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
				return
			}
		}
	}
}

// TestConnector_GetDatabasePermissionForRole is a unit test function that tests the GetDatabasePermissionsForRole method of the Connector struct.
// It verifies the behavior of retrieving database permissions for a role.
func TestConnector_GetDatabasePermissionForRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "get-database-permissions-for-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{
					Name: "SELECT",
				},
				{
					Name: "INSERT",
				},
			},
			wantErr: false,
		},
		{
			name:             "get-database-permissions-for-role-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{
					Name: "SELECT",
				},
				{
					Name: "INSERT",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {

		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the database role to delete
		err := tt.connector.CreateDatabaseRole(ctx, db, tt.role)

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Grant permissions to the role.
		for _, permission := range tt.permissions {
			err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, permission)
			if err != nil {
				t.Errorf("Test case %s: error during setup = %v", tt.name, err)
				return
			}

			// Call the function to test.
			var letSee *model.Permission
			letSee, err = tt.connector.GetDatabasePermissionForRole(ctx, db, tt.role, permission)
			t.Logf("Permission: %v", letSee.Type)
		}

		// Cleanup the database role created for the test.
		errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.GetDatabasePermissionsForRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, err)
				return
			}
		}
	}
}

// TestConnector_RevokePermissionFromRole is a unit test function that tests the RevokePermissionFromRole method of the Connector struct.
// It verifies the behavior of revoking permissions from a role on a database or server.
func TestConnector_RevokePermissionFromRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permission       *model.Permission
		wantErr          bool
	}{
		{
			name:             "revoke-server-permission-from-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "master",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "VIEW SERVER STATE",
			},
			wantErr: false,
		},
		{
			name:             "revoke-database-permission-from-role-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "SELECT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {

		var err error
		dbRestore := tt.connector.Database

		// Override database if specified.
		if tt.databaseOverride != "" {
			tt.connector.Database = tt.databaseOverride
		}

		ctx := context.Background()
		db, _ := tt.connector.Connect()

		// Create the server role to delete
		if tt.connector.Database == "master" {
			err = tt.connector.CreateServerRole(ctx, db, tt.role)
		} else {
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
		}

		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Grant permissions to the role.
		err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, tt.permission)
		if err != nil {
			t.Errorf("Test case %s: error during setup = %v", tt.name, err)
			return
		}

		// Call the function to test.
		err = tt.connector.RevokePermissionFromRole(ctx, db, tt.role, tt.permission)

		// Cleanup the server role created for the test.
		var errCleanup error
		if tt.connector.Database == "master" {
			errCleanup = tt.connector.DeleteServerRole(ctx, db, tt.role)
		} else {
			errCleanup = tt.connector.DeleteDatabaseRole(ctx, db, tt.role)
		}

		// Restore the original database value.
		tt.connector.Database = dbRestore

		// Check if the error condition matches the expectation.
		if (err != nil) != tt.wantErr {
			t.Errorf("Test case %s: Connector.RevokePermissionFromRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		} else if err == nil {
			// Check if the cleanup was successful.
			if errCleanup != nil {
				t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				return
			}
		}
	}
}
