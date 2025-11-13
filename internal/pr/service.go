package pr

import "database/sql"

// PRService provides methods to manage pull requests.

type PRService struct {
	repository *sql.DB
}

func NewPRService(db *sql.DB) *PRService {
	return &PRService{
		repository: db,
	}
}