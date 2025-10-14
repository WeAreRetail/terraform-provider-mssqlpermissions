package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/queries"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestPermissionsDataSource_Metadata(t *testing.T) {
	d := NewPermissionsDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_permissions_to_role"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestPermissionsDataSource_Schema(t *testing.T) {
	d := NewPermissionsDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	t.Run("RoleNameAttribute", func(t *testing.T) {
		attr, exists := resp.Schema.Attributes["role_name"]
		if !exists {
			t.Error("Expected 'role_name' attribute to exist")
			return
		}

		if stringAttr, ok := attr.(schema.StringAttribute); ok {
			if !stringAttr.Required {
				t.Error("Expected 'role_name' attribute to be required")
			}
		} else {
			t.Error("Expected 'role_name' to be a StringAttribute")
		}
	})

	t.Run("PermissionsAttribute", func(t *testing.T) {
		attr, exists := resp.Schema.Attributes["permissions"]
		if !exists {
			t.Error("Expected 'permissions' attribute to exist")
			return
		}

		if listAttr, ok := attr.(schema.ListNestedAttribute); ok {
			if !listAttr.Computed {
				t.Error("Expected 'permissions' attribute to be computed")
			}
		} else {
			t.Error("Expected 'permissions' to be a ListNestedAttribute")
		}
	})
}

func TestPermissionsDataSource_Configure(t *testing.T) {
	d := &permissionsDataSource{}
	ctx := context.Background()

	t.Run("ValidProviderData", func(t *testing.T) {
		mockConnector := &queries.Connector{}

		req := datasource.ConfigureRequest{
			ProviderData: mockConnector,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no errors, got: %v", resp.Diagnostics.Errors())
		}

		if d.connector != mockConnector {
			t.Error("Expected connector to be set to the provided mock connector")
		}
	})

	t.Run("NilProviderData", func(t *testing.T) {
		d := &permissionsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no errors for nil provider data, got: %v", resp.Diagnostics.Errors())
		}

		if d.connector != nil {
			t.Error("Expected connector to remain nil when provider data is nil")
		}
	})

	t.Run("InvalidProviderDataType", func(t *testing.T) {
		d := &permissionsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "invalid_type",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Expected error for invalid provider data type")
		}

		errors := resp.Diagnostics.Errors()
		if len(errors) == 0 {
			t.Error("Expected at least one error in diagnostics")
		} else {
			summary := errors[0].Summary()
			if summary != "Unexpected Data Source Configure Type" {
				t.Errorf("Expected error summary 'Unexpected Data Source Configure Type', got: %s", summary)
			}
		}
	})
}

func TestPermissionsDataSource_InterfaceCompliance(t *testing.T) {
	var _ datasource.DataSource = &permissionsDataSource{}
	var _ datasource.DataSourceWithConfigure = &permissionsDataSource{}
}

func TestNewPermissionsDataSource(t *testing.T) {
	d := NewPermissionsDataSource()

	if d == nil {
		t.Error("Expected NewPermissionsDataSource to return a non-nil data source")
	}

	if _, ok := d.(*permissionsDataSource); !ok {
		t.Error("Expected NewPermissionsDataSource to return *permissionsDataSource")
	}
}

func BenchmarkPermissionsDataSource_Metadata(b *testing.B) {
	d := NewPermissionsDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{ProviderTypeName: "mssqlpermissions"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.MetadataResponse{}
		d.Metadata(ctx, req, resp)
	}
}

func BenchmarkPermissionsDataSource_Schema(b *testing.B) {
	d := NewPermissionsDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.SchemaResponse{}
		d.Schema(ctx, req, resp)
	}
}
