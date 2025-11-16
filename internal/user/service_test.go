package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock PR Repository
type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) CreatePr(ctx context.Context, pullRequestId domain.PRID, authorId domain.UserID, title string) error {
	args := m.Called(ctx, pullRequestId, authorId, title)
	return args.Error(0)
}

func (m *MockPRRepository) AssignReviewer(ctx context.Context, pullRequestId domain.PRID, reviewer domain.UserID) error {
	args := m.Called(ctx, pullRequestId, reviewer)
	return args.Error(0)
}

func (m *MockPRRepository) MergePr(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, error) {
	args := m.Called(ctx, pullRequestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) Exists(ctx context.Context, pullRequestId domain.PRID) (bool, error) {
	args := m.Called(ctx, pullRequestId)
	return args.Bool(0), args.Error(1)
}

func (m *MockPRRepository) CheckIsMerged(ctx context.Context, pullRequestId domain.PRID) (bool, error) {
	args := m.Called(ctx, pullRequestId)
	return args.Bool(0), args.Error(1)
}

func (m *MockPRRepository) IsReviewerAssigned(ctx context.Context, pullRequestId domain.PRID, reviewerId domain.UserID) (bool, error) {
	args := m.Called(ctx, pullRequestId, reviewerId)
	return args.Bool(0), args.Error(1)
}

func (m *MockPRRepository) UnassignReviewer(ctx context.Context, pullRequestId domain.PRID, reviewerId domain.UserID) error {
	args := m.Called(ctx, pullRequestId, reviewerId)
	return args.Error(0)
}

func (m *MockPRRepository) GetPrByPrID(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, error) {
	args := m.Called(ctx, pullRequestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) GetPrByUserID(ctx context.Context, userId domain.UserID) ([]domain.PullRequest, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) GetReviewers(ctx context.Context, pullRequestId domain.PRID) ([]domain.UserID, error) {
	args := m.Called(ctx, pullRequestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]domain.UserID), args.Error(1)
}

func (m *MockPRRepository) GetAssignmentStats(ctx context.Context) (*domain.Stats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.Stats), args.Error(1)
}

// Mock User Repository (reuse from team tests or define here)
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Upsert(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) SetUserIsActive(ctx context.Context, userId domain.UserID, isActive bool) (*domain.User, error) {
	args := m.Called(ctx, userId, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsersByTeam(ctx context.Context, teamName domain.TeamName) ([]domain.UserID, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]domain.UserID), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, userId domain.UserID) (bool, error) {
	args := m.Called(ctx, userId)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userId domain.UserID) (*domain.User, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, userIds []domain.UserID) ([]domain.User, error) {
	args := m.Called(ctx, userIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByTeamName(ctx context.Context, teamName domain.TeamName) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]domain.User), args.Error(1)
}

// Mock TxManager
type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// Execute the function directly for testing
	return fn(ctx)
}

func TestUserService_SetIsActive(t *testing.T) {
	t.Run("success - set to false", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")
		expectedUser := &domain.User{
			ID:       userID,
			Username: "User 1",
			TeamName: "team1",
			IsActive: false,
		}

		mockUserRepo.On("SetUserIsActive", ctx, userID, false).Return(expectedUser, nil)

		user, err := service.SetIsActive(ctx, userID, false)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.False(t, user.IsActive)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("success - set to true", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")
		expectedUser := &domain.User{
			ID:       userID,
			Username: "User 1",
			TeamName: "team1",
			IsActive: true,
		}

		mockUserRepo.On("SetUserIsActive", ctx, userID, true).Return(expectedUser, nil)

		user, err := service.SetIsActive(ctx, userID, true)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.True(t, user.IsActive)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("nonexistent")

		mockUserRepo.On("SetUserIsActive", ctx, userID, true).Return(nil, sql.ErrNoRows)

		user, err := service.SetIsActive(ctx, userID, true)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, user)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")

		mockUserRepo.On("SetUserIsActive", ctx, userID, false).Return(nil, errors.New("db error"))

		user, err := service.SetIsActive(ctx, userID, false)

		assert.Error(t, err)
		assert.Nil(t, user)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetPrOfUser(t *testing.T) {
	t.Run("success - with PRs", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")
		expectedPRs := []domain.PullRequest{
			{ID: "pr1", AuthorID: "author1", Title: "PR 1"},
			{ID: "pr2", AuthorID: "author2", Title: "PR 2"},
		}

		mockUserRepo.On("Exists", ctx, userID).Return(true, nil)
		mockPRRepo.On("GetPrByUserID", ctx, userID).Return(expectedPRs, nil)

		prs, err := service.GetPrOfUser(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, prs, 2)
		assert.Equal(t, expectedPRs, prs)

		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("success - no PRs", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")

		mockUserRepo.On("Exists", ctx, userID).Return(true, nil)
		mockPRRepo.On("GetPrByUserID", ctx, userID).Return([]domain.PullRequest{}, nil)

		prs, err := service.GetPrOfUser(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, prs, 0)

		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("nonexistent")

		mockUserRepo.On("Exists", ctx, userID).Return(false, nil)

		prs, err := service.GetPrOfUser(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, prs)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("error checking existence", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")

		mockUserRepo.On("Exists", ctx, userID).Return(false, errors.New("db error"))

		prs, err := service.GetPrOfUser(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, prs)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("error fetching PRs", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockPRRepo := new(MockPRRepository)
		mockLogger := &logger.FakeLogger{}

		service := &UserService{
			prRepository:   mockPRRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		userID := domain.UserID("user1")

		mockUserRepo.On("Exists", ctx, userID).Return(true, nil)
		mockPRRepo.On("GetPrByUserID", ctx, userID).Return(nil, errors.New("db error"))

		prs, err := service.GetPrOfUser(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, prs)

		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})
}
