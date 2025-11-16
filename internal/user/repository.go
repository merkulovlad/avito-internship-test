package user

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// Compile-time interface check
var _ domain.UserRepositoryInterface = (*UserRepository)(nil)

// UserRepository provides methods to manage users
type UserRepository struct {
	exec *tx.ExecutorImpl
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		exec: tx.NewExecutor(db),
	}
}

func (r *UserRepository) Upsert(ctx context.Context, u *domain.User) error {
	executor := r.exec.DefaultTxOrDB(ctx)

	_, err := executor.ExecContext(ctx, `
        INSERT INTO users (user_id, username, team_name, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) DO UPDATE
        SET username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active
    `,
		u.ID,
		u.Username,
		u.TeamName,
		u.IsActive,
	)

	return err
}

func (r *UserRepository) SetUserIsActive(ctx context.Context, userId domain.UserID, isActive bool) (*domain.User, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING user_id, username, team_name, is_active", isActive, userId)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetActiveUsersByTeam(ctx context.Context, teamName domain.TeamName) ([]domain.UserID, error) {
	executor := r.exec.DefaultTxOrDB(ctx)

	rows, err := executor.QueryContext(ctx,
		`SELECT user_id 
         FROM users 
         WHERE team_name = $1 AND is_active = TRUE`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []domain.UserID

	for rows.Next() {
		var id domain.UserID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (r *UserRepository) Exists(ctx context.Context, userId domain.UserID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", userId)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userId domain.UserID) (*domain.User, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", userId)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetUsersByIDs(ctx context.Context, userIds []domain.UserID) ([]domain.User, error) {
	if len(userIds) == 0 {
		return []domain.User{}, nil
	}

	executor := r.exec.DefaultTxOrDB(ctx)

	// Build the query with proper placeholders
	query := `SELECT user_id, username, team_name, is_active FROM users WHERE user_id = ANY($1)`

	// Convert []UserID to []string for the ANY clause
	ids := make([]string, len(userIds))
	for i, id := range userIds {
		ids[i] = string(id)
	}

	rows, err := executor.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetUsersByTeamName(ctx context.Context, teamName domain.TeamName) ([]domain.User, error) {
	executor := r.exec.DefaultTxOrDB(ctx)

	rows, err := executor.QueryContext(ctx,
		`SELECT user_id, username, team_name, is_active 
		 FROM users
		 WHERE team_name = $1`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
