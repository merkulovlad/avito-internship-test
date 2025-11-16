package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// Compile-time interface check
var _ domain.UserServiceInterface = (*UserService)(nil)

// UserService provides methods to manage users.
type UserService struct {
	prRepository   domain.PullRequestRepositoryInterface
	userRepository domain.UserRepositoryInterface
	txManager      *tx.Manager
}

func NewUserService(db *sql.DB, prRepository domain.PullRequestRepositoryInterface) *UserService {
	return &UserService{
		userRepository: NewUserRepository(db),
		prRepository:   prRepository,
		txManager:      tx.NewManager(db),
	}
}

func (s *UserService) SetIsActive(ctx context.Context, id domain.UserID, isActive bool) (*domain.User, error) {
	users, err := s.userRepository.SetUserIsActive(ctx, id, isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return users, nil
}

func (s *UserService) GetPrOfUser(ctx context.Context, userId domain.UserID) ([]domain.PullRequest, error) {
	// Check if user exists first
	exists, err := s.userRepository.Exists(ctx, userId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrNotFound
	}

	return s.prRepository.GetPrByUserID(ctx, userId)
}
