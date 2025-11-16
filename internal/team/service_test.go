package team

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) CreateTeam(ctx context.Context, t domain.TeamName) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) Exists(ctx context.Context, name domain.TeamName) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

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

type MockTxManager struct {
	mock.Mock
}

// Ensure MockTxManager implements tx.ManagerInterface
var _ tx.ManagerInterface = (*MockTxManager)(nil)

func (m *MockTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// Execute the function directly for testing
	return fn(ctx)
}

func TestTeamService_CreateTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			txManager:      func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("team1")
		members := []domain.User{
			{ID: "user1", Username: "User 1", IsActive: true},
			{ID: "user2", Username: "User 2", IsActive: true},
		}

		mockTeamRepo.On("Exists", mock.Anything, teamName).Return(false, nil)
		mockTeamRepo.On("CreateTeam", mock.Anything, teamName).Return(nil)

		for _, member := range members {
			expectedMember := member
			expectedMember.TeamName = teamName
			mockUserRepo.On("Upsert", mock.Anything, &expectedMember).Return(nil)
		}

		team, err := service.CreateTeam(ctx, teamName, members)

		assert.NoError(t, err)
		assert.NotNil(t, team)
		assert.Equal(t, teamName, team.Name)

		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team already exists", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			txManager:      func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("team1")
		members := []domain.User{}

		mockTeamRepo.On("Exists", mock.Anything, teamName).Return(true, nil)

		team, err := service.CreateTeam(ctx, teamName, members)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTeamAlreadyExists, err)
		assert.Nil(t, team)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("error checking existence", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			txManager:      func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("team1")
		members := []domain.User{}

		mockTeamRepo.On("Exists", mock.Anything, teamName).Return(false, errors.New("db error"))

		team, err := service.CreateTeam(ctx, teamName, members)

		assert.Error(t, err)
		assert.Nil(t, team)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("error creating team", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			txManager:      func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("team1")
		members := []domain.User{}

		mockTeamRepo.On("Exists", mock.Anything, teamName).Return(false, nil)
		mockTeamRepo.On("CreateTeam", mock.Anything, teamName).Return(errors.New("db error"))

		team, err := service.CreateTeam(ctx, teamName, members)

		assert.Error(t, err)
		assert.Nil(t, team)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("error adding member", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			txManager:      func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("team1")
		members := []domain.User{
			{ID: "user1", Username: "User 1", IsActive: true},
		}

		mockTeamRepo.On("Exists", mock.Anything, teamName).Return(false, nil)
		mockTeamRepo.On("CreateTeam", mock.Anything, teamName).Return(nil)

		expectedMember := members[0]
		expectedMember.TeamName = teamName
		mockUserRepo.On("Upsert", mock.Anything, &expectedMember).Return(errors.New("db error"))

		team, err := service.CreateTeam(ctx, teamName, members)

		assert.Error(t, err)
		assert.Nil(t, team)

		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetTeamByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("team1")
		expectedTeam := &domain.Team{Name: teamName}

		mockTeamRepo.On("GetTeamByName", ctx, teamName).Return(expectedTeam, nil)

		team, err := service.GetTeamByName(ctx, teamName)

		assert.NoError(t, err)
		assert.NotNil(t, team)
		assert.Equal(t, teamName, team.Name)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := domain.TeamName("nonexistent")

		mockTeamRepo.On("GetTeamByName", ctx, teamName).Return(nil, domain.ErrNotFound)

		team, err := service.GetTeamByName(ctx, teamName)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, team)

		mockTeamRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetTeamMembers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := "team1"
		expectedUsers := []domain.User{
			{ID: "user1", Username: "User 1", TeamName: domain.TeamName(teamName), IsActive: true},
			{ID: "user2", Username: "User 2", TeamName: domain.TeamName(teamName), IsActive: true},
		}

		mockTeamRepo.On("Exists", mock.Anything, domain.TeamName(teamName)).Return(true, nil)
		mockUserRepo.On("GetUsersByTeamName", mock.Anything, domain.TeamName(teamName)).Return(expectedUsers, nil)

		users, err := service.GetTeamMembers(ctx, teamName)

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, expectedUsers, users)

		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := "nonexistent"

		mockTeamRepo.On("Exists", mock.Anything, domain.TeamName(teamName)).Return(false, nil)

		users, err := service.GetTeamMembers(ctx, teamName)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, users)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("error checking existence", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := "team1"

		mockTeamRepo.On("Exists", mock.Anything, domain.TeamName(teamName)).Return(false, errors.New("db error"))

		users, err := service.GetTeamMembers(ctx, teamName)

		assert.Error(t, err)
		assert.Nil(t, users)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("error fetching members", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &TeamService{
			teamRepository: mockTeamRepo,
			userRepository: mockUserRepo,
			logger:         mockLogger,
		}

		ctx := context.Background()
		teamName := "team1"

		mockTeamRepo.On("Exists", mock.Anything, domain.TeamName(teamName)).Return(true, nil)
		mockUserRepo.On("GetUsersByTeamName", mock.Anything, domain.TeamName(teamName)).Return(nil, sql.ErrConnDone)

		users, err := service.GetTeamMembers(ctx, teamName)

		assert.Error(t, err)
		assert.Nil(t, users)

		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
