//go:build integration

package queries

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

// TestParameterMutationPrevention verifies that functions don't mutate input parameters
func TestParameterMutationPrevention(t *testing.T) {
	ctx := context.Background()

	// Mock connector - we'll only test the logic that creates copies, not DB operations
	connector := &Connector{}

	t.Run("GrantPermissionToRole should not mutate input permission", func(t *testing.T) {
		// Create a permission with a specific state
		originalPermission := &model.Permission{
			Name:  "SELECT",
			State: "R", // Original state
			Type:  "SL",
		}

		// Store the original state for comparison
		originalState := originalPermission.State

		// This will fail at the database level, but we want to verify the parameter isn't mutated
		// The function should create a copy before setting State = "G"
		_ = connector.GrantPermissionToRole(ctx, nil, &model.Role{Name: "test"}, originalPermission)

		// Verify the original permission wasn't modified
		if originalPermission.State != originalState {
			t.Errorf("GrantPermissionToRole mutated input parameter: expected state %s, got %s", originalState, originalPermission.State)
		}
	})

	t.Run("DenyPermissionToRole should not mutate input permission", func(t *testing.T) {
		// Create a permission with a specific state
		originalPermission := &model.Permission{
			Name:  "INSERT",
			State: "G", // Original state
			Type:  "IN",
		}

		// Store the original state for comparison
		originalState := originalPermission.State

		// This will fail at the database level, but we want to verify the parameter isn't mutated
		// The function should create a copy before setting State = "D"
		_ = connector.DenyPermissionToRole(ctx, nil, &model.Role{Name: "test"}, originalPermission)

		// Verify the original permission wasn't modified
		if originalPermission.State != originalState {
			t.Errorf("DenyPermissionToRole mutated input parameter: expected state %s, got %s", originalState, originalPermission.State)
		}
	})

	t.Run("GrantPermissionOnSchemaToRole should not mutate input permission", func(t *testing.T) {
		// Create a permission with a specific state
		originalPermission := &model.Permission{
			Name:  "UPDATE",
			State: "R", // Original state
			Type:  "UP",
		}

		// Store the original state for comparison
		originalState := originalPermission.State

		// This will fail at the database level, but we want to verify the parameter isn't mutated
		_ = connector.GrantPermissionOnSchemaToRole(ctx, nil, &model.Role{Name: "test"}, "custom", originalPermission)

		// Verify the original permission wasn't modified
		if originalPermission.State != originalState {
			t.Errorf("GrantPermissionOnSchemaToRole mutated input parameter: expected state %s, got %s", originalState, originalPermission.State)
		}
	})

	t.Run("DenyPermissionOnSchemaToRole should not mutate input permission", func(t *testing.T) {
		// Create a permission with a specific state
		originalPermission := &model.Permission{
			Name:  "DELETE",
			State: "G", // Original state
			Type:  "DL",
		}

		// Store the original state for comparison
		originalState := originalPermission.State

		// This will fail at the database level, but we want to verify the parameter isn't mutated
		_ = connector.DenyPermissionOnSchemaToRole(ctx, nil, &model.Role{Name: "test"}, "custom", originalPermission)

		// Verify the original permission wasn't modified
		if originalPermission.State != originalState {
			t.Errorf("DenyPermissionOnSchemaToRole mutated input parameter: expected state %s, got %s", originalState, originalPermission.State)
		}
	})

	t.Run("CreateUser should not mutate input user", func(t *testing.T) {
		// Create a user with empty DefaultSchema to test mutation prevention
		originalUser := &model.User{
			Name:          "testuser",
			DefaultSchema: "", // This should trigger the defaulting logic
			Password:      "password123",
		}

		// Store the original values for comparison
		originalDefaultSchema := originalUser.DefaultSchema

		// This will fail at the database level, but we want to verify the parameter isn't mutated
		// The function should create a copy before setting DefaultSchema = "dbo"
		_ = connector.CreateUser(ctx, nil, originalUser)

		// Verify the original user wasn't modified
		if originalUser.DefaultSchema != originalDefaultSchema {
			t.Errorf("CreateUser mutated input parameter: expected DefaultSchema %s, got %s", originalDefaultSchema, originalUser.DefaultSchema)
		}
	})

	t.Run("CreateDatabaseRole should not mutate input role", func(t *testing.T) {
		// Create a role with PrincipalID = 0 to test mutation prevention
		originalRole := &model.Role{
			Name:        "testrole",
			PrincipalID: 0, // This should trigger the defaulting logic
		}

		// Store the original values for comparison
		originalPrincipalID := originalRole.PrincipalID

		// This will fail at the database level, but we want to verify the parameter isn't mutated
		// The function should create a copy before setting PrincipalID = 1
		_ = connector.CreateDatabaseRole(ctx, nil, originalRole)

		// Verify the original role wasn't modified
		if originalRole.PrincipalID != originalPrincipalID {
			t.Errorf("CreateDatabaseRole mutated input parameter: expected PrincipalID %d, got %d", originalPrincipalID, originalRole.PrincipalID)
		}
	})
}
