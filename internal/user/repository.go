package user

import (
	"database/sql"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// UserService provides methods to manage users and their pull requests.

type UserService struct {
	repository *sql.DB
	txManager  *tx.Manager
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		repository: db,
		txManager:  tx.NewManager(db),
	}
}


