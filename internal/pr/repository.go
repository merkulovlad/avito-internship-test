package pr

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

type PRRepository struct {
	exec *tx.ExecutorImpl
}

func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{
		exec: tx.NewExecutor(db),
	}
}

func (r *PRRepository) CreatePr(ctx context.Context, pullRequestId domain.PRID, authorId domain.UserID, title string) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx, "INSERT INTO pull_requests (pull_request_id, user_id, title, created_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)", pullRequestId, authorId, title)

	return err
}

func (r *PRRepository) AssignReviewer(ctx context.Context, pullRequestId domain.PRID, reviewer domain.UserID) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx,
		`INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
     VALUES ($1, $2)
     ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING`,
		pullRequestId, reviewer,
	)

	return err
}

func (r *PRRepository) MergePr(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, error) {
	executor := r.exec.DefaultTxOrDB(ctx)

	_, err := executor.ExecContext(ctx,
		"UPDATE pull_requests SET is_merged = TRUE, merged_at = CURRENT_TIMESTAMP WHERE pull_request_id = $1 AND is_merged = FALSE",
		pullRequestId,
	)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRowContext(ctx, `
        SELECT pull_request_id, title, author_id, is_merged, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, pullRequestId)

	var pr domain.PullRequest
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.IsMerged, &pr.MergedAt); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PRRepository) Exists(ctx context.Context, pullRequestId domain.PRID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", pullRequestId)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PRRepository) CheckIsMerged(ctx context.Context, pullRequestId domain.PRID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT is_merged FROM pull_requests WHERE pull_request_id = $1", pullRequestId)

	var isMerged bool
	if err := row.Scan(&isMerged); err != nil {
		return false, err
	}
	return isMerged, nil
}

func (r *PRRepository) IsReviewerAssigned(ctx context.Context, pullRequestId domain.PRID, reviewId domain.UserID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2)", pullRequestId, reviewId)

	var isAssigned bool
	if err := row.Scan(&isAssigned); err != nil {
		return false, err
	}

	return isAssigned, nil
}

func (r *PRRepository) UnassignReviewer(ctx context.Context, pullRequestId domain.PRID, reviewerId domain.UserID) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx, "DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2", pullRequestId, reviewerId)

	return err
}

func (r *PRRepository) GetPrByID(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT pull_request_id, author_id, title, created_at, is_merged, merged_at FROM pull_requests WHERE pull_request_id = $1", pullRequestId)

	var pr domain.PullRequest
	if err := row.Scan(&pr.ID, &pr.AuthorID, &pr.Title, &pr.CreatedAt, &pr.IsMerged, &pr.MergedAt); err != nil {
		return nil, err
	}

	return &pr, nil
}
