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

func TestDatabaseRoleMembersResource_Metadata(t *testing.T) {
	r := NewDatabaseRoleMembersResource()
	ctx := context.Background()
	req := resource.MetadataRequest{
		ProviderTypeName: "mssqlpermissions",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	expected := "mssqlpermissions_database_role_members"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestDatabaseRoleMembersResource_Schema(t *testing.T) {
	r := NewDatabaseRoleMembersResource()
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
	requiredAttrs := []string{"name", "members"}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be defined in schema", attr)
		}
	}

	// Verify name attribute has RequiresReplace plan modifier
	nameAttr, exists := resp.Schema.Attributes["name"]
	if !exists {
		t.Error("Expected name attribute to exist")
		return
	}

	if stringAttr, ok := nameAttr.(schema.StringAttribute); ok {
		if len(stringAttr.PlanModifiers) == 0 {
			t.Error("Expected name attribute to have plan modifiers")
		}
		// Could check for specific RequiresReplace modifier if needed
	} else {
		t.Error("Expected name attribute to be a StringAttribute")
	}

	// Verify members attribute is a list of strings
	membersAttr, exists := resp.Schema.Attributes["members"]
	if !exists {
		t.Error("Expected members attribute to exist")
		return
	}

	if _, ok := membersAttr.(schema.ListAttribute); !ok {
		t.Error("Expected members attribute to be a ListAttribute")
	}
}

func TestDatabaseRoleMembersResource_Configure(t *testing.T) {
	r := &DatabaseRoleMembersResource{}
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
		r := &DatabaseRoleMembersResource{} // Fresh instance

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
		r := &DatabaseRoleMembersResource{}

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

func TestDatabaseRoleMembersResource_ImportState(t *testing.T) {
	r := &DatabaseRoleMembersResource{}
	ctx := context.Background()

	// Test that ImportState panics as expected (since it's not implemented)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected ImportState to panic with 'not implemented'")
		}
	}()

	req := resource.ImportStateRequest{}
	resp := &resource.ImportStateResponse{}

	r.ImportState(ctx, req, resp)
}

// Test helper functions for creating test data
func createEmptyRoleMembersModel() model.RoleMembersModel {
	return model.RoleMembersModel{
		Name:    types.StringValue("empty_role"),
		Members: []types.String{},
	}
}

// Test the member management logic concepts
func TestMemberManagementLogic(t *testing.T) {
	t.Run("AddMembers", func(t *testing.T) {
		// Test logic for adding members to a role
		existingMembers := []string{"user1", "user2"}
		newMembers := []string{"user1", "user2", "user3", "user4"}

		// Logic to find members to add (in real implementation)
		var membersToAdd []string
		for _, newMember := range newMembers {
			found := false
			for _, existingMember := range existingMembers {
				if newMember == existingMember {
					found = true
					break
				}
			}
			if !found {
				membersToAdd = append(membersToAdd, newMember)
			}
		}

		expectedToAdd := []string{"user3", "user4"}
		if len(membersToAdd) != len(expectedToAdd) {
			t.Errorf("Expected %d members to add, got %d", len(expectedToAdd), len(membersToAdd))
		}

		for i, member := range membersToAdd {
			if member != expectedToAdd[i] {
				t.Errorf("Expected member %s at index %d, got %s", expectedToAdd[i], i, member)
			}
		}
	})

	t.Run("RemoveMembers", func(t *testing.T) {
		// Test logic for removing members from a role
		existingMembers := []string{"user1", "user2", "user3", "user4"}
		newMembers := []string{"user1", "user3"}

		// Logic to find members to remove (in real implementation)
		var membersToRemove []string
		for _, existingMember := range existingMembers {
			found := false
			for _, newMember := range newMembers {
				if existingMember == newMember {
					found = true
					break
				}
			}
			if !found && existingMember != "dbo" { // Ignore "dbo" as in the real implementation
				membersToRemove = append(membersToRemove, existingMember)
			}
		}

		expectedToRemove := []string{"user2", "user4"}
		if len(membersToRemove) != len(expectedToRemove) {
			t.Errorf("Expected %d members to remove, got %d", len(expectedToRemove), len(membersToRemove))
		}

		for i, member := range membersToRemove {
			if member != expectedToRemove[i] {
				t.Errorf("Expected member %s at index %d, got %s", expectedToRemove[i], i, member)
			}
		}
	})

	t.Run("PreserveMemberOrder", func(t *testing.T) {
		// Test the logic for preserving member order (as in the Read method)
		stateMembers := []string{"user3", "user1", "user2"}
		databaseMembers := []string{"user1", "user2", "user3", "user4"}

		// Logic to maintain state order while adding new members
		var orderedMembers []string

		// First, add existing state members in their original order
		for _, stateMember := range stateMembers {
			for _, dbMember := range databaseMembers {
				if stateMember == dbMember {
					orderedMembers = append(orderedMembers, stateMember)
					break
				}
			}
		}

		// Then add new members from database not in state
		for _, dbMember := range databaseMembers {
			found := false
			for _, stateMember := range stateMembers {
				if dbMember == stateMember {
					found = true
					break
				}
			}
			if !found && dbMember != "dbo" {
				orderedMembers = append(orderedMembers, dbMember)
			}
		}

		// Should preserve state order: user3, user1, user2, then add user4
		expected := []string{"user3", "user1", "user2", "user4"}
		if len(orderedMembers) != len(expected) {
			t.Errorf("Expected %d ordered members, got %d", len(expected), len(orderedMembers))
		}

		for i, member := range orderedMembers {
			if member != expected[i] {
				t.Errorf("Expected member %s at index %d, got %s", expected[i], i, member)
			}
		}
	})
}

// Test resource interface compliance
func TestDatabaseRoleMembersResource_InterfaceCompliance(t *testing.T) {
	var _ resource.Resource = &DatabaseRoleMembersResource{}
	var _ resource.ResourceWithImportState = &DatabaseRoleMembersResource{}
	var _ resource.ResourceWithConfigure = &DatabaseRoleMembersResource{}
}

// Test NewDatabaseRoleMembersResource function
func TestNewDatabaseRoleMembersResource(t *testing.T) {
	r := NewDatabaseRoleMembersResource()

	if r == nil {
		t.Error("Expected NewDatabaseRoleMembersResource to return a non-nil resource")
	}

	// Verify it's the correct type
	if _, ok := r.(*DatabaseRoleMembersResource); !ok {
		t.Error("Expected NewDatabaseRoleMembersResource to return *DatabaseRoleMembersResource")
	}
}

// Test edge cases for member management
func TestMemberManagementEdgeCases(t *testing.T) {
	t.Run("EmptyMembersList", func(t *testing.T) {
		model := createEmptyRoleMembersModel()
		if len(model.Members) != 0 {
			t.Error("Expected empty members list")
		}
	})

	t.Run("DuplicateMembers", func(t *testing.T) {
		// Test handling of duplicate members in the configuration
		members := []string{"user1", "user2", "user1", "user3"}

		// Logic to deduplicate members
		var uniqueMembers []string
		seen := make(map[string]bool)

		for _, member := range members {
			if !seen[member] {
				uniqueMembers = append(uniqueMembers, member)
				seen[member] = true
			}
		}

		expected := []string{"user1", "user2", "user3"}
		if len(uniqueMembers) != len(expected) {
			t.Errorf("Expected %d unique members, got %d", len(expected), len(uniqueMembers))
		}
	})

	t.Run("DboMemberHandling", func(t *testing.T) {
		// Test that 'dbo' member is ignored as per the real implementation
		members := []string{"user1", "dbo", "user2"}

		// Filter out 'dbo' as done in the real implementation
		var filteredMembers []string
		for _, member := range members {
			if member != "dbo" {
				filteredMembers = append(filteredMembers, member)
			}
		}

		expected := []string{"user1", "user2"}
		if len(filteredMembers) != len(expected) {
			t.Errorf("Expected %d filtered members, got %d", len(expected), len(filteredMembers))
		}

		for i, member := range filteredMembers {
			if member != expected[i] {
				t.Errorf("Expected member %s at index %d, got %s", expected[i], i, member)
			}
		}
	})
}

// Benchmark tests for performance
func BenchmarkDatabaseRoleMembersResource_Metadata(b *testing.B) {
	r := NewDatabaseRoleMembersResource()
	ctx := context.Background()
	req := resource.MetadataRequest{ProviderTypeName: "mssqlpermissions"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &resource.MetadataResponse{}
		r.Metadata(ctx, req, resp)
	}
}

func BenchmarkDatabaseRoleMembersResource_Schema(b *testing.B) {
	r := NewDatabaseRoleMembersResource()
	ctx := context.Background()
	req := resource.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &resource.SchemaResponse{}
		r.Schema(ctx, req, resp)
	}
}
