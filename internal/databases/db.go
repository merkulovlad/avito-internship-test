// Package databases provides database connection functionality. Also includes migration helpers.
package databases

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/merkulovlad/avito-internship-test/internal/config"
)

// DRIVER is the database driver name.
const DRIVER = "postgres"

// NewDB creates and returns a new database connection pool.
func NewDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := cfg.DSN()

	db, err := sql.Open(DRIVER, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	db.SetMaxOpenConns(cfg.MaxConnections)

	maxIdleConns := cfg.MaxConnections / 2
	if maxIdleConns < 1 {
		maxIdleConns = 1
	}

	db.SetMaxIdleConns(maxIdleConns)

	return db, nil
}
