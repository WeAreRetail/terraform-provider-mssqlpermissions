//go:build integration

package queries

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

// ============================================================================
// VALIDATION FUNCTION TESTS
// ============================================================================

// TestValidateRoleName tests the validateRoleName function
func TestValidateRoleName(t *testing.T) {
	tests := []struct {
		name    string
		role    *model.Role
		wantErr bool
	}{
		{
			name:    "valid-role-name",
			role:    &model.Role{Name: "TestRole"},
			wantErr: false,
		},
		{
			name:    "nil-role",
			role:    nil,
			wantErr: true,
		},
		{
			name:    "empty-role-name",
			role:    &model.Role{Name: ""},
			wantErr: true,
		},
		{
			name:    "invalid-role-name-with-special-chars",
			role:    &model.Role{Name: "Test@Role"},
			wantErr: true,
		},
		{
			name:    "role-name-too-long",
			role:    &model.Role{Name: string(make([]byte, 130))}, // Exceeds 128 char limit
			wantErr: true,
		},
		{
			name:    "role-name-with-numbers",
			role:    &model.Role{Name: "TestRole123"},
			wantErr: false,
		},
		{
			name:    "role-name-with-underscores",
			role:    &model.Role{Name: "Test_Role_123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoleName(tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoleName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidatePermissionName tests the validatePermissionName function
func TestValidatePermissionName(t *testing.T) {
	tests := []struct {
		name       string
		permission *model.Permission
		wantErr    bool
	}{
		{
			name:       "valid-permission-name",
			permission: &model.Permission{Name: "SELECT"},
			wantErr:    false,
		},
		{
			name:       "nil-permission",
			permission: nil,
			wantErr:    true,
		},
		{
			name:       "empty-permission-name",
			permission: &model.Permission{Name: ""},
			wantErr:    true,
		},
		{
			name:       "invalid-permission-name-with-special-chars",
			permission: &model.Permission{Name: "SELECT@TABLE"},
			wantErr:    true,
		},
		{
			name:       "permission-name-too-long",
			permission: &model.Permission{Name: string(make([]byte, 130))}, // Exceeds 128 char limit
			wantErr:    true,
		},
		{
			name:       "permission-name-with-underscores",
			permission: &model.Permission{Name: "VIEW_DEFINITION"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePermissionName(tt.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePermissionName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidatePermissionState tests the validatePermissionState function
func TestValidatePermissionState(t *testing.T) {
	tests := []struct {
		name       string
		permission *model.Permission
		want       string
		wantErr    bool
	}{
		{
			name:       "valid-grant-state",
			permission: &model.Permission{State: "G"},
			want:       "GRANT",
			wantErr:    false,
		},
		{
			name:       "valid-deny-state",
			permission: &model.Permission{State: "D"},
			want:       "DENY",
			wantErr:    false,
		},
		{
			name:       "empty-state-defaults-to-grant",
			permission: &model.Permission{State: ""},
			want:       "GRANT",
			wantErr:    false,
		},
		{
			name:       "valid-grant-state-desc",
			permission: &model.Permission{StateDesc: "GRANT"},
			want:       "GRANT",
			wantErr:    false,
		},
		{
			name:       "valid-deny-state-desc",
			permission: &model.Permission{StateDesc: "DENY"},
			want:       "DENY",
			wantErr:    false,
		},
		{
			name:       "empty-state-desc-defaults-to-grant",
			permission: &model.Permission{StateDesc: ""},
			want:       "GRANT",
			wantErr:    false,
		},
		{
			name:       "deny-state-overrides-grant-state-desc",
			permission: &model.Permission{State: "D", StateDesc: "GRANT"},
			want:       "DENY",
			wantErr:    false,
		},
		{
			name:       "invalid-state",
			permission: &model.Permission{State: "X"},
			want:       "",
			wantErr:    true,
		},
		{
			name:       "invalid-state-desc",
			permission: &model.Permission{StateDesc: "INVALID"},
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePermissionState(tt.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePermissionState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validatePermissionState() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateSQLIdentifier tests the validateSQLIdentifier function
func TestValidateSQLIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantErr    bool
	}{
		{
			name:       "valid-identifier",
			identifier: "ValidName",
			wantErr:    false,
		},
		{
			name:       "valid-identifier-with-underscore",
			identifier: "_ValidName",
			wantErr:    false,
		},
		{
			name:       "valid-identifier-with-numbers",
			identifier: "ValidName123",
			wantErr:    false,
		},
		{
			name:       "empty-identifier",
			identifier: "",
			wantErr:    true,
		},
		{
			name:       "identifier-too-long",
			identifier: string(make([]byte, 130)), // Exceeds 128 char limit
			wantErr:    true,
		},
		{
			name:       "identifier-starts-with-number",
			identifier: "123Invalid",
			wantErr:    true,
		},
		{
			name:       "identifier-with-special-chars",
			identifier: "Invalid@Name",
			wantErr:    true,
		},
		{
			name:       "identifier-with-spaces",
			identifier: "Invalid Name",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSQLIdentifier(tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSQLIdentifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateSchemaName tests the validateSchemaName function
func TestValidateSchemaName(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		wantErr    bool
	}{
		{
			name:       "valid-schema-name",
			schemaName: "dbo",
			wantErr:    false,
		},
		{
			name:       "valid-schema-name-with-underscore",
			schemaName: "test_schema",
			wantErr:    false,
		},
		{
			name:       "empty-schema-name",
			schemaName: "",
			wantErr:    true,
		},
		{
			name:       "invalid-schema-name-with-special-chars",
			schemaName: "test@schema",
			wantErr:    true,
		},
		{
			name:       "schema-name-too-long",
			schemaName: string(make([]byte, 130)), // Exceeds 128 char limit
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaName(tt.schemaName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSchemaName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

// TestCreateTestPermission tests the CreateTestPermission helper function
func TestCreateTestPermission(t *testing.T) {
	tests := []struct {
		name      string
		permName  string
		state     string
		wantName  string
		wantState string
	}{
		{
			name:      "create-select-permission-with-grant",
			permName:  "SELECT",
			state:     "G",
			wantName:  "SELECT",
			wantState: "G",
		},
		{
			name:      "create-insert-permission-with-deny",
			permName:  "INSERT",
			state:     "D",
			wantName:  "INSERT",
			wantState: "D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateTestPermission(tt.permName, tt.state)
			if got.Name != tt.wantName {
				t.Errorf("CreateTestPermission().Name = %v, want %v", got.Name, tt.wantName)
			}
			if got.State != tt.wantState {
				t.Errorf("CreateTestPermission().State = %v, want %v", got.State, tt.wantState)
			}
		})
	}
}

// TestCreateTestRole tests the CreateTestRole helper function
func TestCreateTestRole(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		wantName string
	}{
		{
			name:     "create-test-role",
			roleName: "TestRole",
			wantName: "TestRole",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateTestRole(tt.roleName)
			if got.Name != tt.wantName {
				t.Errorf("CreateTestRole().Name = %v, want %v", got.Name, tt.wantName)
			}
		})
	}
}

// TestCreateTestPermissionWithStateDesc tests the CreateTestPermissionWithStateDesc helper function
func TestCreateTestPermissionWithStateDesc(t *testing.T) {
	tests := []struct {
		name          string
		permName      string
		stateDesc     string
		wantName      string
		wantStateDesc string
	}{
		{
			name:          "create-permission-with-grant-desc",
			permName:      "SELECT",
			stateDesc:     "GRANT",
			wantName:      "SELECT",
			wantStateDesc: "GRANT",
		},
		{
			name:          "create-permission-with-deny-desc",
			permName:      "INSERT",
			stateDesc:     "DENY",
			wantName:      "INSERT",
			wantStateDesc: "DENY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateTestPermissionWithStateDesc(tt.permName, tt.stateDesc)
			if got.Name != tt.wantName {
				t.Errorf("CreateTestPermissionWithStateDesc().Name = %v, want %v", got.Name, tt.wantName)
			}
			if got.StateDesc != tt.wantStateDesc {
				t.Errorf("CreateTestPermissionWithStateDesc().StateDesc = %v, want %v", got.StateDesc, tt.wantStateDesc)
			}
		})
	}
}

// ============================================================================
// DATABASE PERMISSION ASSIGNMENT TESTS
// ============================================================================

// TestConnector_GrantPermissionToRole tests the GrantPermissionToRole method
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
			name:             "grant-select-permission-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "SELECT",
			},
			wantErr: false,
		},
		{
			name:             "grant-insert-permission-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "INSERT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, tt.permission)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: GrantPermissionToRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}

// TestConnector_DenyPermissionToRole tests the DenyPermissionToRole method
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
			name:             "deny-select-permission-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "SELECT",
			},
			wantErr: false,
		},
		{
			name:             "deny-insert-permission-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "INSERT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.DenyPermissionToRole(ctx, db, tt.role, tt.permission)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: DenyPermissionToRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}

// TestConnector_RevokePermissionFromRole tests the RevokePermissionFromRole method
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
			name:             "revoke-select-permission-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "SELECT",
			},
			wantErr: false,
		},
		{
			name:             "revoke-insert-permission-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permission: &model.Permission{
				Name: "INSERT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// First grant the permission so we have something to revoke
			err = tt.connector.GrantPermissionToRole(ctx, db, tt.role, tt.permission)
			if err != nil {
				t.Errorf("Test case %s: error during setup (granting permission) = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.RevokePermissionFromRole(ctx, db, tt.role, tt.permission)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: RevokePermissionFromRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}

// TestConnector_GrantPermissionsToRole tests the GrantPermissionsToRole method
func TestConnector_GrantPermissionsToRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "grant-multiple-permissions-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{Name: "SELECT"},
				{Name: "INSERT"},
				{Name: "UPDATE"},
			},
			wantErr: false,
		},
		{
			name:             "grant-multiple-permissions-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{Name: "SELECT"},
				{Name: "INSERT"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.GrantPermissionsToRole(ctx, db, tt.role, tt.permissions)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: GrantPermissionsToRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}

// ============================================================================
// DATABASE PERMISSION QUERY TESTS (EXISTING)
// ============================================================================

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

// ============================================================================
// SCHEMA-LEVEL PERMISSION TESTS
// ============================================================================

// TestConnector_GrantPermissionOnSchemaToRole tests the GrantPermissionOnSchemaToRole method
func TestConnector_GrantPermissionOnSchemaToRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		schema           string
		permission       *model.Permission
		wantErr          bool
	}{
		{
			name:             "grant-select-on-schema-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			schema: "dbo",
			permission: &model.Permission{
				Name: "SELECT",
			},
			wantErr: false,
		},
		{
			name:             "grant-insert-on-schema-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			schema: "dbo",
			permission: &model.Permission{
				Name: "INSERT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.GrantPermissionOnSchemaToRole(ctx, db, tt.role, tt.schema, tt.permission)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: GrantPermissionOnSchemaToRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}

// ============================================================================
// ADDITIONAL PERMISSION TESTS
// ============================================================================// TestConnector_GrantPermissionsToRoleWithTransaction tests the GrantPermissionsToRoleWithTransaction method
func TestConnector_GrantPermissionsToRoleWithTransaction(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "grant-multiple-permissions-with-transaction-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{Name: "SELECT", State: "GRANT"},
				{Name: "INSERT", State: "GRANT"},
				{Name: "UPDATE", State: "GRANT"},
			},
			wantErr: false,
		},
		{
			name:             "grant-multiple-permissions-with-transaction-on-azureSQL",
			connector:        testConnectors.azureSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{Name: "SELECT", State: "GRANT"},
				{Name: "DELETE", State: "GRANT"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.GrantPermissionsToRoleWithTransaction(ctx, db, tt.role, tt.permissions)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: GrantPermissionsToRoleWithTransaction() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}

// TestConnector_GetSchemaPermissionsForRole tests the GetSchemaPermissionsForRole method
func TestConnector_GetSchemaPermissionsForRole(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		schema           string
		wantErr          bool
	}{
		{
			name:             "get-schema-permissions-for-role-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: "db_datareader",
			},
			schema:  "dbo",
			wantErr: false,
		},
		{
			name:             "get-schema-permissions-for-nonexistent-role",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: "NonexistentRole123",
			},
			schema:  "dbo",
			wantErr: false, // Should not error, just return empty list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Test the function
			permissions, err := tt.connector.GetSchemaPermissionsForRole(ctx, db, tt.role, tt.schema)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: GetSchemaPermissionsForRole() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if err == nil {
				// Just verify we got a result (could be empty for non-existent roles)
				t.Logf("Test case %s: Retrieved %d schema permissions", tt.name, len(permissions))
			}
		})
	}
}

// TestConnector_DenyPermissionsToRoleWithTransaction tests the DenyPermissionsToRoleWithTransaction method
func TestConnector_DenyPermissionsToRoleWithTransaction(t *testing.T) {
	tests := []struct {
		name             string
		connector        *Connector
		databaseOverride string
		role             *model.Role
		permissions      []*model.Permission
		wantErr          bool
	}{
		{
			name:             "deny-multiple-permissions-with-transaction-on-LocalSQL",
			connector:        testConnectors.localSQL,
			databaseOverride: "ApplicationDB",
			role: &model.Role{
				Name: generateRandomString(10),
			},
			permissions: []*model.Permission{
				{Name: "SELECT", State: "DENY"},
				{Name: "INSERT", State: "DENY"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbRestore := tt.connector.Database

			// Override database if specified.
			if tt.databaseOverride != "" {
				tt.connector.Database = tt.databaseOverride
			}

			ctx := context.Background()
			db, err := tt.connector.Connect()
			if err != nil {
				t.Errorf("Test case %s: failed to connect = %v", tt.name, err)
				return
			}

			// Create the database role
			err = tt.connector.CreateDatabaseRole(ctx, db, tt.role)
			if err != nil {
				t.Errorf("Test case %s: error during role creation = %v", tt.name, err)
				return
			}

			// Test the function
			err = tt.connector.DenyPermissionsToRoleWithTransaction(ctx, db, tt.role, tt.permissions)

			// Cleanup the database role
			errCleanup := tt.connector.DeleteDatabaseRole(ctx, db, tt.role)

			// Restore the original database value
			tt.connector.Database = dbRestore

			// Check if the error condition matches the expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %s: DenyPermissionsToRoleWithTransaction() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			} else if err == nil {
				// Check if the cleanup was successful
				if errCleanup != nil {
					t.Errorf("Test case %s: error during cleanup = %v", tt.name, errCleanup)
				}
			}
		})
	}
}
