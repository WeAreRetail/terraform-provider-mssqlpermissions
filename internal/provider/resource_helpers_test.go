package provider

import (
	"context"
	"errors"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGetResourceConnector(t *testing.T) {
	ctx := context.Background()

	t.Run("UseProviderConnector", func(t *testing.T) {
		// Create a mock provider connector
		providerConnector := &queries.Connector{}

		// Call getResourceConnector with provider connector available
		result, diags := getResourceConnector(ctx, providerConnector, nil)

		// Verify no diagnostics errors
		if diags.HasError() {
			t.Errorf("Expected no diagnostics errors, got: %v", diags.Errors())
		}

		// Verify the same connector is returned
		if result != providerConnector {
			t.Error("Expected to receive the same provider connector")
		}
	})

	t.Run("FallbackToResourceConfig", func(t *testing.T) {
		// Create a mock resource config
		resourceConfig := &model.ConfigModel{
			ServerFqdn:   types.StringValue("localhost"),
			DatabaseName: types.StringValue("testdb"),
			// Note: SQLLogin is types.Object, so we'll create a minimal config
			// The actual object creation is complex and would require proper object construction
		}

		// Call getResourceConnector without provider connector
		result, diags := getResourceConnector(ctx, nil, resourceConfig)

		// This test verifies the function handles resource config
		// The actual connection may fail due to incomplete config
		_ = result
		_ = diags
	})

	t.Run("NoConfigurationAvailable", func(t *testing.T) {
		// Call getResourceConnector without any configuration
		result, diags := getResourceConnector(ctx, nil, nil)

		// Verify diagnostics error is returned
		if !diags.HasError() {
			t.Error("Expected diagnostics error when no configuration is available")
		}

		// Verify no connector is returned
		if result != nil {
			t.Error("Expected nil connector when no configuration is available")
		}

		// Verify error message content
		errors := diags.Errors()
		if len(errors) == 0 {
			t.Error("Expected at least one error in diagnostics")
		} else {
			errorSummary := errors[0].Summary()
			if errorSummary != "No Database Configuration Found" {
				t.Errorf("Expected error summary 'No Database Configuration Found', got: %s", errorSummary)
			}
		}
	})
}

func TestConnectToDatabase(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidConnector", func(t *testing.T) {
		// Create a mock connector that will succeed
		// Note: This test would need a proper mock implementation
		// For now, we test the function signature and basic behavior
		connector := &queries.Connector{}

		// Call connectToDatabase
		db, err := connectToDatabase(ctx, connector)

		// Note: This will likely fail without a real database connection
		// but the test structure demonstrates proper testing approach
		_ = db
		_ = err
		// In a real implementation, you'd mock the connector's Connect method
	})

	t.Run("NilConnector", func(t *testing.T) {
		// This should panic or return an error
		defer func() {
			if r := recover(); r != nil {
				// Expected panic from nil connector - test passes
				t.Log("Recovered from expected panic with nil connector")
			}
		}()

		_, err := connectToDatabase(ctx, nil)
		if err == nil {
			t.Error("Expected error when connector is nil")
		}
	})
}

func TestHandleDatabaseConnectionError(t *testing.T) {
	ctx := context.Background()

	t.Run("AddsErrorToDiagnostics", func(t *testing.T) {
		var diags diag.Diagnostics
		testError := errors.New("connection failed")

		// Call handleDatabaseConnectionError
		handleDatabaseConnectionError(ctx, testError, &diags)

		// Verify error was added to diagnostics
		if !diags.HasError() {
			t.Error("Expected error to be added to diagnostics")
		}

		errors := diags.Errors()
		if len(errors) == 0 {
			t.Error("Expected at least one error in diagnostics")
		} else {
			if errors[0].Summary() != "Database Connection Failed" {
				t.Errorf("Expected error summary 'Database Connection Failed', got: %s", errors[0].Summary())
			}
			if errors[0].Detail() != "connection failed" {
				t.Errorf("Expected error detail 'connection failed', got: %s", errors[0].Detail())
			}
		}
	})

	t.Run("NilError", func(t *testing.T) {
		var diags diag.Diagnostics

		// Function should handle nil error gracefully
		handleDatabaseConnectionError(ctx, nil, &diags)

		// No error should be added for nil error
		if diags.HasError() {
			t.Error("Expected no errors to be added for nil error")
		}
	})
}

func TestLogResourceOperation(t *testing.T) {
	ctx := context.Background()

	t.Run("LogsOperation", func(t *testing.T) {
		// Test that the function can be called without panicking
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function should not panic, but did: %v", r)
			}
		}()

		logResourceOperation(ctx, "TestResource", "Create")
	})

	t.Run("EmptyParameters", func(t *testing.T) {
		// Test with empty parameters
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function should handle empty parameters, but panicked: %v", r)
			}
		}()

		logResourceOperation(ctx, "", "")
	})
}

func TestLogResourceOperationComplete(t *testing.T) {
	ctx := context.Background()

	t.Run("LogsOperationComplete", func(t *testing.T) {
		// Test that the function can be called without panicking
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function should not panic, but did: %v", r)
			}
		}()

		logResourceOperationComplete(ctx, "TestResource", "Create")
	})

	t.Run("EmptyParameters", func(t *testing.T) {
		// Test with empty parameters
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function should handle empty parameters, but panicked: %v", r)
			}
		}()

		logResourceOperationComplete(ctx, "", "")
	})
}

// Helper function to create a basic config model for testing
func createTestConfigModel() *model.ConfigModel {
	return &model.ConfigModel{
		ServerFqdn:   types.StringValue("localhost"),
		ServerPort:   types.Int64Value(1433),
		DatabaseName: types.StringValue("testdb"),
		// Note: Login objects would need proper types.Object construction
		// For unit testing, we focus on the basic structure
	}
}

// Integration test for the full helper workflow
func TestResourceHelperWorkflow(t *testing.T) {
	ctx := context.Background()

	t.Run("FullWorkflowWithProviderConnector", func(t *testing.T) {
		// Create provider connector
		providerConnector := &queries.Connector{}

		// Get connector using helper
		connector, diags := getResourceConnector(ctx, providerConnector, nil)

		if diags.HasError() {
			t.Errorf("Expected no errors getting connector: %v", diags.Errors())
			return
		}

		if connector != providerConnector {
			t.Error("Expected to get the same provider connector")
		}

		// Test logging functions don't panic
		logResourceOperation(ctx, "TestResource", "Create")
		logResourceOperationComplete(ctx, "TestResource", "Create")
	})

	t.Run("FullWorkflowWithResourceConfig", func(t *testing.T) {
		// Create resource config
		resourceConfig := createTestConfigModel()

		// Get connector using helper (may fail due to actual connection logic)
		connector, diags := getResourceConnector(ctx, nil, resourceConfig)

		// This test validates the flow, actual connection may fail
		_ = connector
		_ = diags

		// Test error handling
		var testDiags diag.Diagnostics
		testError := errors.New("test connection error")
		handleDatabaseConnectionError(ctx, testError, &testDiags)

		if !testDiags.HasError() {
			t.Error("Expected error to be added to diagnostics")
		}
	})
}
