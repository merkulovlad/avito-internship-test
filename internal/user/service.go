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

func NewUserService(db *sql.DB, prRepository domain.PullRequestRepositoryInterface, logger logger.InterfaceLogger) *UserService {
	return &UserService{
		userRepository: NewUserRepository(db),
		prRepository:   prRepository,
		txManager:      tx.NewManager(db),
		logger:         logger,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, id domain.UserID, isActive bool) (*domain.User, error) {
	s.logger.Infof("Setting is_active=%v for user %s", isActive, id)

	users, err := s.userRepository.SetUserIsActive(ctx, id, isActive)
	if err != nil {
		s.logger.Errorf("Failed to set is_active=%v for user %s: %v", isActive, id, err)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}

	s.logger.Infof("Set is_active=%v for user %s successfully", isActive, id)

	return users, nil
}

func (s *UserService) GetPrOfUser(ctx context.Context, userId domain.UserID) ([]domain.PullRequest, error) {
	// Check if user exists first
	exists, err := s.userRepository.Exists(ctx, userId)
	if err != nil {
		s.logger.Errorf("Failed to check existence of user %s: %v", userId, err)
		return nil, err
	}

	s.logger.Infof("Checking existence of user %s", userId)

	if !exists {
		return nil, domain.ErrNotFound
	}

	s.logger.Infof("User %s exists", userId)

	return s.prRepository.GetPrByUserID(ctx, userId)
}
