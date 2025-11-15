package user

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// UserService provides methods to manage users.
type UserService struct {
	repository *UserRepository
	txManager  *tx.Manager
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		repository: NewUserRepository(db),
		txManager:  tx.NewManager(db),
	}
}

func (s *UserService) SetIsActive(ctx context.Context, u *domain.User) (*domain.User, error) {
	id := u.ID
	return s.repository.SetUserIsActive(ctx, id, u.IsActive)
}
