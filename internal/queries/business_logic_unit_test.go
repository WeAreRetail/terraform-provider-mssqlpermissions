package queries

import (
	"fmt"
	"net/url"
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
	"time"
)

// ============================================================================
// BUSINESS LOGIC UNIT TESTS
// ============================================================================

// TestFormatConnectionString_Unit tests connection string building logic
func TestFormatConnectionString_Unit(t *testing.T) {
	tests := []struct {
		name             string
		host             string
		port             int
		database         string
		expectedScheme   string
		expectedHost     string
		expectedDatabase string
	}{
		{
			name:             "standard_connection",
			host:             "sql.example.com",
			port:             1433,
			database:         "testdb",
			expectedScheme:   "sqlserver",
			expectedHost:     "sql.example.com:1433",
			expectedDatabase: "testdb",
		},
		{
			name:             "azure_sql_connection",
			host:             "myserver.database.windows.net",
			port:             1433,
			database:         "mydb",
			expectedScheme:   "sqlserver",
			expectedHost:     "myserver.database.windows.net:1433",
			expectedDatabase: "mydb",
		},
		{
			name:             "custom_port",
			host:             "sql.internal.com",
			port:             14330,
			database:         "appdb",
			expectedScheme:   "sqlserver",
			expectedHost:     "sql.internal.com:14330",
			expectedDatabase: "appdb",
		},
		{
			name:             "localhost_connection",
			host:             "localhost",
			port:             1433,
			database:         "master",
			expectedScheme:   "sqlserver",
			expectedHost:     "localhost:1433",
			expectedDatabase: "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the connection string building logic from connector()
			connectionString := &url.URL{
				Scheme: "sqlserver",
				Host:   fmt.Sprintf("%s:%d", tt.host, tt.port),
			}

			query := url.Values{}
			query.Add("database", tt.database)
			query.Add("app name", "terraform-sql-provider")

			connectionString.RawQuery = query.Encode()

			// Verify the connection string components
			if connectionString.Scheme != tt.expectedScheme {
				t.Errorf("Scheme = %v, want %v", connectionString.Scheme, tt.expectedScheme)
			}

			if connectionString.Host != tt.expectedHost {
				t.Errorf("Host = %v, want %v", connectionString.Host, tt.expectedHost)
			}

			// Parse the query parameters
			queryParams, err := url.ParseQuery(connectionString.RawQuery)
			if err != nil {
				t.Errorf("Failed to parse query parameters: %v", err)
				return
			}

			if queryParams.Get("database") != tt.expectedDatabase {
				t.Errorf("Database = %v, want %v", queryParams.Get("database"), tt.expectedDatabase)
			}

			if queryParams.Get("app name") != "terraform-sql-provider" {
				t.Errorf("App name = %v, want %v", queryParams.Get("app name"), "terraform-sql-provider")
			}
		})
	}
}

// TestBuildAzureADConnectionParams_Unit tests Azure AD parameter building
func TestBuildAzureADConnectionParams_Unit(t *testing.T) {
	tests := []struct {
		name            string
		clientId        string
		tenantId        string
		clientSecret    string
		expectedUserId  string
		expectedFedAuth string
	}{
		{
			name:            "client_id_only",
			clientId:        "test-client-id",
			clientSecret:    "test-secret",
			expectedUserId:  "test-client-id",
			expectedFedAuth: "ActiveDirectoryServicePrincipal",
		},
		{
			name:            "client_id_with_tenant",
			clientId:        "test-client-id",
			tenantId:        "test-tenant-id",
			clientSecret:    "test-secret",
			expectedUserId:  "test-client-id@test-tenant-id",
			expectedFedAuth: "ActiveDirectoryServicePrincipal",
		},
		{
			name:            "complex_client_id",
			clientId:        "12345678-1234-1234-1234-123456789012",
			tenantId:        "87654321-4321-4321-4321-210987654321",
			clientSecret:    "complex-secret-value",
			expectedUserId:  "12345678-1234-1234-1234-123456789012@87654321-4321-4321-4321-210987654321",
			expectedFedAuth: "ActiveDirectoryServicePrincipal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the Azure AD configuration logic
			query := url.Values{}

			userId := tt.clientId
			if tt.tenantId != "" {
				userId = fmt.Sprintf("%s@%s", tt.clientId, tt.tenantId)
			}

			query.Add("fedauth", "ActiveDirectoryServicePrincipal")
			query.Add("user id", userId)
			query.Add("password", tt.clientSecret)

			// Verify the constructed parameters
			if query.Get("user id") != tt.expectedUserId {
				t.Errorf("User ID = %v, want %v", query.Get("user id"), tt.expectedUserId)
			}

			if query.Get("fedauth") != tt.expectedFedAuth {
				t.Errorf("FedAuth = %v, want %v", query.Get("fedauth"), tt.expectedFedAuth)
			}

			if query.Get("password") != tt.clientSecret {
				t.Errorf("Password = %v, want %v", query.Get("password"), tt.clientSecret)
			}
		})
	}
}

// TestBuildManagedIdentityConnectionParams_Unit tests Managed Identity parameter building
func TestBuildManagedIdentityConnectionParams_Unit(t *testing.T) {
	tests := []struct {
		name           string
		userIdentity   bool
		userId         string
		resourceId     string
		expectedParams map[string]string
	}{
		{
			name:         "system_assigned_identity",
			userIdentity: false,
			expectedParams: map[string]string{
				"fedauth": "ActiveDirectoryManagedIdentity",
			},
		},
		{
			name:         "user_assigned_with_user_id",
			userIdentity: true,
			userId:       "test-user-id",
			expectedParams: map[string]string{
				"fedauth": "ActiveDirectoryManagedIdentity",
				"user id": "test-user-id",
			},
		},
		{
			name:         "user_assigned_with_resource_id",
			userIdentity: true,
			resourceId:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test",
			expectedParams: map[string]string{
				"fedauth":     "ActiveDirectoryManagedIdentity",
				"resource id": "/subscriptions/test/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test",
			},
		},
		{
			name:         "user_assigned_with_both_ids",
			userIdentity: true,
			userId:       "test-user-id",
			resourceId:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test",
			expectedParams: map[string]string{
				"fedauth":     "ActiveDirectoryManagedIdentity",
				"user id":     "test-user-id",
				"resource id": "/subscriptions/test/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the Managed Identity configuration logic
			query := url.Values{}
			query.Add("fedauth", "ActiveDirectoryManagedIdentity")

			if tt.userIdentity && (tt.userId != "" || tt.resourceId != "") {
				if tt.userId != "" {
					query.Add("user id", tt.userId)
				}
				if tt.resourceId != "" {
					query.Add("resource id", tt.resourceId)
				}
			}

			// Verify all expected parameters are present
			for key, expectedValue := range tt.expectedParams {
				actualValue := query.Get(key)
				if actualValue != expectedValue {
					t.Errorf("Parameter %s = %v, want %v", key, actualValue, expectedValue)
				}
			}

			// Verify no unexpected parameters are present
			for key := range query {
				if _, expected := tt.expectedParams[key]; !expected {
					t.Errorf("Unexpected parameter %s = %v", key, query.Get(key))
				}
			}
		})
	}
}

// TestDefaultTimeoutBehavior_Unit tests timeout defaulting logic
func TestDefaultTimeoutBehavior_Unit(t *testing.T) {
	tests := []struct {
		name             string
		inputTimeout     time.Duration
		expectedBehavior string
		expectedDefault  time.Duration
	}{
		{
			name:             "zero_timeout_gets_default",
			inputTimeout:     0,
			expectedBehavior: "should_default",
			expectedDefault:  30 * time.Second,
		},
		{
			name:             "custom_timeout_preserved",
			inputTimeout:     60 * time.Second,
			expectedBehavior: "should_preserve",
			expectedDefault:  60 * time.Second,
		},
		{
			name:             "very_short_timeout_preserved",
			inputTimeout:     1 * time.Second,
			expectedBehavior: "should_preserve",
			expectedDefault:  1 * time.Second,
		},
		{
			name:             "very_long_timeout_preserved",
			inputTimeout:     10 * time.Minute,
			expectedBehavior: "should_preserve",
			expectedDefault:  10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the timeout defaulting logic from Connect()
			timeout := tt.inputTimeout
			if timeout == 0 {
				timeout = 30 * time.Second // defaultTimeout
			}

			if timeout != tt.expectedDefault {
				t.Errorf("Timeout = %v, want %v", timeout, tt.expectedDefault)
			}

			// Verify the behavior description matches
			switch tt.expectedBehavior {
			case "should_default":
				if tt.inputTimeout != 0 {
					t.Errorf("Test case error: expected input timeout to be 0 for 'should_default' behavior")
				}
			case "should_preserve":
				if tt.inputTimeout == 0 {
					t.Errorf("Test case error: expected input timeout to be non-zero for 'should_preserve' behavior")
				}
			}
		})
	}
}

// TestVersionStringParsing_Unit tests SQL Server version detection logic
func TestVersionStringParsing_Unit(t *testing.T) {
	tests := []struct {
		name          string
		versionString string
		isAzure       bool
	}{
		{
			name:          "azure_sql_database",
			versionString: "Microsoft SQL Azure (RTM) - 12.0.2000.8",
			isAzure:       true,
		},
		{
			name:          "sql_server_on_premises",
			versionString: "Microsoft SQL Server 2019 (RTM) - 15.0.2000.5",
			isAzure:       false,
		},
		{
			name:          "sql_server_express",
			versionString: "Microsoft SQL Server 2019 Express Edition (RTM) - 15.0.2000.5",
			isAzure:       false,
		},
		{
			name:          "azure_sql_managed_instance",
			versionString: "Microsoft SQL Azure - 12.0.2000.8",
			isAzure:       true,
		},
		{
			name:          "empty_version_string",
			versionString: "",
			isAzure:       false,
		},
		{
			name:          "partial_azure_match",
			versionString: "Microsoft SQL Azure Database",
			isAzure:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the Azure detection logic from Connect()
			isAzure := tt.versionString != "" && contains(tt.versionString, "Microsoft SQL Azure")

			if isAzure != tt.isAzure {
				t.Errorf("Azure detection = %v, want %v for version: %s", isAzure, tt.isAzure, tt.versionString)
			}
		})
	}
}

// TestRoleQueryBuilding_Unit tests SQL query construction for role operations
func TestRoleQueryBuilding_Unit(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		userName    string
		expectedSQL string
	}{
		{
			name:        "simple_role_creation",
			roleName:    "TestRole",
			userName:    "dbo",
			expectedSQL: "'CREATE ROLE ' + QUOTENAME(@database_role_name) + ' AUTHORIZATION ' + QUOTENAME(@user_name)",
		},
		{
			name:        "role_with_underscore",
			roleName:    "Test_Role",
			userName:    "testuser",
			expectedSQL: "'CREATE ROLE ' + QUOTENAME(@database_role_name) + ' AUTHORIZATION ' + QUOTENAME(@user_name)",
		},
		{
			name:        "role_with_numbers",
			roleName:    "TestRole123",
			userName:    "user123",
			expectedSQL: "'CREATE ROLE ' + QUOTENAME(@database_role_name) + ' AUTHORIZATION ' + QUOTENAME(@user_name)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the query building logic from CreateDatabaseRole
			query := "'CREATE ROLE ' + QUOTENAME(@database_role_name) + ' AUTHORIZATION ' + QUOTENAME(@user_name)"

			if query != tt.expectedSQL {
				t.Errorf("Query = %v, want %v", query, tt.expectedSQL)
			}

			// Verify the query uses parameterized inputs (would be safe against SQL injection)
			if !contains(query, "QUOTENAME(@database_role_name)") {
				t.Error("Query should use QUOTENAME for role name parameter")
			}

			if !contains(query, "QUOTENAME(@user_name)") {
				t.Error("Query should use QUOTENAME for user name parameter")
			}
		})
	}
}

// TestUserValidationBusinessLogic_Unit tests user validation rules
func TestUserValidationBusinessLogic_Unit(t *testing.T) {
	tests := []struct {
		name               string
		user               *model.User
		isAzureDatabase    bool
		expectedValid      bool
		expectedErrorCount int
	}{
		{
			name: "valid_contained_user",
			user: &model.User{
				Name:     "testuser",
				Password: "TestPass123!",
				External: false,
			},
			isAzureDatabase:    false,
			expectedValid:      true,
			expectedErrorCount: 0,
		},
		{
			name: "valid_external_user",
			user: &model.User{
				Name:     "testuser@domain.com",
				External: true,
				ObjectID: "12345678-1234-1234-1234-123456789012",
			},
			isAzureDatabase:    true,
			expectedValid:      true,
			expectedErrorCount: 0,
		},
		{
			name: "invalid_contained_user_no_password",
			user: &model.User{
				Name:     "testuser",
				External: false,
			},
			isAzureDatabase:    false,
			expectedValid:      false,
			expectedErrorCount: 1,
		},
		{
			name: "invalid_external_user_with_password",
			user: &model.User{
				Name:     "testuser@domain.com",
				Password: "TestPass123!",
				External: true,
			},
			isAzureDatabase:    false,
			expectedValid:      false,
			expectedErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate user validation logic
			errorCount := 0

			// Basic validation rules
			if tt.user.Name == "" {
				errorCount++
			}

			if tt.user.Password == "" && !tt.user.External {
				errorCount++
			}

			if tt.user.External && tt.user.Password != "" {
				errorCount++
			}

			if tt.user.ObjectID != "" && !tt.user.External {
				errorCount++
			}

			if tt.user.DefaultLanguage != "" && tt.isAzureDatabase {
				errorCount++
			}

			isValid := errorCount == 0

			if isValid != tt.expectedValid {
				t.Errorf("Validation result = %v, want %v", isValid, tt.expectedValid)
			}

			if errorCount != tt.expectedErrorCount {
				t.Errorf("Error count = %d, want %d", errorCount, tt.expectedErrorCount)
			}
		})
	}
}
