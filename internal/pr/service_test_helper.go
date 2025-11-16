package pr

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// newTestTxManager creates a transaction manager for testing using sqlmock
// expectCommit: true for success cases, false for error cases that should rollback
func newTestTxManager(expectCommit bool) (*tx.Manager, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	// Expect begin for any transaction
	mock.ExpectBegin()

	if expectCommit {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}

	cleanup := func() {
		_ = db.Close()
	}

	return tx.NewManager(db), mock, cleanup
}
