package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDatabaseRoleDataSource_Metadata(t *testing.T) {
	d := NewDatabaseRoleDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_database_role"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestDatabaseRoleDataSource_Schema(t *testing.T) {
	d := NewDatabaseRoleDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	// Verify schema is not nil
	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	// Check for required attributes
	requiredAttrs := []string{
		"name", "members", "principal_id",
		"type", "type_description", "owning_principal", "is_fixed_role",
	}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be defined in schema", attr)
		}
	}

	// Verify specific attribute types
	t.Run("NameAttribute", func(t *testing.T) {
		nameAttr, exists := resp.Schema.Attributes["name"]
		if !exists {
			t.Error("Expected name attribute to exist")
			return
		}

		if stringAttr, ok := nameAttr.(schema.StringAttribute); ok {
			if !stringAttr.Optional {
				t.Error("Expected name attribute to be optional")
			}
			if !stringAttr.Computed {
				t.Error("Expected name attribute to be computed")
			}
		} else {
			t.Error("Expected name attribute to be a StringAttribute")
		}
	})

	t.Run("MembersAttribute", func(t *testing.T) {
		membersAttr, exists := resp.Schema.Attributes["members"]
		if !exists {
			t.Error("Expected members attribute to exist")
			return
		}

		if listAttr, ok := membersAttr.(schema.ListAttribute); ok {
			if !listAttr.Computed {
				t.Error("Expected members attribute to be computed")
			}
			if listAttr.ElementType != types.StringType {
				t.Error("Expected members element type to be StringType")
			}
		} else {
			t.Error("Expected members attribute to be a ListAttribute")
		}
	})

	t.Run("PrincipalIdAttribute", func(t *testing.T) {
		principalIdAttr, exists := resp.Schema.Attributes["principal_id"]
		if !exists {
			t.Error("Expected principal_id attribute to exist")
			return
		}

		if int64Attr, ok := principalIdAttr.(schema.Int64Attribute); ok {
			if !int64Attr.Computed {
				t.Error("Expected principal_id attribute to be computed")
			}
		} else {
			t.Error("Expected principal_id attribute to be an Int64Attribute")
		}
	})

	t.Run("IsFixedRoleAttribute", func(t *testing.T) {
		isFixedRoleAttr, exists := resp.Schema.Attributes["is_fixed_role"]
		if !exists {
			t.Error("Expected is_fixed_role attribute to exist")
			return
		}

		if boolAttr, ok := isFixedRoleAttr.(schema.BoolAttribute); ok {
			if !boolAttr.Computed {
				t.Error("Expected is_fixed_role attribute to be computed")
			}
		} else {
			t.Error("Expected is_fixed_role attribute to be a BoolAttribute")
		}
	})
}

func TestUserDataSource_Metadata(t *testing.T) {
	d := NewUserDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_user"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestUserDataSource_Schema(t *testing.T) {
	d := NewUserDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	// Verify schema is not nil
	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	// Check for required attributes based on the user data source structure
	expectedAttrs := []string{
		"name", "external", "principal_id",
		"default_schema", "default_language", "sid", "object_id",
	}
	for _, attr := range expectedAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be defined in schema", attr)
		}
	}

	// Verify specific attribute types
	t.Run("NameAttribute", func(t *testing.T) {
		nameAttr, exists := resp.Schema.Attributes["name"]
		if !exists {
			t.Error("Expected name attribute to exist")
			return
		}

		if stringAttr, ok := nameAttr.(schema.StringAttribute); ok {
			// Name should be optional and computed for data sources
			if !stringAttr.Optional {
				t.Error("Expected name attribute to be optional")
			}
			if !stringAttr.Computed {
				t.Error("Expected name attribute to be computed")
			}
		} else {
			t.Error("Expected name attribute to be a StringAttribute")
		}
	})

	t.Run("ExternalAttribute", func(t *testing.T) {
		externalAttr, exists := resp.Schema.Attributes["external"]
		if !exists {
			t.Error("Expected external attribute to exist")
			return
		}

		if boolAttr, ok := externalAttr.(schema.BoolAttribute); ok {
			if !boolAttr.Computed {
				t.Error("Expected external attribute to be computed")
			}
		} else {
			t.Error("Expected external attribute to be a BoolAttribute")
		}
	})

	t.Run("PrincipalIdAttribute", func(t *testing.T) {
		principalIdAttr, exists := resp.Schema.Attributes["principal_id"]
		if !exists {
			t.Error("Expected principal_id attribute to exist")
			return
		}

		if int64Attr, ok := principalIdAttr.(schema.Int64Attribute); ok {
			if !int64Attr.Computed {
				t.Error("Expected principal_id attribute to be computed")
			}
		} else {
			t.Error("Expected principal_id attribute to be an Int64Attribute")
		}
	})
}

// Test helper functions for creating test data models
func createTestRoleModel() model.RoleModel {
	return model.RoleModel{
		Name:            types.StringValue("test_role"),
		PrincipalID:     types.Int64Value(123),
		Type:            types.StringValue("R"),
		TypeDescription: types.StringValue("DATABASE_ROLE"),
		OwningPrincipal: types.StringValue("dbo"),
		IsFixedRole:     types.BoolValue(false),
	}
}

func createTestRoleDataSourceModel() model.RoleDataSourceModel {
	return model.RoleDataSourceModel{
		Name:            types.StringValue("test_role"),
		PrincipalID:     types.Int64Value(123),
		Type:            types.StringValue("R"),
		TypeDescription: types.StringValue("DATABASE_ROLE"),
		OwningPrincipal: types.StringValue("dbo"),
		IsFixedRole:     types.BoolValue(false),
		Members: []types.String{
			types.StringValue("user1"),
			types.StringValue("user2"),
		},
	}
}

func createEmptyUserModel() model.UserDataModel {
	return model.UserDataModel{
		Name:            types.StringValue("empty_user"),
		External:        types.BoolValue(true),
		PrincipalID:     types.Int64Value(789),
		DefaultSchema:   types.StringValue("guest"),
		DefaultLanguage: types.StringValue("English"),
		SID:             types.StringValue("0x1111111111111111111111111111111111111111"),
		ObjectID:        types.StringNull(), // Null object ID
	}
}

// Test data source interface compliance
func TestDatabaseRoleDataSource_InterfaceCompliance(t *testing.T) {
	var _ datasource.DataSource = &databaseRoleDataSource{}
}

func TestUserDataSource_InterfaceCompliance(t *testing.T) {
	var _ datasource.DataSource = &userDataSource{}
}

// Test NewDataSource functions
func TestNewDatabaseRoleDataSource(t *testing.T) {
	d := NewDatabaseRoleDataSource()

	if d == nil {
		t.Error("Expected NewDatabaseRoleDataSource to return a non-nil data source")
	}

	// Verify it's the correct type
	if _, ok := d.(*databaseRoleDataSource); !ok {
		t.Error("Expected NewDatabaseRoleDataSource to return *databaseRoleDataSource")
	}
}

func TestNewUserDataSource(t *testing.T) {
	d := NewUserDataSource()

	if d == nil {
		t.Error("Expected NewUserDataSource to return a non-nil data source")
	}

	// Verify it's the correct type
	if _, ok := d.(*userDataSource); !ok {
		t.Error("Expected NewUserDataSource to return *userDataSource")
	}
}

// Test edge cases for data source models
func TestDataSourceModelEdgeCases(t *testing.T) {
	t.Run("RoleWithNoMembers", func(t *testing.T) {
		role := createTestRoleDataSourceModel()
		role.Members = []types.String{} // Empty members

		if len(role.Members) != 0 {
			t.Error("Expected empty members list")
		}
	})

	t.Run("UserWithNullObjectID", func(t *testing.T) {
		user := createEmptyUserModel()

		if !user.ObjectID.IsNull() {
			t.Error("Expected ObjectID to be null")
		}
	})

	t.Run("FixedDatabaseRole", func(t *testing.T) {
		role := createTestRoleModel()
		role.Name = types.StringValue("db_owner")
		role.IsFixedRole = types.BoolValue(true)

		if !role.IsFixedRole.ValueBool() {
			t.Error("Expected IsFixedRole to be true for fixed roles")
		}
	})

	t.Run("ExternalUser", func(t *testing.T) {
		user := createEmptyUserModel()
		user.External = types.BoolValue(true)

		if !user.External.ValueBool() {
			t.Error("Expected External to be true for external users")
		}
	})
}

// Test schema validation logic
func TestSchemaValidation(t *testing.T) {
	t.Run("DatabaseRoleSchema_MarkdownDescription", func(t *testing.T) {
		d := NewDatabaseRoleDataSource()
		ctx := context.Background()
		req := datasource.SchemaRequest{}
		resp := &datasource.SchemaResponse{}

		d.Schema(ctx, req, resp)

		expected := "Database role data source."
		if resp.Schema.MarkdownDescription != expected {
			t.Errorf("Expected MarkdownDescription %s, got %s", expected, resp.Schema.MarkdownDescription)
		}
	})

	t.Run("UserSchema_MarkdownDescription", func(t *testing.T) {
		d := NewUserDataSource()
		ctx := context.Background()
		req := datasource.SchemaRequest{}
		resp := &datasource.SchemaResponse{}

		d.Schema(ctx, req, resp)

		expected := "User data source."
		if resp.Schema.MarkdownDescription != expected {
			t.Errorf("Expected MarkdownDescription %s, got %s", expected, resp.Schema.MarkdownDescription)
		}
	})
}

// Benchmark tests for performance
func BenchmarkDatabaseRoleDataSource_Metadata(b *testing.B) {
	d := NewDatabaseRoleDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{ProviderTypeName: "mssqlpermissions"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.MetadataResponse{}
		d.Metadata(ctx, req, resp)
	}
}

func BenchmarkDatabaseRoleDataSource_Schema(b *testing.B) {
	d := NewDatabaseRoleDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.SchemaResponse{}
		d.Schema(ctx, req, resp)
	}
}

func BenchmarkUserDataSource_Metadata(b *testing.B) {
	d := NewUserDataSource()
	ctx := context.Background()
	req := datasource.MetadataRequest{ProviderTypeName: "mssqlpermissions"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.MetadataResponse{}
		d.Metadata(ctx, req, resp)
	}
}

func BenchmarkUserDataSource_Schema(b *testing.B) {
	d := NewUserDataSource()
	ctx := context.Background()
	req := datasource.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &datasource.SchemaResponse{}
		d.Schema(ctx, req, resp)
	}
}
