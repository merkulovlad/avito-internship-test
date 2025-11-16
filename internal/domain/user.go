package domain

import "context"

// UserID represents a unique user identifier.
type UserID string

// UserRepositoryInterface defines methods for user data persistence.
type UserRepositoryInterface interface {
	// Upsert creates or updates a user.
	Upsert(ctx context.Context, u *User) error
	// SetUserIsActive updates the active status of a user.
	SetUserIsActive(ctx context.Context, userId UserID, isActive bool) (*User, error)
	// GetActiveUsersByTeam retrieves all active user IDs for a given team.
	GetActiveUsersByTeam(ctx context.Context, teamName TeamName) ([]UserID, error)
	// Exists checks if a user with the given ID exists.
	Exists(ctx context.Context, userId UserID) (bool, error)
	// GetUserByID retrieves a user by their ID.
	GetUserByID(ctx context.Context, userId UserID) (*User, error)
	// GetUsersByIDs retrieves multiple users by their IDs.
	GetUsersByIDs(ctx context.Context, userIds []UserID) ([]User, error)
	// GetUsersByTeamName retrieves all users belonging to a team.
	GetUsersByTeamName(ctx context.Context, teamName TeamName) ([]User, error)
}

// User represents a user entity.
type User struct {
	// ID is the unique identifier for the user.
	ID UserID
	// Username is the display name of the user.
	Username string
	// TeamName is the team this user belongs to.
	TeamName TeamName
	// IsActive indicates whether the user is currently active.
	IsActive bool
}
