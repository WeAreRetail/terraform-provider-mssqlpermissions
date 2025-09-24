package queries

import (
	"context"
	"database/sql"
)

// ============================================================================
// DATABASE INTERFACES FOR MOCKING AND DEPENDENCY INJECTION
// ============================================================================

// DatabaseExecutor interface abstracts database operations for testing
type DatabaseExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PingContext(ctx context.Context) error
}

// TransactionExecutor interface abstracts transaction operations for testing
type TransactionExecutor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// DatabaseInterface combines both executor interfaces
type DatabaseInterface interface {
	DatabaseExecutor
	TransactionExecutor
}

// Ensure *sql.DB implements our interface
var _ DatabaseInterface = (*sql.DB)(nil)

// MockDatabaseExecutor is a mock implementation for unit testing
type MockDatabaseExecutor struct {
	ExecContextFunc     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContextFunc    func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContextFunc func(ctx context.Context, query string, args ...interface{}) *sql.Row
	PingContextFunc     func(ctx context.Context) error
	BeginTxFunc         func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func (m *MockDatabaseExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.ExecContextFunc != nil {
		return m.ExecContextFunc(ctx, query, args...)
	}
	return &MockResult{}, nil
}

func (m *MockDatabaseExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryContextFunc != nil {
		return m.QueryContextFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockDatabaseExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.QueryRowContextFunc != nil {
		return m.QueryRowContextFunc(ctx, query, args...)
	}
	return nil
}

func (m *MockDatabaseExecutor) PingContext(ctx context.Context) error {
	if m.PingContextFunc != nil {
		return m.PingContextFunc(ctx)
	}
	return nil
}

func (m *MockDatabaseExecutor) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, opts)
	}
	return nil, nil
}

// MockResult implements sql.Result for testing
type MockResult struct {
	LastInsertIdValue int64
	RowsAffectedValue int64
	ErrorValue        error
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.LastInsertIdValue, m.ErrorValue
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.RowsAffectedValue, m.ErrorValue
}

// ============================================================================
// ADAPTER FUNCTIONS FOR BACKWARD COMPATIBILITY
// ============================================================================
//
// Note: Additional helper functions can be added here in the future if needed
// for gradual migration or backwards compatibility.
