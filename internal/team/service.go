package team

import "database/sql"

// TeamService provides methods to manage teams and their members.

type TeamService struct {
	repository *sql.DB
}

func NewTeamService(db *sql.DB) *TeamService {
	return &TeamService{
		repository: db,
	}
}