package tx

import (
	"context"
	"database/sql"
)

type ctxTxKey struct{}

// withTx adds a transaction to the context.
func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, ctxTxKey{}, tx)
}

// TxFromContext retrieves a transaction from the context.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx)
	return tx, ok
}
