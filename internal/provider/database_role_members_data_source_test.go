package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/queries"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestDatabaseRoleMembersDataSource_Metadata(t *testing.T) {
	d := NewDatabaseRoleMembersDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_database_role_members"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestDatabaseRoleMembersDataSource_Schema(t *testing.T) {
	d := NewDatabaseRoleMembersDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	// Verify schema is not nil
	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	t.Run("NameAttribute", func(t *testing.T) {
		nameAttr, exists := resp.Schema.Attributes["name"]
		if !exists {
			t.Error("Expected 'name' attribute to exist")
			return
		}

		if stringAttr, ok := nameAttr.(schema.StringAttribute); ok {
			if !stringAttr.Required {
				t.Error("Expected 'name' attribute to be required")
			}
		} else {
			t.Error("Expected 'name' to be a StringAttribute")
		}
	})

	t.Run("MembersAttribute", func(t *testing.T) {
		membersAttr, exists := resp.Schema.Attributes["members"]
		if !exists {
			t.Error("Expected 'members' attribute to exist")
			return
		}

		if listAttr, ok := membersAttr.(schema.ListAttribute); ok {
			if !listAttr.Computed {
				t.Error("Expected 'members' attribute to be computed")
			}
		} else {
			t.Error("Expected 'members' to be a ListAttribute")
		}
	})
}

func TestDatabaseRoleMembersDataSource_Configure(t *testing.T) {
	d := &databaseRoleMembersDataSource{}
	ctx := context.Background()

	t.Run("ValidProviderData", func(t *testing.T) {
		mockConnector := &queries.Connector{}

		req := datasource.ConfigureRequest{
			ProviderData: mockConnector,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		// Verify no errors
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no errors, got: %v", resp.Diagnostics.Errors())
		}

		// Verify connector was set
		if d.connector != mockConnector {
			t.Error("Expected connector to be set to the provided mock connector")
		}
	})

	t.Run("NilProviderData", func(t *testing.T) {
		d := &databaseRoleMembersDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

		// Should not error when provider data is nil
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no errors for nil provider data, got: %v", resp.Diagnostics.Errors())
		}

		// Connector should remain nil
		if d.connector != nil {
			t.Error("Expected connector to remain nil when provider data is nil")
		}
	})

	t.Run("InvalidProviderDataType", func(t *testing.T) {
		d := &databaseRoleMembersDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "invalid_type",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(ctx, req, resp)

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
			if summary != "Unexpected Data Source Configure Type" {
				t.Errorf("Expected error summary 'Unexpected Data Source Configure Type', got: %s", summary)
			}
		}
	})
}

// Test data source interface compliance
func TestDatabaseRoleMembersDataSource_InterfaceCompliance(t *testing.T) {
	var _ datasource.DataSource = &databaseRoleMembersDataSource{}
	var _ datasource.DataSourceWithConfigure = &databaseRoleMembersDataSource{}
}

// Test NewDatabaseRoleMembersDataSource function
func TestNewDatabaseRoleMembersDataSource(t *testing.T) {
	d := NewDatabaseRoleMembersDataSource()

	if d == nil {
		t.Error("Expected NewDatabaseRoleMembersDataSource to return a non-nil data source")
	}

	// Verify it's the correct type
	if _, ok := d.(*databaseRoleMembersDataSource); !ok {
		t.Error("Expected NewDatabaseRoleMembersDataSource to return *databaseRoleMembersDataSource")
	}
}

// Benchmark tests for performance
func BenchmarkDatabaseRoleMembersDataSource_Metadata(b *testing.B) {
	d := NewDatabaseRoleMembersDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{ProviderTypeName: "mssqlpermissions"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.MetadataResponse{}
		d.Metadata(ctx, req, resp)
	}
}

func BenchmarkDatabaseRoleMembersDataSource_Schema(b *testing.B) {
	d := NewDatabaseRoleMembersDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.SchemaResponse{}
		d.Schema(ctx, req, resp)
	}
}
