package tx

import (
	"context"
	"database/sql"
)

// ExecutorImpl provides access to transaction or database.
type ExecutorImpl struct {
	db *sql.DB
}

// NewExecutor creates a new ExecutorImpl instance.
func NewExecutor(db *sql.DB) *ExecutorImpl {
	return &ExecutorImpl{db: db}
}

// DefaultTxOrDB returns transaction from context if exists, otherwise returns the database
func (e *ExecutorImpl) DefaultTxOrDB(ctx context.Context) Executor {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}

	return e.db
}
