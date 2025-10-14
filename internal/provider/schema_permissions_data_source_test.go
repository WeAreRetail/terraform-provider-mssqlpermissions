package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/queries"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestSchemaPermissionsDataSource_Metadata(t *testing.T) {
	d := NewSchemaPermissionsDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_schema_permissions"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestSchemaPermissionsDataSource_Schema(t *testing.T) {
	d := NewSchemaPermissionsDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	// Verify schema is not nil
	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	t.Run("RoleNameAttribute", func(t *testing.T) {
		roleNameAttr, exists := resp.Schema.Attributes["role_name"]
		if !exists {
			t.Error("Expected 'role_name' attribute to exist")
			return
		}

		if stringAttr, ok := roleNameAttr.(schema.StringAttribute); ok {
			if !stringAttr.Required {
				t.Error("Expected 'role_name' attribute to be required")
			}
		} else {
			t.Error("Expected 'role_name' to be a StringAttribute")
		}
	})

	t.Run("SchemaNameAttribute", func(t *testing.T) {
		schemaNameAttr, exists := resp.Schema.Attributes["schema_name"]
		if !exists {
			t.Error("Expected 'schema_name' attribute to exist")
			return
		}

		if stringAttr, ok := schemaNameAttr.(schema.StringAttribute); ok {
			if !stringAttr.Required {
				t.Error("Expected 'schema_name' attribute to be required")
			}
		} else {
			t.Error("Expected 'schema_name' to be a StringAttribute")
		}
	})

	t.Run("PermissionsAttribute", func(t *testing.T) {
		permissionsAttr, exists := resp.Schema.Attributes["permissions"]
		if !exists {
			t.Error("Expected 'permissions' attribute to exist")
			return
		}

		if listAttr, ok := permissionsAttr.(schema.ListNestedAttribute); ok {
			if !listAttr.Computed {
				t.Error("Expected 'permissions' attribute to be computed")
			}
		} else {
			t.Error("Expected 'permissions' to be a ListNestedAttribute")
		}
	})
}

func TestSchemaPermissionsDataSource_Configure(t *testing.T) {
	d := &schemaPermissionsDataSource{}
	ctx := context.Background()

	t.Run("ValidProviderData", func(t *testing.T) {
		mockConnector := &queries.Connector{}

		req := datasource.ConfigureRequest{
			ProviderData: mockConnector,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		if d.connector != mockConnector {
			t.Error("Expected connector to be set from provider data")
		}
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no diagnostics, got: %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("NilProviderData", func(t *testing.T) {
		d := &schemaPermissionsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no diagnostics for nil provider data, got: %v", resp.Diagnostics.Errors())
		}

		if d.connector != nil {
			t.Error("Expected connector to remain nil when provider data is nil")
		}
	})

	t.Run("InvalidProviderDataType", func(t *testing.T) {
		d := &schemaPermissionsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "invalid",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Expected error diagnostic for invalid provider data type")
		}
	})
}

func TestSchemaPermissionsDataSource_InterfaceCompliance(t *testing.T) {
	var _ datasource.DataSource = (*schemaPermissionsDataSource)(nil)
	var _ datasource.DataSourceWithConfigure = (*schemaPermissionsDataSource)(nil)
}

func TestNewSchemaPermissionsDataSource(t *testing.T) {
	d := NewSchemaPermissionsDataSource()
	if d == nil {
		t.Error("Expected NewSchemaPermissionsDataSource to return non-nil")
	}
	if _, ok := d.(*schemaPermissionsDataSource); !ok {
		t.Error("Expected NewSchemaPermissionsDataSource to return *schemaPermissionsDataSource")
	}
}
