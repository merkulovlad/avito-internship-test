package tx

import (
	"context"
	"database/sql"
)

// Executor defines the interface for database query execution.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ManagerInterface defines the interface for transaction management.
type ManagerInterface interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// ExecutorProvider defines the interface for providing an executor.
type ExecutorProvider interface {
	DefaultTxOrDB(ctx context.Context) Executor
}
