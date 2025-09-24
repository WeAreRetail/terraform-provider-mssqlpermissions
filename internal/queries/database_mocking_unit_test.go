package queries

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"terraform-provider-mssqlpermissions/internal/queries/model"
)

// TestDatabaseInterface_Unit tests the database interface functionality
func TestDatabaseInterface_Unit(t *testing.T) {
	t.Run("MockDatabaseExecutor_ExecContext", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				if query == "CREATE ROLE [test_role]" {
					return &MockResult{RowsAffectedValue: 1}, nil
				}
				return nil, fmt.Errorf("unexpected query: %s", query)
			},
		}

		result, err := mockDB.ExecContext(context.Background(), "CREATE ROLE [test_role]")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("MockDatabaseExecutor_QueryRowContext", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			QueryRowContextFunc: func(ctx context.Context, query string, args ...interface{}) *sql.Row {
				// In a real test, you'd return a mock row with expected data
				// For this demonstration, we'll return nil (which would normally indicate no results)
				return nil
			},
		}

		row := mockDB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM sys.database_roles WHERE name = ?", "test_role")
		if row != nil {
			t.Error("Expected nil row for mock implementation")
		}
	})

	t.Run("MockDatabaseExecutor_PingContext", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			PingContextFunc: func(ctx context.Context) error {
				return nil // Simulate successful ping
			},
		}

		err := mockDB.PingContext(context.Background())
		if err != nil {
			t.Errorf("Expected no error for ping, got %v", err)
		}
	})
}

// TestRoleFunctions_Unit demonstrates unit testing of role functions with mocked database
func TestRoleFunctions_Unit(t *testing.T) {
	t.Run("CreateRole_Success", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				expectedQuery := "CREATE ROLE [test_role]"
				if query != expectedQuery {
					t.Errorf("Expected query %s, got %s", expectedQuery, query)
				}
				return &MockResult{RowsAffectedValue: 1}, nil
			},
		}

		// Test the pattern for how we could refactor createRole to use the interface
		role := model.Role{Name: "test_role"}

		// Simulate what a refactored createRole function would do
		query := fmt.Sprintf("CREATE ROLE [%s]", role.Name)
		result, err := mockDB.ExecContext(context.Background(), query)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("CreateRole_Error", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				return nil, fmt.Errorf("role already exists")
			},
		}

		role := model.Role{Name: "existing_role"}
		query := fmt.Sprintf("CREATE ROLE [%s]", role.Name)
		_, err := mockDB.ExecContext(context.Background(), query)

		if err == nil {
			t.Error("Expected error for existing role, got nil")
		}

		expectedMsg := "role already exists"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

// TestPermissionFunctions_Unit demonstrates unit testing of permission functions with mocked database
func TestPermissionFunctions_Unit(t *testing.T) {
	t.Run("GrantPermission_Success", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				expectedQuery := "GRANT SELECT ON SCHEMA::[dbo] TO [test_role]"
				if query != expectedQuery {
					t.Errorf("Expected query %s, got %s", expectedQuery, query)
				}
				return &MockResult{RowsAffectedValue: 1}, nil
			},
		}

		// Test the pattern for how we could refactor grantPermission to use the interface
		permission := model.Permission{
			State: "GRANT",
			Name:  "SELECT",
		}
		schema := "dbo"
		role := "test_role"

		// Simulate what a refactored grantPermission function would do
		query := fmt.Sprintf("%s %s ON SCHEMA::[%s] TO [%s]",
			permission.State, permission.Name, schema, role)
		result, err := mockDB.ExecContext(context.Background(), query)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("RevokePermission_Success", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				expectedQuery := "REVOKE INSERT ON SCHEMA::[test_schema] FROM [test_role]"
				if query != expectedQuery {
					t.Errorf("Expected query %s, got %s", expectedQuery, query)
				}
				return &MockResult{RowsAffectedValue: 1}, nil
			},
		}

		permission := model.Permission{
			State: "REVOKE",
			Name:  "INSERT",
		}
		schema := "test_schema"
		role := "test_role"

		var query string
		if permission.State == "REVOKE" {
			query = fmt.Sprintf("%s %s ON SCHEMA::[%s] FROM [%s]",
				permission.State, permission.Name, schema, role)
		}

		result, err := mockDB.ExecContext(context.Background(), query)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})
}

// TestUserFunctions_Unit demonstrates unit testing of user functions with mocked database
func TestUserFunctions_Unit(t *testing.T) {
	t.Run("CreateUser_Success", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				expectedQuery := "CREATE USER [test_user] WITH PASSWORD = 'test_password'"
				if query != expectedQuery {
					t.Errorf("Expected query %s, got %s", expectedQuery, query)
				}
				return &MockResult{RowsAffectedValue: 1}, nil
			},
		}

		user := model.User{
			Name:     "test_user",
			Password: "test_password",
		}

		// Simulate what a refactored createUser function would do
		query := fmt.Sprintf("CREATE USER [%s] WITH PASSWORD = '%s'", user.Name, user.Password)
		result, err := mockDB.ExecContext(context.Background(), query)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("CreateUser_AzureAD", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				expectedQuery := "CREATE USER [test_user] FROM EXTERNAL PROVIDER"
				if query != expectedQuery {
					t.Errorf("Expected query %s, got %s", expectedQuery, query)
				}
				return &MockResult{RowsAffectedValue: 1}, nil
			},
		}

		user := model.User{
			Name:     "test_user",
			ObjectID: "12345678-1234-1234-1234-123456789012", // Azure AD user
		}

		// Simulate Azure AD user creation
		var query string
		if user.ObjectID != "" {
			query = fmt.Sprintf("CREATE USER [%s] FROM EXTERNAL PROVIDER", user.Name)
		}

		result, err := mockDB.ExecContext(context.Background(), query)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})
}

// TestErrorHandling_Unit tests error handling with mocked database failures
func TestErrorHandling_Unit(t *testing.T) {
	t.Run("DatabaseConnection_Error", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			PingContextFunc: func(ctx context.Context) error {
				return fmt.Errorf("connection failed: server unreachable")
			},
		}

		err := mockDB.PingContext(context.Background())
		if err == nil {
			t.Error("Expected connection error, got nil")
		}

		expectedMsg := "connection failed: server unreachable"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Query_Error", func(t *testing.T) {
		mockDB := &MockDatabaseExecutor{
			ExecContextFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
				return nil, fmt.Errorf("syntax error in SQL statement")
			},
		}

		_, err := mockDB.ExecContext(context.Background(), "INVALID SQL QUERY")
		if err == nil {
			t.Error("Expected SQL error, got nil")
		}

		expectedMsg := "syntax error in SQL statement"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}
