package provider

import (
	"context"
	"database/sql"
	"errors"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"
	"testing"
)

type mockDatabaseRoleCreateOperations struct {
	rolesToReturn         []*qmodel.Role
	defaultRole           *qmodel.Role
	getRoleCallCount      int
	createRoleCallCount   int
	createRoleShouldError bool
}

type mockDatabaseRoleDeleteOperations struct {
	roleToReturn        *qmodel.Role
	getRoleErr          error
	deleteRoleErr       error
	getRoleCallCount    int
	deleteRoleCallCount int
}

func (m *mockDatabaseRoleCreateOperations) GetDatabaseRole(_ context.Context, _ *sql.DB, _ *qmodel.Role) (*qmodel.Role, error) {
	m.getRoleCallCount++
	if len(m.rolesToReturn) > 0 {
		result := m.rolesToReturn[0]
		m.rolesToReturn = m.rolesToReturn[1:]
		return result, nil
	}

	return m.defaultRole, nil
}

func (m *mockDatabaseRoleCreateOperations) CreateDatabaseRole(_ context.Context, _ *sql.DB, _ *qmodel.Role) error {
	m.createRoleCallCount++
	if m.createRoleShouldError {
		return sql.ErrConnDone
	}
	return nil
}

func (m *mockDatabaseRoleDeleteOperations) GetDatabaseRole(_ context.Context, _ *sql.DB, _ *qmodel.Role) (*qmodel.Role, error) {
	m.getRoleCallCount++
	return m.roleToReturn, m.getRoleErr
}

func (m *mockDatabaseRoleDeleteOperations) DeleteDatabaseRole(_ context.Context, _ *sql.DB, _ *qmodel.Role) error {
	m.deleteRoleCallCount++
	return m.deleteRoleErr
}

func TestEnsureDatabaseRoleForCreate_unit(t *testing.T) {
	ctx := context.Background()

	t.Run("Create standard role calls CreateDatabaseRole once", func(t *testing.T) {
		mockConnector := &mockDatabaseRoleCreateOperations{
			rolesToReturn: []*qmodel.Role{
				nil,
				{Name: "custom_role", IsFixedRole: false},
			},
		}

		role := &qmodel.Role{Name: "custom_role"}

		_, err := ensureDatabaseRoleForCreate(ctx, mockConnector, nil, role)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mockConnector.createRoleCallCount != 1 {
			t.Fatalf("expected CreateDatabaseRole to be called once, got %d", mockConnector.createRoleCallCount)
		}

		if mockConnector.getRoleCallCount != 2 {
			t.Fatalf("expected GetDatabaseRole to be called twice (before and after create), got %d", mockConnector.getRoleCallCount)
		}
	})

	t.Run("Create existing role fails", func(t *testing.T) {
		mockConnector := &mockDatabaseRoleCreateOperations{
			defaultRole: &qmodel.Role{Name: "existing_role", IsFixedRole: false},
		}

		role := &qmodel.Role{Name: "existing_role"}

		_, err := ensureDatabaseRoleForCreate(ctx, mockConnector, nil, role)
		if err == nil {
			t.Fatal("expected error when role already exists and is not built-in")
		}

		if mockConnector.createRoleCallCount != 0 {
			t.Fatalf("expected CreateDatabaseRole not to be called, got %d", mockConnector.createRoleCallCount)
		}

		if mockConnector.getRoleCallCount != 1 {
			t.Fatalf("expected GetDatabaseRole to be called once, got %d", mockConnector.getRoleCallCount)
		}
	})

	t.Run("Create built-in role does nothing", func(t *testing.T) {
		mockConnector := &mockDatabaseRoleCreateOperations{
			defaultRole: &qmodel.Role{Name: "db_owner", IsFixedRole: true},
		}

		role := &qmodel.Role{Name: "db_owner"}

		_, err := ensureDatabaseRoleForCreate(ctx, mockConnector, nil, role)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mockConnector.createRoleCallCount != 0 {
			t.Fatalf("expected CreateDatabaseRole not to be called, got %d", mockConnector.createRoleCallCount)
		}

		if mockConnector.getRoleCallCount != 1 {
			t.Fatalf("expected GetDatabaseRole to be called once, got %d", mockConnector.getRoleCallCount)
		}
	})
}

func TestEnsureDatabaseRoleDeleted_unit(t *testing.T) {
	ctx := context.Background()

	t.Run("No deletion of built-in role", func(t *testing.T) {
		mockConnector := &mockDatabaseRoleDeleteOperations{
			roleToReturn: &qmodel.Role{Name: "db_owner", IsFixedRole: true},
		}

		err := ensureDatabaseRoleDeleted(ctx, mockConnector, nil, &qmodel.Role{Name: "db_owner"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mockConnector.deleteRoleCallCount != 0 {
			t.Fatalf("expected DeleteDatabaseRole not to be called, got %d", mockConnector.deleteRoleCallCount)
		}
	})

	t.Run("Deletion of non built-in role", func(t *testing.T) {
		mockConnector := &mockDatabaseRoleDeleteOperations{
			roleToReturn: &qmodel.Role{Name: "custom_role", IsFixedRole: false},
		}

		err := ensureDatabaseRoleDeleted(ctx, mockConnector, nil, &qmodel.Role{Name: "custom_role"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mockConnector.deleteRoleCallCount != 1 {
			t.Fatalf("expected DeleteDatabaseRole to be called once, got %d", mockConnector.deleteRoleCallCount)
		}
	})

	t.Run("No error if role does not exist", func(t *testing.T) {
		mockConnector := &mockDatabaseRoleDeleteOperations{
			getRoleErr: errors.New("database role not found"),
		}

		err := ensureDatabaseRoleDeleted(ctx, mockConnector, nil, &qmodel.Role{Name: "manually_deleted_role"})
		if err != nil {
			t.Fatalf("expected no error when role is missing, got: %v", err)
		}

		if mockConnector.deleteRoleCallCount != 0 {
			t.Fatalf("expected DeleteDatabaseRole not to be called, got %d", mockConnector.deleteRoleCallCount)
		}
	})
}
