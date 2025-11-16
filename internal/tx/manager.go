package tx

import (
	"context"
	"database/sql"
	"fmt"
)

var _ ManagerInterface = (*Manager)(nil)

// Manager handles database transaction management.
type Manager struct {
	db *sql.DB
}

// NewManager creates a new Manager instance.
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// Do executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	ctxWithTx := withTx(ctx, tx)

	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original: %w)", rbErr, err)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
