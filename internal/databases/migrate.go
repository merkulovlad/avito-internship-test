package databases

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	dialect      = "postgres"
	migrationDir = "migrations"
)

// RunMigrations applies database migrations using goose.
func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if err := goose.Up(db, migrationDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
