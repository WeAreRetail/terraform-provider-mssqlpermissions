package queries

import (
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

// ============================================================================
// PURE UNIT TESTS - No database dependencies
// ============================================================================

// TestValidateRoleName_Unit tests the validateRoleName function with comprehensive edge cases
func TestValidateRoleName_Unit(t *testing.T) {
	tests := []struct {
		name    string
		role    *model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_role_name",
			role:    &model.Role{Name: "TestRole"},
			wantErr: false,
		},
		{
			name:    "valid_role_with_underscore",
			role:    &model.Role{Name: "Test_Role"},
			wantErr: false,
		},
		{
			name:    "valid_role_with_numbers",
			role:    &model.Role{Name: "TestRole123"},
			wantErr: false,
		},
		{
			name:    "valid_role_mixed_case",
			role:    &model.Role{Name: "TestRole_123_ABC"},
			wantErr: false,
		},
		{
			name:    "nil_role",
			role:    nil,
			wantErr: true,
			errMsg:  "role name cannot be empty",
		},
		{
			name:    "empty_role_name",
			role:    &model.Role{Name: ""},
			wantErr: true,
			errMsg:  "role name cannot be empty",
		},
		{
			name:    "role_name_with_spaces",
			role:    &model.Role{Name: "Test Role"},
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "role_name_with_special_chars",
			role:    &model.Role{Name: "Test@Role"},
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "role_name_starts_with_number",
			role:    &model.Role{Name: "123TestRole"},
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "role_name_too_long",
			role:    &model.Role{Name: string(make([]byte, 129))}, // Exceeds 128 char limit
			wantErr: true,
			errMsg:  "SQL identifier too long",
		},
		{
			name:    "role_name_max_length",
			role:    &model.Role{Name: generateValidIdentifier(128)}, // Exactly at limit with valid chars
			wantErr: false,
		},
		{
			name:    "role_name_with_hyphen",
			role:    &model.Role{Name: "Test-Role"},
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "role_name_with_dot",
			role:    &model.Role{Name: "Test.Role"},
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "role_name_single_char",
			role:    &model.Role{Name: "A"},
			wantErr: false,
		},
		{
			name:    "role_name_underscore_start",
			role:    &model.Role{Name: "_TestRole"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoleName(tt.role)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateRoleName() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateRoleName() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateRoleName() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidatePermissionName_Unit tests the validatePermissionName function with comprehensive edge cases
func TestValidatePermissionName_Unit(t *testing.T) {
	tests := []struct {
		name       string
		permission *model.Permission
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid_permission_select",
			permission: &model.Permission{Name: "SELECT"},
			wantErr:    false,
		},
		{
			name:       "valid_permission_insert",
			permission: &model.Permission{Name: "INSERT"},
			wantErr:    false,
		},
		{
			name:       "valid_permission_mixed_case",
			permission: &model.Permission{Name: "Create_Table"},
			wantErr:    false,
		},
		{
			name:       "valid_permission_with_underscore",
			permission: &model.Permission{Name: "ALTER_SCHEMA"},
			wantErr:    false,
		},
		{
			name:       "nil_permission",
			permission: nil,
			wantErr:    true,
			errMsg:     "permission name cannot be empty",
		},
		{
			name:       "empty_permission_name",
			permission: &model.Permission{Name: ""},
			wantErr:    true,
			errMsg:     "permission name cannot be empty",
		},
		{
			name:       "permission_with_spaces",
			permission: &model.Permission{Name: "SELECT ALL"},
			wantErr:    true,
			errMsg:     "invalid SQL identifier format",
		},
		{
			name:       "permission_with_special_chars",
			permission: &model.Permission{Name: "SELECT@TABLE"},
			wantErr:    true,
			errMsg:     "invalid SQL identifier format",
		},
		{
			name:       "permission_starts_with_number",
			permission: &model.Permission{Name: "123SELECT"},
			wantErr:    true,
			errMsg:     "invalid SQL identifier format",
		},
		{
			name:       "permission_name_too_long",
			permission: &model.Permission{Name: string(make([]byte, 129))}, // Exceeds 128 char limit
			wantErr:    true,
			errMsg:     "SQL identifier too long",
		},
		{
			name:       "permission_name_max_length",
			permission: &model.Permission{Name: generateValidIdentifier(128)}, // Exactly at limit with valid chars
			wantErr:    false,
		},
		{
			name:       "permission_single_char",
			permission: &model.Permission{Name: "X"},
			wantErr:    false,
		},
		{
			name:       "permission_underscore_start",
			permission: &model.Permission{Name: "_EXECUTE"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePermissionName(tt.permission)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePermissionName() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validatePermissionName() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validatePermissionName() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidatePermissionState_Unit tests the validatePermissionState function
func TestValidatePermissionState_Unit(t *testing.T) {
	tests := []struct {
		name       string
		permission *model.Permission
		wantVerb   string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "grant_state",
			permission: &model.Permission{State: "G"},
			wantVerb:   "GRANT",
			wantErr:    false,
		},
		{
			name:       "deny_state",
			permission: &model.Permission{State: "D"},
			wantVerb:   "DENY",
			wantErr:    false,
		},
		{
			name:       "grant_state_desc",
			permission: &model.Permission{StateDesc: "GRANT"},
			wantVerb:   "GRANT",
			wantErr:    false,
		},
		{
			name:       "deny_state_desc",
			permission: &model.Permission{StateDesc: "DENY"},
			wantVerb:   "DENY",
			wantErr:    false,
		},
		{
			name:       "empty_state_defaults_to_grant",
			permission: &model.Permission{},
			wantVerb:   "GRANT",
			wantErr:    false,
		},
		{
			name:       "state_overrides_state_desc",
			permission: &model.Permission{State: "D", StateDesc: "GRANT"},
			wantVerb:   "DENY",
			wantErr:    false,
		},
		{
			name:       "invalid_state",
			permission: &model.Permission{State: "X"},
			wantVerb:   "",
			wantErr:    true,
			errMsg:     "invalid state value",
		},
		{
			name:       "invalid_state_desc",
			permission: &model.Permission{StateDesc: "INVALID"},
			wantVerb:   "",
			wantErr:    true,
			errMsg:     "invalid state value",
		},
		{
			name:       "lowercase_state_desc",
			permission: &model.Permission{StateDesc: "grant"},
			wantVerb:   "",
			wantErr:    true,
			errMsg:     "invalid state value",
		},
		{
			name:       "lowercase_deny_state_desc",
			permission: &model.Permission{StateDesc: "deny"},
			wantVerb:   "",
			wantErr:    true,
			errMsg:     "invalid state value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verb, err := validatePermissionState(tt.permission)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePermissionState() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validatePermissionState() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validatePermissionState() unexpected error = %v", err)
				}
				if verb != tt.wantVerb {
					t.Errorf("validatePermissionState() verb = %v, want %v", verb, tt.wantVerb)
				}
			}
		})
	}
}

// TestValidateSQLIdentifier_Unit tests the validateSQLIdentifier function
func TestValidateSQLIdentifier_Unit(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_identifier",
			id:      "TestIdentifier",
			wantErr: false,
		},
		{
			name:    "valid_identifier_underscore",
			id:      "Test_Identifier",
			wantErr: false,
		},
		{
			name:    "valid_identifier_numbers",
			id:      "TestId123",
			wantErr: false,
		},
		{
			name:    "valid_starts_with_underscore",
			id:      "_TestIdentifier",
			wantErr: false,
		},
		{
			name:    "empty_identifier",
			id:      "",
			wantErr: true,
			errMsg:  "SQL identifier cannot be empty",
		},
		{
			name:    "identifier_too_long",
			id:      string(make([]byte, 129)), // Exceeds 128 char limit
			wantErr: true,
			errMsg:  "SQL identifier too long",
		},
		{
			name:    "identifier_starts_with_number",
			id:      "123Test",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "identifier_with_spaces",
			id:      "Test Identifier",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "identifier_with_special_chars",
			id:      "Test@Identifier",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "identifier_with_hyphen",
			id:      "Test-Identifier",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "identifier_single_char_letter",
			id:      "A",
			wantErr: false,
		},
		{
			name:    "identifier_single_char_underscore",
			id:      "_",
			wantErr: false,
		},
		{
			name:    "identifier_max_length",
			id:      generateValidIdentifier(128), // Exactly at limit with valid chars
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSQLIdentifier(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateSQLIdentifier() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateSQLIdentifier() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateSQLIdentifier() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateSchemaName_Unit tests the validateSchemaName function
func TestValidateSchemaName_Unit(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_schema_name",
			schema:  "dbo",
			wantErr: false,
		},
		{
			name:    "valid_schema_with_underscore",
			schema:  "test_schema",
			wantErr: false,
		},
		{
			name:    "valid_schema_mixed_case",
			schema:  "TestSchema123",
			wantErr: false,
		},
		{
			name:    "empty_schema_name",
			schema:  "",
			wantErr: true,
			errMsg:  "schema name cannot be empty",
		},
		{
			name:    "schema_name_too_long",
			schema:  string(make([]byte, 129)), // Exceeds 128 char limit
			wantErr: true,
			errMsg:  "SQL identifier too long",
		},
		{
			name:    "schema_with_spaces",
			schema:  "test schema",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "schema_with_special_chars",
			schema:  "test@schema",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "schema_starts_with_number",
			schema:  "123schema",
			wantErr: true,
			errMsg:  "invalid SQL identifier format",
		},
		{
			name:    "schema_starts_with_underscore",
			schema:  "_schema",
			wantErr: false,
		},
		{
			name:    "schema_single_char",
			schema:  "s",
			wantErr: false,
		},
		{
			name:    "schema_max_length",
			schema:  generateValidIdentifier(128), // Exactly at limit with valid chars
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaName(tt.schema)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateSchemaName() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateSchemaName() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateSchemaName() unexpected error = %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to generate a valid SQL identifier of specified length
func generateValidIdentifier(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	// Start with a letter
	result[0] = 'A'

	// Fill rest with valid characters
	for i := 1; i < length; i++ {
		// Cycle through A-Z, a-z, 0-9, _
		switch i % 63 {
		case 0:
			result[i] = 'A'
		case 26:
			result[i] = 'a'
		case 52:
			result[i] = '0'
		case 62:
			result[i] = '_'
		default:
			if i%63 < 26 {
				result[i] = byte('A' + (i % 63))
			} else if i%63 < 52 {
				result[i] = byte('a' + (i%63 - 26))
			} else {
				result[i] = byte('0' + (i%63 - 52))
			}
		}
	}

	return string(result)
}
