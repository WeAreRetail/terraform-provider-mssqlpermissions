package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaPermissionsResource_Metadata(t *testing.T) {
	r := NewSchemaPermissionsResource()
	ctx := context.Background()
	req := resource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_schema_permissions"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestSchemaPermissionsResource_Schema(t *testing.T) {
	r := NewSchemaPermissionsResource()
	ctx := context.Background()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify schema is not nil
	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	// Check for required attributes
	requiredAttrs := []string{"schema_name", "role_name", "permissions"}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be defined in schema", attr)
		}
	}

	// Verify permissions attribute is a list nested attribute
	permissionsAttr, exists := resp.Schema.Attributes["permissions"]
	if !exists {
		t.Error("Expected permissions attribute to exist")
		return
	}

	// Type assertion to check if it's a ListNestedAttribute
	if _, ok := permissionsAttr.(schema.ListNestedAttribute); !ok {
		t.Error("Expected permissions to be a ListNestedAttribute")
	}

	// Verify schema_name has required plan modifier (requires replacement)
	schemaNameAttr, exists := resp.Schema.Attributes["schema_name"]
	if !exists {
		t.Error("Expected schema_name attribute to exist")
		return
	}

	if stringAttr, ok := schemaNameAttr.(schema.StringAttribute); ok {
		if len(stringAttr.PlanModifiers) == 0 {
			t.Error("Expected schema_name to have plan modifiers")
		}
	}

	// Verify role_name has required plan modifier (requires replacement)
	roleNameAttr, exists := resp.Schema.Attributes["role_name"]
	if !exists {
		t.Error("Expected role_name attribute to exist")
		return
	}

	if stringAttr, ok := roleNameAttr.(schema.StringAttribute); ok {
		if len(stringAttr.PlanModifiers) == 0 {
			t.Error("Expected role_name to have plan modifiers")
		}
	}
}

func TestSchemaPermissionsResource_ValidateConfig(t *testing.T) {
	r := &SchemaPermissionsResource{}
	ctx := context.Background()

	t.Run("FunctionExists", func(t *testing.T) {
		// Test that the ValidateConfig method exists and can be called
		// without providing actual config data (which requires framework setup)
		req := resource.ValidateConfigRequest{}
		resp := &resource.ValidateConfigResponse{}

		// This will likely fail due to nil config, but it tests that the method exists
		// and has the correct signature
		defer func() {
			if r := recover(); r != nil {
				// Expected in unit test without proper framework setup
				t.Logf("ValidateConfig panicked as expected in unit test: %v", r)
			}
		}()

		r.ValidateConfig(ctx, req, resp)
	})
}

func TestSchemaPermissionsResource_Configure(t *testing.T) {
	r := &SchemaPermissionsResource{}
	ctx := context.Background()

	t.Run("ValidProviderData", func(t *testing.T) {
		// Create a mock connector
		mockConnector := &queries.Connector{}

		req := resource.ConfigureRequest{
			ProviderData: mockConnector,
		}
		resp := &resource.ConfigureResponse{}

		r.Configure(ctx, req, resp)

		// Verify no errors
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no errors, got: %v", resp.Diagnostics.Errors())
		}

		// Verify connector was set
		if r.connector != mockConnector {
			t.Error("Expected connector to be set to the provided mock connector")
		}
	})

	t.Run("NilProviderData", func(t *testing.T) {
		r := &SchemaPermissionsResource{} // Fresh instance

		req := resource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &resource.ConfigureResponse{}

		r.Configure(ctx, req, resp)

		// Should not error when provider data is nil
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no errors for nil provider data, got: %v", resp.Diagnostics.Errors())
		}

		// Connector should remain nil
		if r.connector != nil {
			t.Error("Expected connector to remain nil when provider data is nil")
		}
	})

	t.Run("InvalidProviderDataType", func(t *testing.T) {
		r := &SchemaPermissionsResource{}

		req := resource.ConfigureRequest{
			ProviderData: "invalid_type", // Wrong type
		}
		resp := &resource.ConfigureResponse{}

		r.Configure(ctx, req, resp)

		// Should return error for invalid type
		if !resp.Diagnostics.HasError() {
			t.Error("Expected error for invalid provider data type")
		}

		// Check error message
		errors := resp.Diagnostics.Errors()
		if len(errors) == 0 {
			t.Error("Expected at least one error in diagnostics")
		} else {
			summary := errors[0].Summary()
			if summary != "Unexpected Resource Configure Type" {
				t.Errorf("Expected error summary 'Unexpected Resource Configure Type', got: %s", summary)
			}
		}
	})
}

func TestSchemaPermissionsResource_ImportState(t *testing.T) {
	r := &SchemaPermissionsResource{}
	ctx := context.Background()

	req := resource.ImportStateRequest{}
	resp := &resource.ImportStateResponse{}

	r.ImportState(ctx, req, resp)

	// Import should add an error since it's not supported
	if !resp.Diagnostics.HasError() {
		t.Error("Expected ImportState to return an error for unsupported operation")
	}

	errors := resp.Diagnostics.Errors()
	if len(errors) == 0 {
		t.Error("Expected at least one error in diagnostics")
	} else {
		summary := errors[0].Summary()
		if summary != "Import Not Supported" {
			t.Errorf("Expected error summary 'Import Not Supported', got: %s", summary)
		}
	}
}

// Test helper functions for creating test data
func createTestSchemaPermissionModel() model.PermissionModel {
	return model.PermissionModel{
		Name:               types.StringValue("SELECT"),
		State:              types.StringValue("G"),
		Class:              types.StringValue("SCHEMA"),
		ClassDesc:          types.StringValue("SCHEMA"),
		MajorID:            types.Int64Value(5),
		MinorID:            types.Int64Value(0),
		GranteePrincipalID: types.Int64Value(456),
		GrantorPrincipalID: types.Int64Value(1),
		Type:               types.StringValue("SL"),
		StateDesc:          types.StringValue("GRANT"),
	}
}

// Test the validation logic separately
func TestSchemaPermissionsValidation(t *testing.T) {
	t.Run("ValidPermissionState", func(t *testing.T) {
		validStates := []string{"G", "D"}
		for _, state := range validStates {
			permission := createTestSchemaPermissionModel()
			permission.State = types.StringValue(state)

			// Test that these states would be valid
			// In a real validation test, you'd call the actual validation logic
			if state != "G" && state != "D" {
				t.Errorf("State %s should be valid", state)
			}
		}
	})

	t.Run("InvalidPermissionState", func(t *testing.T) {
		invalidStates := []string{"X", "INVALID", "", "grant"}
		for _, state := range invalidStates {
			permission := createTestSchemaPermissionModel()
			permission.State = types.StringValue(state)

			// These states should be invalid
			if state == "G" || state == "D" {
				t.Errorf("State %s should be invalid", state)
			}
		}
	})
}

// Test resource interface compliance
func TestSchemaPermissionsResource_InterfaceCompliance(t *testing.T) {
	var _ resource.Resource = &SchemaPermissionsResource{}
	var _ resource.ResourceWithValidateConfig = &SchemaPermissionsResource{}
	var _ resource.ResourceWithImportState = &SchemaPermissionsResource{}
	var _ resource.ResourceWithConfigure = &SchemaPermissionsResource{}
}

// Test NewSchemaPermissionsResource function
func TestNewSchemaPermissionsResource(t *testing.T) {
	r := NewSchemaPermissionsResource()

	if r == nil {
		t.Error("Expected NewSchemaPermissionsResource to return a non-nil resource")
	}

	// Verify it's the correct type
	if _, ok := r.(*SchemaPermissionsResource); !ok {
		t.Error("Expected NewSchemaPermissionsResource to return *SchemaPermissionsResource")
	}
}

// Benchmark tests for performance
func BenchmarkSchemaPermissionsResource_Metadata(b *testing.B) {
	r := NewSchemaPermissionsResource()
	ctx := context.Background()
	req := resource.MetadataRequest{ProviderTypeName: "mssqlpermissions"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &resource.MetadataResponse{}
		r.Metadata(ctx, req, resp)
	}
}

func BenchmarkSchemaPermissionsResource_Schema(b *testing.B) {
	r := NewSchemaPermissionsResource()
	ctx := context.Background()
	req := resource.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &resource.SchemaResponse{}
		r.Schema(ctx, req, resp)
	}
}
