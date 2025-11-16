package pr

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRRepository_CreatePr(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prID := domain.PRID("pr1")
		authorID := domain.UserID("user1")
		title := "Test PR"

		mock.ExpectExec("INSERT INTO pull_requests").
			WithArgs(prID, authorID, title).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreatePr(ctx, prID, authorID, title)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		prID := domain.PRID("pr1")
		authorID := domain.UserID("user1")
		title := "Test PR"

		mock.ExpectExec("INSERT INTO pull_requests").
			WithArgs(prID, authorID, title).
			WillReturnError(sql.ErrConnDone)

		err := repo.CreatePr(ctx, prID, authorID, title)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_AssignReviewer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prID := domain.PRID("pr1")
		reviewerID := domain.UserID("user1")

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(prID, reviewerID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AssignReviewer(ctx, prID, reviewerID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		prID := domain.PRID("pr1")
		reviewerID := domain.UserID("user1")

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(prID, reviewerID).
			WillReturnError(sql.ErrConnDone)

		err := repo.AssignReviewer(ctx, prID, reviewerID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_MergePr(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prID := domain.PRID("pr1")
		mergedAt := time.Now()

		mock.ExpectExec("UPDATE pull_requests SET is_merged = TRUE").
			WithArgs(prID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		rows := sqlmock.NewRows([]string{"pull_request_id", "title", "author_id", "is_merged", "merged_at"}).
			AddRow(prID, "Test PR", "user1", true, mergedAt)

		mock.ExpectQuery("SELECT pull_request_id, title, author_id, is_merged, merged_at").
			WithArgs(prID).
			WillReturnRows(rows)

		reviewerRows := sqlmock.NewRows([]string{"user_id"}).
			AddRow("user2").
			AddRow("user3")

		mock.ExpectQuery("SELECT user_id FROM pr_reviewers").
			WithArgs(prID).
			WillReturnRows(reviewerRows)

		pr, err := repo.MergePr(ctx, prID)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, prID, pr.ID)
		assert.True(t, pr.IsMerged)
		assert.NotNil(t, pr.MergedAt)
		assert.Len(t, pr.AssignedReviewers, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		prID := domain.PRID("pr1")

		mock.ExpectExec("UPDATE pull_requests SET is_merged = TRUE").
			WithArgs(prID).
			WillReturnError(sql.ErrConnDone)

		pr, err := repo.MergePr(ctx, prID)
		assert.Error(t, err)
		assert.Nil(t, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("exists", func(t *testing.T) {
		prID := domain.PRID("pr1")

		rows := sqlmock.NewRows([]string{"exists"}).
			AddRow(true)

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM pull_requests WHERE pull_request_id = \\$1\\)").
			WithArgs(prID).
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, prID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("does not exist", func(t *testing.T) {
		prID := domain.PRID("nonexistent")

		rows := sqlmock.NewRows([]string{"exists"}).
			AddRow(false)

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM pull_requests WHERE pull_request_id = \\$1\\)").
			WithArgs(prID).
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, prID)
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_CheckIsMerged(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("is merged", func(t *testing.T) {
		prID := domain.PRID("pr1")

		rows := sqlmock.NewRows([]string{"is_merged"}).
			AddRow(true)

		mock.ExpectQuery("SELECT is_merged FROM pull_requests WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnRows(rows)

		isMerged, err := repo.CheckIsMerged(ctx, prID)
		assert.NoError(t, err)
		assert.True(t, isMerged)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("is not merged", func(t *testing.T) {
		prID := domain.PRID("pr1")

		rows := sqlmock.NewRows([]string{"is_merged"}).
			AddRow(false)

		mock.ExpectQuery("SELECT is_merged FROM pull_requests WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnRows(rows)

		isMerged, err := repo.CheckIsMerged(ctx, prID)
		assert.NoError(t, err)
		assert.False(t, isMerged)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_IsReviewerAssigned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("is assigned", func(t *testing.T) {
		prID := domain.PRID("pr1")
		reviewerID := domain.UserID("user1")

		rows := sqlmock.NewRows([]string{"exists"}).
			AddRow(true)

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM pr_reviewers").
			WithArgs(prID, reviewerID).
			WillReturnRows(rows)

		isAssigned, err := repo.IsReviewerAssigned(ctx, prID, reviewerID)
		assert.NoError(t, err)
		assert.True(t, isAssigned)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("is not assigned", func(t *testing.T) {
		prID := domain.PRID("pr1")
		reviewerID := domain.UserID("user1")

		rows := sqlmock.NewRows([]string{"exists"}).
			AddRow(false)

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM pr_reviewers").
			WithArgs(prID, reviewerID).
			WillReturnRows(rows)

		isAssigned, err := repo.IsReviewerAssigned(ctx, prID, reviewerID)
		assert.NoError(t, err)
		assert.False(t, isAssigned)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_UnassignReviewer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prID := domain.PRID("pr1")
		reviewerID := domain.UserID("user1")

		mock.ExpectExec("DELETE FROM pr_reviewers WHERE pull_request_id = \\$1 AND user_id = \\$2").
			WithArgs(prID, reviewerID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UnassignReviewer(ctx, prID, reviewerID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		prID := domain.PRID("pr1")
		reviewerID := domain.UserID("user1")

		mock.ExpectExec("DELETE FROM pr_reviewers WHERE pull_request_id = \\$1 AND user_id = \\$2").
			WithArgs(prID, reviewerID).
			WillReturnError(sql.ErrConnDone)

		err := repo.UnassignReviewer(ctx, prID, reviewerID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_GetPrByPrID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prID := domain.PRID("pr1")
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{"pull_request_id", "author_id", "title", "created_at", "is_merged", "merged_at"}).
			AddRow(prID, "user1", "Test PR", createdAt, false, nil)

		mock.ExpectQuery("SELECT pull_request_id, author_id, title, created_at, is_merged, merged_at FROM pull_requests WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnRows(rows)

		pr, err := repo.GetPrByPrID(ctx, prID)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, prID, pr.ID)
		assert.Equal(t, "Test PR", pr.Title)
		assert.False(t, pr.IsMerged)
		assert.Nil(t, pr.MergedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		prID := domain.PRID("nonexistent")

		mock.ExpectQuery("SELECT pull_request_id, author_id, title, created_at, is_merged, merged_at FROM pull_requests WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnError(sql.ErrNoRows)

		pr, err := repo.GetPrByPrID(ctx, prID)
		assert.Error(t, err)
		assert.Nil(t, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_GetPrByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success - multiple PRs", func(t *testing.T) {
		userID := domain.UserID("user1")

		rows := sqlmock.NewRows([]string{"pull_request_id", "author_id", "title", "created_at", "is_merged", "merged_at"}).
			AddRow("pr1", "author1", "PR 1", time.Now(), false, nil).
			AddRow("pr2", "author2", "PR 2", time.Now(), true, time.Now())

		mock.ExpectQuery("SELECT pr.pull_request_id, pr.author_id, pr.title, pr.created_at, pr.is_merged, pr.merged_at").
			WithArgs(userID).
			WillReturnRows(rows)

		prs, err := repo.GetPrByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, prs, 2)
		assert.Equal(t, domain.PRID("pr1"), prs[0].ID)
		assert.Equal(t, domain.PRID("pr2"), prs[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - no PRs", func(t *testing.T) {
		userID := domain.UserID("user1")

		rows := sqlmock.NewRows([]string{"pull_request_id", "author_id", "title", "created_at", "is_merged", "merged_at"})

		mock.ExpectQuery("SELECT pr.pull_request_id, pr.author_id, pr.title, pr.created_at, pr.is_merged, pr.merged_at").
			WithArgs(userID).
			WillReturnRows(rows)

		prs, err := repo.GetPrByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, prs, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		userID := domain.UserID("user1")

		mock.ExpectQuery("SELECT pr.pull_request_id, pr.author_id, pr.title, pr.created_at, pr.is_merged, pr.merged_at").
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		prs, err := repo.GetPrByUserID(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, prs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_GetReviewers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success - multiple reviewers", func(t *testing.T) {
		prID := domain.PRID("pr1")

		rows := sqlmock.NewRows([]string{"user_id"}).
			AddRow("user1").
			AddRow("user2")

		mock.ExpectQuery("SELECT user_id FROM pr_reviewers WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnRows(rows)

		reviewers, err := repo.GetReviewers(ctx, prID)
		assert.NoError(t, err)
		assert.Len(t, reviewers, 2)
		assert.Equal(t, domain.UserID("user1"), reviewers[0])
		assert.Equal(t, domain.UserID("user2"), reviewers[1])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - no reviewers", func(t *testing.T) {
		prID := domain.PRID("pr1")

		rows := sqlmock.NewRows([]string{"user_id"})

		mock.ExpectQuery("SELECT user_id FROM pr_reviewers WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnRows(rows)

		reviewers, err := repo.GetReviewers(ctx, prID)
		assert.NoError(t, err)
		assert.Len(t, reviewers, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		prID := domain.PRID("pr1")

		mock.ExpectQuery("SELECT user_id FROM pr_reviewers WHERE pull_request_id = \\$1").
			WithArgs(prID).
			WillReturnError(sql.ErrConnDone)

		reviewers, err := repo.GetReviewers(ctx, prID)
		assert.Error(t, err)
		assert.Nil(t, reviewers)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepository_GetAssignmentStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	repo := NewPRRepository(db)
	ctx := context.Background()

	t.Run("success - with data", func(t *testing.T) {
		// Mock user assignments query
		userRows := sqlmock.NewRows([]string{"user_id", "assignments"}).
			AddRow("user1", 10).
			AddRow("user2", 5).
			AddRow("user3", 3)

		mock.ExpectQuery("SELECT user_id, COUNT\\(\\*\\) as assignments FROM pr_reviewers GROUP BY user_id").
			WillReturnRows(userRows)

		// Mock PR reviewers query
		prRows := sqlmock.NewRows([]string{"pull_request_id", "reviewers"}).
			AddRow("pr-1001", 2).
			AddRow("pr-1002", 1).
			AddRow("pr-1003", 2)

		mock.ExpectQuery("SELECT pull_request_id, COUNT\\(\\*\\) as reviewers FROM pr_reviewers GROUP BY pull_request_id").
			WillReturnRows(prRows)

		stats, err := repo.GetAssignmentStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify user stats
		assert.Len(t, stats.ByUser, 3)
		assert.Equal(t, domain.UserID("user1"), stats.ByUser[0].UserID)
		assert.Equal(t, 10, stats.ByUser[0].Assignments)
		assert.Equal(t, domain.UserID("user2"), stats.ByUser[1].UserID)
		assert.Equal(t, 5, stats.ByUser[1].Assignments)
		assert.Equal(t, domain.UserID("user3"), stats.ByUser[2].UserID)
		assert.Equal(t, 3, stats.ByUser[2].Assignments)

		// Verify PR stats
		assert.Len(t, stats.ByPR, 3)
		assert.Equal(t, domain.PRID("pr-1001"), stats.ByPR[0].PullRequestID)
		assert.Equal(t, 2, stats.ByPR[0].Reviewers)
		assert.Equal(t, domain.PRID("pr-1002"), stats.ByPR[1].PullRequestID)
		assert.Equal(t, 1, stats.ByPR[1].Reviewers)
		assert.Equal(t, domain.PRID("pr-1003"), stats.ByPR[2].PullRequestID)
		assert.Equal(t, 2, stats.ByPR[2].Reviewers)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - empty data", func(t *testing.T) {
		// Mock empty user assignments query
		userRows := sqlmock.NewRows([]string{"user_id", "assignments"})

		mock.ExpectQuery("SELECT user_id, COUNT\\(\\*\\) as assignments FROM pr_reviewers GROUP BY user_id").
			WillReturnRows(userRows)

		// Mock empty PR reviewers query
		prRows := sqlmock.NewRows([]string{"pull_request_id", "reviewers"})

		mock.ExpectQuery("SELECT pull_request_id, COUNT\\(\\*\\) as reviewers FROM pr_reviewers GROUP BY pull_request_id").
			WillReturnRows(prRows)

		stats, err := repo.GetAssignmentStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Len(t, stats.ByUser, 0)
		assert.Len(t, stats.ByPR, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - user query fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT user_id, COUNT\\(\\*\\) as assignments FROM pr_reviewers GROUP BY user_id").
			WillReturnError(sql.ErrConnDone)

		stats, err := repo.GetAssignmentStats(ctx)
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - pr query fails", func(t *testing.T) {
		// Mock successful user assignments query
		userRows := sqlmock.NewRows([]string{"user_id", "assignments"}).
			AddRow("user1", 10)

		mock.ExpectQuery("SELECT user_id, COUNT\\(\\*\\) as assignments FROM pr_reviewers GROUP BY user_id").
			WillReturnRows(userRows)

		// Mock failed PR reviewers query
		mock.ExpectQuery("SELECT pull_request_id, COUNT\\(\\*\\) as reviewers FROM pr_reviewers GROUP BY pull_request_id").
			WillReturnError(sql.ErrConnDone)

		stats, err := repo.GetAssignmentStats(ctx)
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - user scan fails", func(t *testing.T) {
		// Mock user assignments query with wrong column type
		userRows := sqlmock.NewRows([]string{"user_id", "assignments"}).
			AddRow("user1", "invalid") // string instead of int

		mock.ExpectQuery("SELECT user_id, COUNT\\(\\*\\) as assignments FROM pr_reviewers GROUP BY user_id").
			WillReturnRows(userRows)

		stats, err := repo.GetAssignmentStats(ctx)
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - pr scan fails", func(t *testing.T) {
		// Mock successful user assignments query
		userRows := sqlmock.NewRows([]string{"user_id", "assignments"}).
			AddRow("user1", 10)

		mock.ExpectQuery("SELECT user_id, COUNT\\(\\*\\) as assignments FROM pr_reviewers GROUP BY user_id").
			WillReturnRows(userRows)

		// Mock PR reviewers query with wrong column type
		prRows := sqlmock.NewRows([]string{"pull_request_id", "reviewers"}).
			AddRow("pr-1001", "invalid") // string instead of int

		mock.ExpectQuery("SELECT pull_request_id, COUNT\\(\\*\\) as reviewers FROM pr_reviewers GROUP BY pull_request_id").
			WillReturnRows(prRows)

		stats, err := repo.GetAssignmentStats(ctx)
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
