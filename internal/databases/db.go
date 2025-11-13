package databases

import (
	"database/sql"
	"time"
	"github.com/merkulovlad/avito-internship-test/internal/config"
)

const DRIVER = "postgres"

func NewDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := cfg.DSN()
	db, err := sql.Open(DRIVER, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Duration(cfg.ConnectionTimeout) * time.Second)
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxConnections / 2)

	return db, nil
}