// Package team implements team management business logic.
package team

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// Compile-time interface check
var _ domain.TeamRepositoryInterface = (*TeamRepository)(nil)

// TeamRepository provides database operations for teams.
type TeamRepository struct {
	exec *tx.ExecutorImpl
}

// NewTeamRepository creates a new TeamRepository instance.
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{
		exec: tx.NewExecutor(db),
	}
}

// CreateTeam inserts a new team into the database.
func (r *TeamRepository) CreateTeam(ctx context.Context, t domain.TeamName) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx, "INSERT INTO teams (team_name) VALUES ($1)", t)

	return err
}

// GetTeamByName retrieves a team by its name from the database.
func (r *TeamRepository) GetTeamByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT team_name FROM teams WHERE team_name = $1", name)

	var t domain.Team
	if err := row.Scan(&t.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}

	return &t, nil
}

// Exists checks if a team with the given name exists in the database.
func (r *TeamRepository) Exists(ctx context.Context, name domain.TeamName) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", name)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
