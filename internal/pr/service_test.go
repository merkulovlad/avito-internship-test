package pr

import (
	"context"
	"testing"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
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

func TestPRService_CreatePr(t *testing.T) {
	t.Run("success - with reviewers", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		authorID := domain.UserID("user1")
		title := "Test PR"

		author := &domain.User{
			ID:       authorID,
			Username: "Author",
			TeamName: "team1",
			IsActive: true,
		}

		activeUsers := []domain.UserID{"user1", "user2", "user3"}
		reviewerUsers := []domain.User{
			{ID: "user2", Username: "Reviewer 1", TeamName: "team1", IsActive: true},
			{ID: "user3", Username: "Reviewer 2", TeamName: "team1", IsActive: true},
		}

		mockPRRepo.On("Exists", mock.Anything, prID).Return(false, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, authorID).Return(author, nil)
		mockPRRepo.On("CreatePr", mock.Anything, prID, authorID, title).Return(nil)
		mockUserRepo.On("GetActiveUsersByTeam", mock.Anything, author.TeamName).Return(activeUsers, nil)
		mockPRRepo.On("AssignReviewer", mock.Anything, prID, mock.Anything).Return(nil)
		mockUserRepo.On("GetUsersByIDs", mock.Anything, mock.Anything).Return(reviewerUsers, nil)

		pr, reviewers, err := service.CreatePr(ctx, prID, authorID, title)

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, prID, pr.ID)
		assert.Equal(t, authorID, pr.AuthorID)
		assert.Equal(t, title, pr.Title)
		assert.False(t, pr.IsMerged)
		assert.NotNil(t, reviewers)
		assert.Len(t, reviewers, 2)
	})

	t.Run("PR already exists", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		authorID := domain.UserID("user1")
		title := "Test PR"

		mockPRRepo.On("Exists", mock.Anything, prID).Return(true, nil)

		pr, reviewers, err := service.CreatePr(ctx, prID, authorID, title)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPrAlreadyExists, err)
		assert.Nil(t, pr)
		assert.Nil(t, reviewers)
	})
}

func TestPRService_MergePr(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")

		mergedPR := &domain.PullRequest{
			ID:       prID,
			AuthorID: "user1",
			Title:    "Test PR",
			IsMerged: true,
		}

		reviewerIDs := []domain.UserID{"user2", "user3"}
		reviewerUsers := []domain.User{
			{ID: "user2", Username: "Reviewer 1"},
			{ID: "user3", Username: "Reviewer 2"},
		}

		mockPRRepo.On("Exists", mock.Anything, prID).Return(true, nil)
		mockPRRepo.On("MergePr", mock.Anything, prID).Return(mergedPR, nil)
		mockPRRepo.On("GetReviewers", mock.Anything, prID).Return(reviewerIDs, nil)
		mockUserRepo.On("GetUsersByIDs", mock.Anything, reviewerIDs).Return(reviewerUsers, nil)

		pr, reviewers, err := service.MergePr(ctx, prID)

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.True(t, pr.IsMerged)
		assert.Len(t, reviewers, 2)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("nonexistent")

		mockPRRepo.On("Exists", mock.Anything, prID).Return(false, nil)

		pr, reviewers, err := service.MergePr(ctx, prID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, pr)
		assert.Nil(t, reviewers)
	})
}

func TestPRService_GetReviewers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:   mockPRRepo,
			userRepo: mockUserRepo,
			logger:   mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")

		reviewerIDs := []domain.UserID{"user2", "user3"}
		reviewerUsers := []domain.User{
			{ID: "user2", Username: "Reviewer 1"},
			{ID: "user3", Username: "Reviewer 2"},
		}

		mockPRRepo.On("Exists", mock.Anything, prID).Return(true, nil)
		mockPRRepo.On("GetReviewers", mock.Anything, prID).Return(reviewerIDs, nil)
		mockUserRepo.On("GetUsersByIDs", mock.Anything, reviewerIDs).Return(reviewerUsers, nil)

		reviewers, err := service.GetReviewers(ctx, prID)

		assert.NoError(t, err)
		assert.Len(t, reviewers, 2)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:   mockPRRepo,
			userRepo: mockUserRepo,
			logger:   mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("nonexistent")

		mockPRRepo.On("Exists", mock.Anything, prID).Return(false, nil)

		reviewers, err := service.GetReviewers(ctx, prID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, reviewers)
	})
}

func TestPRService_ReassignReviewer(t *testing.T) {
	t.Run("success - reassign reviewer", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("user2")
		authorID := domain.UserID("user1")

		pr := &domain.PullRequest{
			ID:                prID,
			AuthorID:          authorID,
			Title:             "Test PR",
			IsMerged:          false,
			AssignedReviewers: []domain.UserID{oldReviewerID, "user3"},
		}

		oldReviewer := &domain.User{
			ID:       oldReviewerID,
			Username: "Old Reviewer",
			TeamName: "team1",
			IsActive: true,
		}

		activeUsers := []domain.UserID{"user1", "user2", "user3", "user4"}

		// Pre-checks
		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(true, nil)
		mockPRRepo.On("CheckIsMerged", ctx, prID).Return(false, nil)
		mockPRRepo.On("IsReviewerAssigned", ctx, prID, oldReviewerID).Return(true, nil)

		// Transaction operations
		mockPRRepo.On("GetPrByPrID", mock.Anything, prID).Return(pr, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, oldReviewerID).Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveUsersByTeam", mock.Anything, oldReviewer.TeamName).Return(activeUsers, nil)
		mockPRRepo.On("UnassignReviewer", mock.Anything, prID, oldReviewerID).Return(nil)
		mockPRRepo.On("AssignReviewer", mock.Anything, prID, mock.Anything).Return(nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.PR)
		assert.NotEmpty(t, result.ReplacedBy)
		// New reviewer should not be the author or old reviewer
		assert.NotEqual(t, authorID, result.ReplacedBy)
		assert.NotEqual(t, oldReviewerID, result.ReplacedBy)
		// New reviewer should be from the active users
		assert.Contains(t, activeUsers, result.ReplacedBy)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:   mockPRRepo,
			userRepo: mockUserRepo,
			logger:   mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("nonexistent")
		oldReviewerID := domain.UserID("user2")

		mockPRRepo.On("Exists", ctx, prID).Return(false, nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:   mockPRRepo,
			userRepo: mockUserRepo,
			logger:   mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("nonexistent")

		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(false, nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("PR already merged", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:   mockPRRepo,
			userRepo: mockUserRepo,
			logger:   mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("user2")

		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(true, nil)
		mockPRRepo.On("CheckIsMerged", ctx, prID).Return(true, nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPrMerged, err)
		assert.Nil(t, result)
	})

	t.Run("reviewer not assigned to PR", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:   mockPRRepo,
			userRepo: mockUserRepo,
			logger:   mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("user2")

		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(true, nil)
		mockPRRepo.On("CheckIsMerged", ctx, prID).Return(false, nil)
		mockPRRepo.On("IsReviewerAssigned", ctx, prID, oldReviewerID).Return(false, nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotAssigned, err)
		assert.Nil(t, result)
	})

	t.Run("no candidates available", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(false); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("user2")
		authorID := domain.UserID("user1")

		pr := &domain.PullRequest{
			ID:                prID,
			AuthorID:          authorID,
			Title:             "Test PR",
			IsMerged:          false,
			AssignedReviewers: []domain.UserID{oldReviewerID},
		}

		oldReviewer := &domain.User{
			ID:       oldReviewerID,
			Username: "Old Reviewer",
			TeamName: "team1",
			IsActive: true,
		}

		// Only author and old reviewer in the team
		activeUsers := []domain.UserID{"user1", "user2"}

		// Pre-checks
		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(true, nil)
		mockPRRepo.On("CheckIsMerged", ctx, prID).Return(false, nil)
		mockPRRepo.On("IsReviewerAssigned", ctx, prID, oldReviewerID).Return(true, nil)

		// Transaction operations
		mockPRRepo.On("GetPrByPrID", mock.Anything, prID).Return(pr, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, oldReviewerID).Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveUsersByTeam", mock.Anything, oldReviewer.TeamName).Return(activeUsers, nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNoCandidate, err)
		assert.Nil(t, result)
	})

	t.Run("success - single candidate becomes new reviewer", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("user2")
		authorID := domain.UserID("user1")
		expectedNewReviewer := domain.UserID("user3")

		pr := &domain.PullRequest{
			ID:                prID,
			AuthorID:          authorID,
			Title:             "Test PR",
			IsMerged:          false,
			AssignedReviewers: []domain.UserID{oldReviewerID},
		}

		oldReviewer := &domain.User{
			ID:       oldReviewerID,
			Username: "Old Reviewer",
			TeamName: "team1",
			IsActive: true,
		}

		// Only one candidate available (user3)
		activeUsers := []domain.UserID{"user1", "user2", "user3"}

		// Pre-checks
		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(true, nil)
		mockPRRepo.On("CheckIsMerged", ctx, prID).Return(false, nil)
		mockPRRepo.On("IsReviewerAssigned", ctx, prID, oldReviewerID).Return(true, nil)

		// Transaction operations
		mockPRRepo.On("GetPrByPrID", mock.Anything, prID).Return(pr, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, oldReviewerID).Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveUsersByTeam", mock.Anything, oldReviewer.TeamName).Return(activeUsers, nil)
		mockPRRepo.On("UnassignReviewer", mock.Anything, prID, oldReviewerID).Return(nil)
		mockPRRepo.On("AssignReviewer", mock.Anything, prID, expectedNewReviewer).Return(nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.PR)
		assert.Equal(t, expectedNewReviewer, result.ReplacedBy)
		// Verify the updated reviewers list doesn't contain old reviewer
		assert.NotContains(t, result.PR.AssignedReviewers, oldReviewerID)
		assert.Contains(t, result.PR.AssignedReviewers, expectedNewReviewer)
	})

	t.Run("success - multiple reviewers, one gets reassigned", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()
		prID := domain.PRID("pr1")
		oldReviewerID := domain.UserID("user2")
		otherReviewerID := domain.UserID("user3")
		authorID := domain.UserID("user1")

		pr := &domain.PullRequest{
			ID:                prID,
			AuthorID:          authorID,
			Title:             "Test PR",
			IsMerged:          false,
			AssignedReviewers: []domain.UserID{oldReviewerID, otherReviewerID},
		}

		oldReviewer := &domain.User{
			ID:       oldReviewerID,
			Username: "Old Reviewer",
			TeamName: "team1",
			IsActive: true,
		}

		activeUsers := []domain.UserID{"user1", "user2", "user3", "user4", "user5"}

		// Pre-checks
		mockPRRepo.On("Exists", ctx, prID).Return(true, nil)
		mockUserRepo.On("Exists", ctx, oldReviewerID).Return(true, nil)
		mockPRRepo.On("CheckIsMerged", ctx, prID).Return(false, nil)
		mockPRRepo.On("IsReviewerAssigned", ctx, prID, oldReviewerID).Return(true, nil)

		// Transaction operations
		mockPRRepo.On("GetPrByPrID", mock.Anything, prID).Return(pr, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, oldReviewerID).Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveUsersByTeam", mock.Anything, oldReviewer.TeamName).Return(activeUsers, nil)
		mockPRRepo.On("UnassignReviewer", mock.Anything, prID, oldReviewerID).Return(nil)
		mockPRRepo.On("AssignReviewer", mock.Anything, prID, mock.Anything).Return(nil)

		result, err := service.ReassignReviewer(ctx, prID, oldReviewerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.PR)
		// Old reviewer should be removed
		assert.NotContains(t, result.PR.AssignedReviewers, oldReviewerID)
		// Other reviewer should still be there
		assert.Contains(t, result.PR.AssignedReviewers, otherReviewerID)
		// New reviewer should be added
		assert.Contains(t, result.PR.AssignedReviewers, result.ReplacedBy)
		// Should still have 2 reviewers total
		assert.Len(t, result.PR.AssignedReviewers, 2)
	})
}

func TestPRService_PickReviewers(t *testing.T) {
	mockLogger := &logger.FakeLogger{}
	service := &PRService{
		logger: mockLogger,
	}

	t.Run("no candidates", func(t *testing.T) {
		candidates := []domain.UserID{}
		result := service.pickReviewers(candidates, 2)
		assert.Nil(t, result)
	})

	t.Run("fewer candidates than max", func(t *testing.T) {
		candidates := []domain.UserID{"user1"}
		result := service.pickReviewers(candidates, 2)
		assert.Len(t, result, 1)
		assert.Equal(t, domain.UserID("user1"), result[0])
	})

	t.Run("more candidates than max", func(t *testing.T) {
		candidates := []domain.UserID{"user1", "user2", "user3", "user4", "user5"}
		result := service.pickReviewers(candidates, 2)
		assert.Len(t, result, 2)
		// Verify that selected reviewers are from the candidates
		for _, r := range result {
			assert.Contains(t, candidates, r)
		}
	})

	t.Run("equal candidates and max", func(t *testing.T) {
		candidates := []domain.UserID{"user1", "user2"}
		result := service.pickReviewers(candidates, 2)
		assert.Len(t, result, 2)
	})
}

func TestPRService_GetAssignmentStats(t *testing.T) {
	t.Run("success - with stats", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()

		expectedStats := &domain.Stats{
			ByUser: []domain.UserAssignmentStat{
				{UserID: "user1", Assignments: 10},
				{UserID: "user2", Assignments: 5},
				{UserID: "user3", Assignments: 3},
			},
			ByPR: []domain.PRReviewerStat{
				{PullRequestID: "pr-1001", Reviewers: 2},
				{PullRequestID: "pr-1002", Reviewers: 1},
				{PullRequestID: "pr-1003", Reviewers: 2},
			},
		}

		mockPRRepo.On("GetAssignmentStats", ctx).Return(expectedStats, nil)

		stats, err := service.GetAssignmentStats(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, expectedStats, stats)

		// Verify user stats
		assert.Len(t, stats.ByUser, 3)
		assert.Equal(t, domain.UserID("user1"), stats.ByUser[0].UserID)
		assert.Equal(t, 10, stats.ByUser[0].Assignments)
		assert.Equal(t, domain.UserID("user2"), stats.ByUser[1].UserID)
		assert.Equal(t, 5, stats.ByUser[1].Assignments)

		// Verify PR stats
		assert.Len(t, stats.ByPR, 3)
		assert.Equal(t, domain.PRID("pr-1001"), stats.ByPR[0].PullRequestID)
		assert.Equal(t, 2, stats.ByPR[0].Reviewers)

		mockPRRepo.AssertExpectations(t)
	})

	t.Run("success - empty stats", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()

		expectedStats := &domain.Stats{
			ByUser: []domain.UserAssignmentStat{},
			ByPR:   []domain.PRReviewerStat{},
		}

		mockPRRepo.On("GetAssignmentStats", ctx).Return(expectedStats, nil)

		stats, err := service.GetAssignmentStats(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Len(t, stats.ByUser, 0)
		assert.Len(t, stats.ByPR, 0)

		mockPRRepo.AssertExpectations(t)
	})

	t.Run("error - repository fails", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockLogger := &logger.FakeLogger{}

		service := &PRService{
			prRepo:    mockPRRepo,
			userRepo:  mockUserRepo,
			txManager: func() *tx.Manager { m, _, _ := newTestTxManager(true); return m }(),
			logger:    mockLogger,
		}

		ctx := context.Background()

		mockPRRepo.On("GetAssignmentStats", ctx).Return(nil, assert.AnError)

		stats, err := service.GetAssignmentStats(ctx)

		assert.Error(t, err)
		assert.Nil(t, stats)

		mockPRRepo.AssertExpectations(t)
	})
}
