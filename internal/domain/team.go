// Package domain defines core business entities and interfaces for the application.
package domain

import "context"

// TeamName represents a unique team identifier.
type TeamName string

// TeamRepositoryInterface defines methods for team data persistence.
type TeamRepositoryInterface interface {
	// CreateTeam creates a new team with the given name.
	CreateTeam(ctx context.Context, t TeamName) error
	// GetTeamByName retrieves a team by its name.
	GetTeamByName(ctx context.Context, name TeamName) (*Team, error)
	// Exists checks if a team with the given name exists.
	Exists(ctx context.Context, name TeamName) (bool, error)
}

// Team represents a team entity.
type Team struct {
	Name TeamName
}
