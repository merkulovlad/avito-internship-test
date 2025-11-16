package team

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// Compile-time interface check
var _ domain.TeamRepositoryInterface = (*TeamRepository)(nil)

type TeamRepository struct {
	exec *tx.ExecutorImpl
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{
		exec: tx.NewExecutor(db),
	}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, t domain.TeamName) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx, "INSERT INTO teams (team_name) VALUES ($1)", t)

	return err
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT team_name FROM teams WHERE team_name = $1", name)

	var t domain.Team
	if err := row.Scan(&t.Name); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TeamRepository) Exists(ctx context.Context, name domain.TeamName) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", name)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
