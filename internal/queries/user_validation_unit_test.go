package queries

import (
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

// ============================================================================
// USER VALIDATION UNIT TESTS - Tests that require minimal setup
// ============================================================================

// TestValidateUser_Unit tests the validateUser function with different connector configurations
func TestValidateUser_Unit(t *testing.T) {
	// Create connectors for different environments
	localConnector := &Connector{
		isAzureDatabase: false,
	}
	azureConnector := &Connector{
		isAzureDatabase: true,
	}

	tests := []struct {
		name      string
		connector *Connector
		user      *model.User
		wantErr   bool
		errMsg    string
	}{
		// Valid user cases
		{
			name:      "valid_contained_user_local",
			connector: localConnector,
			user: &model.User{
				Name:            "testuser",
				Password:        "TestPassword123!",
				External:        false,
				DefaultLanguage: "",
				ObjectID:        "",
			},
			wantErr: false,
		},
		{
			name:      "valid_external_user_local",
			connector: localConnector,
			user: &model.User{
				Name:     "testuser@domain.com",
				External: true,
				ObjectID: "12345678-1234-1234-1234-123456789012",
			},
			wantErr: false,
		},
		{
			name:      "valid_user_with_default_language_local",
			connector: localConnector,
			user: &model.User{
				Name:            "testuser",
				Password:        "TestPassword123!",
				DefaultLanguage: "us_english",
			},
			wantErr: false,
		},
		{
			name:      "valid_contained_user_azure",
			connector: azureConnector,
			user: &model.User{
				Name:     "testuser",
				Password: "TestPassword123!",
				External: false,
			},
			wantErr: false,
		},
		{
			name:      "valid_external_user_azure",
			connector: azureConnector,
			user: &model.User{
				Name:     "testuser@domain.com",
				External: true,
				ObjectID: "12345678-1234-1234-1234-123456789012",
			},
			wantErr: false,
		},

		// Error cases - Missing name
		{
			name:      "empty_user_name",
			connector: localConnector,
			user: &model.User{
				Name:     "",
				Password: "TestPassword123!",
			},
			wantErr: true,
			errMsg:  "a user must have a name",
		},

		// Error cases - Password validation
		{
			name:      "contained_user_without_password",
			connector: localConnector,
			user: &model.User{
				Name:     "testuser",
				Password: "",
				External: false,
			},
			wantErr: true,
			errMsg:  "a contained user must have a password if it's not external",
		},
		{
			name:      "external_user_with_password",
			connector: localConnector,
			user: &model.User{
				Name:     "testuser@domain.com",
				Password: "TestPassword123!",
				External: true,
			},
			wantErr: true,
			errMsg:  "an external user cannot have a password",
		},

		// Error cases - ObjectID validation
		{
			name:      "contained_user_with_objectid",
			connector: localConnector,
			user: &model.User{
				Name:     "testuser",
				Password: "TestPassword123!",
				External: false,
				ObjectID: "12345678-1234-1234-1234-123456789012",
			},
			wantErr: true,
			errMsg:  "only external user can specify an ObjectID",
		},

		// Error cases - Default language validation
		{
			name:      "azure_user_with_default_language",
			connector: azureConnector,
			user: &model.User{
				Name:            "testuser",
				Password:        "TestPassword123!",
				DefaultLanguage: "us_english",
			},
			wantErr: true,
			errMsg:  "a user cannot have a default language in an Azure Database",
		},

		// Edge cases
		{
			name:      "external_user_without_objectid",
			connector: localConnector,
			user: &model.User{
				Name:     "testuser@domain.com",
				External: true,
				ObjectID: "",
			},
			wantErr: false, // ObjectID is optional for external users
		},
		{
			name:      "user_with_empty_default_language",
			connector: localConnector,
			user: &model.User{
				Name:            "testuser",
				Password:        "TestPassword123!",
				DefaultLanguage: "",
			},
			wantErr: false, // Empty default language should be OK
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.connector.validateUser(tt.user)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateUser() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateUser() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateUser() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestConnectorConfiguration_Unit tests connector field settings
func TestConnectorConfiguration_Unit(t *testing.T) {
	tests := []struct {
		name            string
		isAzureDatabase bool
		wantAzure       bool
	}{
		{
			name:            "local_connector",
			isAzureDatabase: false,
			wantAzure:       false,
		},
		{
			name:            "azure_connector",
			isAzureDatabase: true,
			wantAzure:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &Connector{
				isAzureDatabase: tt.isAzureDatabase,
			}

			if connector.isAzureDatabase != tt.wantAzure {
				t.Errorf("Connector.isAzureDatabase = %v, want %v", connector.isAzureDatabase, tt.wantAzure)
			}
		})
	}
}
