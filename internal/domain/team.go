package domain

import "context"

type TeamName string

// TeamRepositoryInterface defines methods for team data persistence
type TeamRepositoryInterface interface {
	CreateTeam(ctx context.Context, t TeamName) error
	GetTeamByName(ctx context.Context, name TeamName) (*Team, error)
	Exists(ctx context.Context, name TeamName) (bool, error)
}

type Team struct {
	Name TeamName
}
