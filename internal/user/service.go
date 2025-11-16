// Package user implements user management business logic.
package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// Compile-time interface check
var _ domain.UserServiceInterface = (*UserService)(nil)

// UserService provides methods to manage users.
type UserService struct {
	prRepository   domain.PullRequestRepositoryInterface
	userRepository domain.UserRepositoryInterface
	txManager      *tx.Manager
	logger         logger.InterfaceLogger
}

// NewUserService creates a new UserService instance.
func NewUserService(db *sql.DB, prRepository domain.PullRequestRepositoryInterface, logger logger.InterfaceLogger) *UserService {
	return &UserService{
		userRepository: NewUserRepository(db),
		prRepository:   prRepository,
		txManager:      tx.NewManager(db),
		logger:         logger,
	}
}

// SetIsActive updates the active status of a user.
func (s *UserService) SetIsActive(ctx context.Context, id domain.UserID, isActive bool) (*domain.User, error) {
	users, err := s.userRepository.SetUserIsActive(ctx, id, isActive)
	if err != nil {
		s.logger.Errorf("Failed to set is_active=%v for user %s: %v", isActive, id, err)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}

	return users, nil
}

// GetPrOfUser retrieves all pull requests assigned to a user.
func (s *UserService) GetPrOfUser(ctx context.Context, userId domain.UserID) ([]domain.PullRequest, error) {
	exists, err := s.userRepository.Exists(ctx, userId)
	if err != nil {
		s.logger.Errorf("Failed to check existence of user %s: %v", userId, err)
		return nil, err
	}

	if !exists {
		return nil, domain.ErrNotFound
	}

	return s.prRepository.GetPrByUserID(ctx, userId)
}
