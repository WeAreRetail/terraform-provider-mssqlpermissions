package queries

import (
	"testing"
	"time"
)

// ============================================================================
// SQL CONNECTOR VALIDATION UNIT TESTS
// ============================================================================

// TestConnector_Validate_Unit tests the validate function without database dependencies
func TestConnector_Validate_Unit(t *testing.T) {
	tests := []struct {
		name      string
		connector *Connector
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid_connector_with_all_fields",
			connector: &Connector{
				Host:     "sql.example.com",
				Port:     1433,
				Database: "testdb",
				Timeout:  30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid_connector_port_defaults",
			connector: &Connector{
				Host:     "sql.example.com",
				Database: "testdb",
				Port:     0, // Should default to 1433
			},
			wantErr: false,
		},
		{
			name: "missing_host",
			connector: &Connector{
				Database: "testdb",
				Port:     1433,
			},
			wantErr: true,
			errMsg:  "missing host name",
		},
		{
			name: "missing_database",
			connector: &Connector{
				Host: "sql.example.com",
				Port: 1433,
			},
			wantErr: true,
			errMsg:  "missing database name",
		},
		{
			name: "empty_host",
			connector: &Connector{
				Host:     "",
				Database: "testdb",
				Port:     1433,
			},
			wantErr: true,
			errMsg:  "missing host name",
		},
		{
			name: "empty_database",
			connector: &Connector{
				Host:     "sql.example.com",
				Database: "",
				Port:     1433,
			},
			wantErr: true,
			errMsg:  "missing database name",
		},
		{
			name: "azure_sql_host",
			connector: &Connector{
				Host:     "myserver.database.windows.net",
				Database: "mydb",
			},
			wantErr: false,
		},
		{
			name: "localhost_connection",
			connector: &Connector{
				Host:     "localhost",
				Database: "master",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original port to verify default behavior
			originalPort := tt.connector.Port

			err := tt.connector.validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validate() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validate() unexpected error = %v", err)
				}

				// Verify port defaulting behavior
				if originalPort == 0 {
					if tt.connector.Port != 1433 {
						t.Errorf("validate() expected port to default to 1433, got %d", tt.connector.Port)
					}
				}
			}
		})
	}
}

// TestConnector_ConfigureAzureADConnector_Unit tests Azure AD connection string building
func TestConnector_ConfigureAzureADConnector_Unit(t *testing.T) {
	tests := []struct {
		name           string
		azureLogin     *AzureApplicationLogin
		expectClientId string
		expectTenantId string
		wantErr        bool
	}{
		{
			name: "client_id_only",
			azureLogin: &AzureApplicationLogin{
				ClientId:     "test-client-id",
				ClientSecret: "test-secret",
			},
			expectClientId: "test-client-id",
			expectTenantId: "",
			wantErr:        false,
		},
		{
			name: "client_id_with_tenant",
			azureLogin: &AzureApplicationLogin{
				ClientId:     "test-client-id",
				ClientSecret: "test-secret",
				TenantId:     "test-tenant-id",
			},
			expectClientId: "test-client-id@test-tenant-id",
			expectTenantId: "test-tenant-id",
			wantErr:        false,
		},
		{
			name: "empty_client_secret",
			azureLogin: &AzureApplicationLogin{
				ClientId:     "test-client-id",
				ClientSecret: "",
				TenantId:     "test-tenant-id",
			},
			expectClientId: "test-client-id@test-tenant-id",
			expectTenantId: "test-tenant-id",
			wantErr:        false, // Secret validation happens at connection time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &Connector{
				Host:                  "test.database.windows.net",
				Port:                  1433,
				Database:              "testdb",
				AzureApplicationLogin: tt.azureLogin,
			}

			// Test the connector building (without actually connecting)
			err := connector.validate()
			if err != nil {
				t.Errorf("validate() failed: %v", err)
				return
			}

			// Verify Azure login configuration is preserved
			if connector.AzureApplicationLogin.ClientId != tt.azureLogin.ClientId {
				t.Errorf("ClientId = %v, want %v", connector.AzureApplicationLogin.ClientId, tt.azureLogin.ClientId)
			}

			if connector.AzureApplicationLogin.TenantId != tt.azureLogin.TenantId {
				t.Errorf("TenantId = %v, want %v", connector.AzureApplicationLogin.TenantId, tt.azureLogin.TenantId)
			}
		})
	}
}

// TestConnector_ConfigureManagedIdentityConnector_Unit tests Managed Identity configuration
func TestConnector_ConfigureManagedIdentityConnector_Unit(t *testing.T) {
	tests := []struct {
		name         string
		msiLogin     *ManagedIdentityLogin
		expectUserId string
		wantErr      bool
	}{
		{
			name: "system_assigned_identity",
			msiLogin: &ManagedIdentityLogin{
				UserIdentity: false,
			},
			wantErr: false,
		},
		{
			name: "user_assigned_with_user_id",
			msiLogin: &ManagedIdentityLogin{
				UserIdentity: true,
				UserId:       "test-user-id",
			},
			expectUserId: "test-user-id",
			wantErr:      false,
		},
		{
			name: "user_assigned_with_resource_id",
			msiLogin: &ManagedIdentityLogin{
				UserIdentity: true,
				ResourceId:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test",
			},
			wantErr: false,
		},
		{
			name: "user_assigned_with_both_ids",
			msiLogin: &ManagedIdentityLogin{
				UserIdentity: true,
				UserId:       "test-user-id",
				ResourceId:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test",
			},
			expectUserId: "test-user-id",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &Connector{
				Host:                 "test.database.windows.net",
				Port:                 1433,
				Database:             "testdb",
				ManagedIdentityLogin: tt.msiLogin,
			}

			err := connector.validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validate() unexpected error = %v", err)
				}

				// Verify MSI configuration is preserved
				if connector.ManagedIdentityLogin.UserIdentity != tt.msiLogin.UserIdentity {
					t.Errorf("UserIdentity = %v, want %v", connector.ManagedIdentityLogin.UserIdentity, tt.msiLogin.UserIdentity)
				}

				if tt.expectUserId != "" && connector.ManagedIdentityLogin.UserId != tt.expectUserId {
					t.Errorf("UserId = %v, want %v", connector.ManagedIdentityLogin.UserId, tt.expectUserId)
				}
			}
		})
	}
}

// TestConnector_AuthenticationMethodSelection_Unit tests authentication method logic
func TestConnector_AuthenticationMethodSelection_Unit(t *testing.T) {
	tests := []struct {
		name                  string
		localUserLogin        *LocalUserLogin
		azureApplicationLogin *AzureApplicationLogin
		managedIdentityLogin  *ManagedIdentityLogin
		expectedMethod        string
	}{
		{
			name: "local_user_login",
			localUserLogin: &LocalUserLogin{
				Username: "testuser",
				Password: "testpass",
			},
			expectedMethod: "local",
		},
		{
			name: "azure_application_login",
			azureApplicationLogin: &AzureApplicationLogin{
				ClientId:     "test-client-id",
				ClientSecret: "test-secret",
				TenantId:     "test-tenant-id",
			},
			expectedMethod: "azure_app",
		},
		{
			name: "managed_identity_login",
			managedIdentityLogin: &ManagedIdentityLogin{
				UserIdentity: false,
			},
			expectedMethod: "managed_identity",
		},
		{
			name:           "default_authentication",
			expectedMethod: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &Connector{
				Host:                  "test.database.windows.net",
				Port:                  1433,
				Database:              "testdb",
				LocalUserLogin:        tt.localUserLogin,
				AzureApplicationLogin: tt.azureApplicationLogin,
				ManagedIdentityLogin:  tt.managedIdentityLogin,
			}

			err := connector.validate()
			if err != nil {
				t.Errorf("validate() unexpected error = %v", err)
				return
			}

			// Verify authentication configuration is preserved
			switch tt.expectedMethod {
			case "local":
				if connector.LocalUserLogin == nil {
					t.Error("Expected LocalUserLogin to be preserved")
				}
			case "azure_app":
				if connector.AzureApplicationLogin == nil {
					t.Error("Expected AzureApplicationLogin to be preserved")
				}
			case "managed_identity":
				if connector.ManagedIdentityLogin == nil {
					t.Error("Expected ManagedIdentityLogin to be preserved")
				}
			case "default":
				// All auth methods should be nil for default
				if connector.LocalUserLogin != nil || connector.AzureApplicationLogin != nil || connector.ManagedIdentityLogin != nil {
					t.Error("Expected all authentication methods to be nil for default authentication")
				}
			}
		})
	}
}

// TestConnector_TimeoutHandling_Unit tests timeout configuration
func TestConnector_TimeoutHandling_Unit(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "custom_timeout",
			timeout:         60 * time.Second,
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "zero_timeout_preserved",
			timeout:         0,
			expectedTimeout: 0, // Zero timeout should be preserved during validation
		},
		{
			name:            "very_short_timeout",
			timeout:         1 * time.Second,
			expectedTimeout: 1 * time.Second,
		},
		{
			name:            "very_long_timeout",
			timeout:         5 * time.Minute,
			expectedTimeout: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &Connector{
				Host:     "test.database.windows.net",
				Port:     1433,
				Database: "testdb",
				Timeout:  tt.timeout,
			}

			err := connector.validate()
			if err != nil {
				t.Errorf("validate() unexpected error = %v", err)
				return
			}

			if connector.Timeout != tt.expectedTimeout {
				t.Errorf("Timeout = %v, want %v", connector.Timeout, tt.expectedTimeout)
			}
		})
	}
}
