package provider

import (
	"errors"
	"testing"
)

// TestHandleUserReadError tests the actual error handling function used by the UserResource Read method
func TestHandleUserReadError(t *testing.T) {
	tests := []struct {
		name                   string
		err                    error
		expectedShouldRemove   bool
		expectedShouldAddError bool
		expectedErrorMessage   string
	}{
		{
			name:                   "User not found - should remove from state",
			err:                    errors.New("user not found"),
			expectedShouldRemove:   true,
			expectedShouldAddError: false,
		},
		{
			name:                   "Database connection error - should add error",
			err:                    errors.New("database connection failed"),
			expectedShouldRemove:   false,
			expectedShouldAddError: true,
			expectedErrorMessage:   "Error getting user",
		},
		{
			name:                   "SQL syntax error - should add error",
			err:                    errors.New("invalid syntax"),
			expectedShouldRemove:   false,
			expectedShouldAddError: true,
			expectedErrorMessage:   "Error getting user",
		},
		{
			name:                   "No error - should continue normally",
			err:                    nil,
			expectedShouldRemove:   false,
			expectedShouldAddError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the ACTUAL function used in the Read method
			result := HandleUserReadError(tt.err)

			if result.ShouldRemoveFromState != tt.expectedShouldRemove {
				t.Errorf("Expected ShouldRemoveFromState to be %v, got %v", tt.expectedShouldRemove, result.ShouldRemoveFromState)
			}

			if result.ShouldAddError != tt.expectedShouldAddError {
				t.Errorf("Expected ShouldAddError to be %v, got %v", tt.expectedShouldAddError, result.ShouldAddError)
			}

			if tt.expectedErrorMessage != "" && result.ErrorMessage != tt.expectedErrorMessage {
				t.Errorf("Expected ErrorMessage to be '%s', got '%s'", tt.expectedErrorMessage, result.ErrorMessage)
			}
		})
	}
}

// TestHandleDatabaseRoleReadError tests the actual error handling function used by database role resources
func TestHandleDatabaseRoleReadError(t *testing.T) {
	tests := []struct {
		name                   string
		err                    error
		expectedShouldRemove   bool
		expectedShouldAddError bool
		expectedErrorMessage   string
	}{
		{
			name:                   "Database role not found - should remove from state",
			err:                    errors.New("database role not found"),
			expectedShouldRemove:   true,
			expectedShouldAddError: false,
		},
		{
			name:                   "Permission denied - should add error",
			err:                    errors.New("permission denied"),
			expectedShouldRemove:   false,
			expectedShouldAddError: true,
			expectedErrorMessage:   "Error getting role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the ACTUAL function used in the Read methods
			result := HandleDatabaseRoleReadError(tt.err)

			if result.ShouldRemoveFromState != tt.expectedShouldRemove {
				t.Errorf("Expected ShouldRemoveFromState to be %v, got %v", tt.expectedShouldRemove, result.ShouldRemoveFromState)
			}

			if result.ShouldAddError != tt.expectedShouldAddError {
				t.Errorf("Expected ShouldAddError to be %v, got %v", tt.expectedShouldAddError, result.ShouldAddError)
			}

			if tt.expectedErrorMessage != "" && result.ErrorMessage != tt.expectedErrorMessage {
				t.Errorf("Expected ErrorMessage to be '%s', got '%s'", tt.expectedErrorMessage, result.ErrorMessage)
			}
		})
	}
}

// TestHandlePermissionReadError tests the actual error handling function used by permissions resource
func TestHandlePermissionReadError(t *testing.T) {
	tests := []struct {
		name                   string
		err                    error
		expectedShouldRemove   bool
		expectedShouldAddError bool
		expectedErrorMessage   string
	}{
		{
			name:                   "Database role not found - should remove from state",
			err:                    errors.New("database role not found"),
			expectedShouldRemove:   true,
			expectedShouldAddError: false,
		},
		{
			name:                   "Permissions not found - should skip permission",
			err:                    errors.New("permissions not found"),
			expectedShouldRemove:   false,
			expectedShouldAddError: false,
		},
		{
			name:                   "Access denied - should add error",
			err:                    errors.New("access denied"),
			expectedShouldRemove:   false,
			expectedShouldAddError: true,
			expectedErrorMessage:   "Error getting permission for role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the ACTUAL function used in the Read method
			result := HandlePermissionReadError(tt.err)

			if result.ShouldRemoveFromState != tt.expectedShouldRemove {
				t.Errorf("Expected ShouldRemoveFromState to be %v, got %v", tt.expectedShouldRemove, result.ShouldRemoveFromState)
			}

			if result.ShouldAddError != tt.expectedShouldAddError {
				t.Errorf("Expected ShouldAddError to be %v, got %v", tt.expectedShouldAddError, result.ShouldAddError)
			}

			if tt.expectedErrorMessage != "" && result.ErrorMessage != tt.expectedErrorMessage {
				t.Errorf("Expected ErrorMessage to be '%s', got '%s'", tt.expectedErrorMessage, result.ErrorMessage)
			}
		})
	}
}
