# Test Architecture Documentation

## Overview

This document outlines the test architecture for the terraform-provider-mssqlpermissions project, which now includes both unit tests and integration tests with clear separation of concerns.

## Test Types

### Unit Tests
Unit tests are designed to test individual functions and components in isolation without requiring external dependencies like database connections. They run quickly and can be executed in any environment.

**Files:**
- `validation_unit_test.go` - Tests for validation functions (validateRoleName, validatePermissionName, etc.)
- `user_validation_unit_test.go` - Tests for user validation logic and connector configurations
- `database_mocking_unit_test.go` - Tests demonstrating database mocking patterns for future refactoring

**Characteristics:**
- No database connections required
- Use mocked dependencies
- Fast execution (< 10ms per test)
- Can run in CI/CD without infrastructure setup
- Focus on pure business logic and validation

**Run command:**
```bash
go test -v ./internal/queries -run "Test.*_Unit"
```

### Integration Tests
Integration tests require actual database connections and test the full end-to-end functionality with real SQL Server instances.

**Files:**
- `permissions_test.go` - Integration tests for permission operations
- `user_test.go` - Integration tests for user CRUD operations
- `databaseRole_test.go` - Integration tests for database role operations
- `sql_test.go` - Integration tests for SQL execution utilities

**Characteristics:**
- Require database setup (local Docker or Azure SQL)
- Test real database interactions
- Slower execution (seconds per test)
- Require infrastructure and configuration
- Test complete workflows including error handling

**Run command:**
```bash
go test -v ./internal/queries -run "Test.*" -skip "Test.*_Unit"
```

## Build Tags for Test Separation

To better separate unit tests from integration tests, you can use build tags:

### Option 1: Tag Integration Tests
Add `//go:build integration` to the top of integration test files and run:

```bash
# Run only unit tests (default)
go test -v ./internal/queries

# Run only integration tests
go test -tags=integration -v ./internal/queries

# Run all tests
go test -tags=integration -v ./internal/queries -run "Test"
```

### Option 2: Tag Unit Tests
Add `//go:build unit` to unit test files and run:

```bash
# Run only unit tests
go test -tags=unit -v ./internal/queries

# Run only integration tests (default)
go test -v ./internal/queries

# Run all tests
go test -tags=unit,integration -v ./internal/queries
```

## Test Coverage Analysis

### Before Unit Tests
- Only integration tests existed
- Required database setup for any testing
- Slow test execution
- Environment dependencies made testing difficult

### After Unit Tests
- Comprehensive unit test coverage for validation functions
- Database mocking infrastructure for future refactoring
- Fast feedback loop for core business logic
- CI/CD friendly testing without infrastructure dependencies

## Database Mocking Architecture

The project now includes database interfaces and mocking infrastructure:

**Files:**
- `database_interface.go` - Defines database interfaces for dependency injection
- `database_mocking_unit_test.go` - Demonstrates mocking patterns

**Key Components:**
- `DatabaseExecutor` interface - Abstracts database operations
- `MockDatabaseExecutor` - Mock implementation for testing
- `MockResult` - Mock implementation of sql.Result
- Adapter functions for backward compatibility

## Migration Strategy

To gradually adopt unit testing for existing database operations:

1. **Extract Business Logic**: Move validation and transformation logic to pure functions
2. **Inject Dependencies**: Use the DatabaseInterface for all database operations
3. **Add Unit Tests**: Test pure functions with mocks
4. **Keep Integration Tests**: Maintain existing integration tests for end-to-end validation

## Best Practices

### Unit Tests
- Test one function or component at a time
- Use descriptive test names with scenarios
- Test both success and error cases
- Include edge cases and boundary conditions
- Keep tests fast and deterministic

### Integration Tests
- Test complete workflows
- Use realistic test data
- Clean up resources after tests
- Test error handling with real failure scenarios
- Document required test environment setup

### Test Data Management
- Use table-driven tests for multiple scenarios
- Generate valid test data programmatically
- Avoid hardcoded values that might become invalid
- Use helper functions for common test setup

## Future Improvements

1. **Refactor Existing Functions**: Gradually introduce dependency injection to existing database functions
2. **Increase Unit Test Coverage**: Add unit tests for more business logic functions
3. **Parallel Test Execution**: Optimize test performance with parallel execution
4. **Test Data Factories**: Create builders for complex test data structures
5. **Property-Based Testing**: Consider adding property-based tests for validation functions
