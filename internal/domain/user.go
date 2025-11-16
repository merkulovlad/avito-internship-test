package domain

import "context"

type UserID string

// UserRepositoryInterface defines methods for user data persistence
type UserRepositoryInterface interface {
	Upsert(ctx context.Context, u *User) error
	SetUserIsActive(ctx context.Context, userId UserID, isActive bool) (*User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName TeamName) ([]UserID, error)
	Exists(ctx context.Context, userId UserID) (bool, error)
	GetUserByID(ctx context.Context, userId UserID) (*User, error)
	GetUsersByTeamName(ctx context.Context, teamName TeamName) ([]User, error)
}

type User struct {
	ID       UserID
	Username string
	TeamName TeamName
	IsActive bool
}
