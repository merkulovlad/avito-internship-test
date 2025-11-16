// Package pr implements pull request management business logic.
package pr

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
)

// Compile-time interface check
var _ domain.PullRequestRepositoryInterface = (*PRRepository)(nil)

// PRRepository provides database operations for pull requests.
type PRRepository struct {
	exec *tx.ExecutorImpl
}

// NewPRRepository creates a new PRRepository instance.
func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{
		exec: tx.NewExecutor(db),
	}
}

// CreatePr inserts a new pull request into the database.
func (r *PRRepository) CreatePr(ctx context.Context, pullRequestId domain.PRID, authorId domain.UserID, title string) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx, "INSERT INTO pull_requests (pull_request_id, author_id, title, created_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)", pullRequestId, authorId, title)

	return err
}

// AssignReviewer assigns a reviewer to a pull request.
func (r *PRRepository) AssignReviewer(ctx context.Context, pullRequestId domain.PRID, reviewer domain.UserID) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx,
		`INSERT INTO pr_reviewers (pull_request_id, user_id)
     VALUES ($1, $2)
     ON CONFLICT (pull_request_id, user_id) DO NOTHING`,
		pullRequestId, reviewer,
	)

	return err
}

// MergePr marks a pull request as merged in the database.
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

	// Fetch assigned reviewers
	rows, err := executor.QueryContext(ctx, `
		SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
	`, pullRequestId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	pr.AssignedReviewers = []domain.UserID{}

	for rows.Next() {
		var reviewerID domain.UserID
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}

		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &pr, nil
}

// Exists checks if a pull request with the given ID exists in the database.
func (r *PRRepository) Exists(ctx context.Context, pullRequestId domain.PRID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", pullRequestId)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// CheckIsMerged checks if a pull request is merged.
func (r *PRRepository) CheckIsMerged(ctx context.Context, pullRequestId domain.PRID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT is_merged FROM pull_requests WHERE pull_request_id = $1", pullRequestId)

	var isMerged bool
	if err := row.Scan(&isMerged); err != nil {
		return false, err
	}

	return isMerged, nil
}

// IsReviewerAssigned checks if a reviewer is assigned to a pull request.
func (r *PRRepository) IsReviewerAssigned(ctx context.Context, pullRequestId domain.PRID, reviewId domain.UserID) (bool, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)", pullRequestId, reviewId)

	var isAssigned bool
	if err := row.Scan(&isAssigned); err != nil {
		return false, err
	}

	return isAssigned, nil
}

// UnassignReviewer removes a reviewer from a pull request.
func (r *PRRepository) UnassignReviewer(ctx context.Context, pullRequestId domain.PRID, reviewerId domain.UserID) error {
	executor := r.exec.DefaultTxOrDB(ctx)
	_, err := executor.ExecContext(ctx, "DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2", pullRequestId, reviewerId)

	return err
}

// GetPrByPrID retrieves a pull request by its ID from the database.
func (r *PRRepository) GetPrByPrID(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, error) {
	executor := r.exec.DefaultTxOrDB(ctx)
	row := executor.QueryRowContext(ctx, "SELECT pull_request_id, author_id, title, created_at, is_merged, merged_at FROM pull_requests WHERE pull_request_id = $1", pullRequestId)

	var pr domain.PullRequest
	if err := row.Scan(&pr.ID, &pr.AuthorID, &pr.Title, &pr.CreatedAt, &pr.IsMerged, &pr.MergedAt); err != nil {
		return nil, err
	}

	return &pr, nil
}

// GetPrByUserID retrieves all pull requests assigned to a user from the database.
func (r *PRRepository) GetPrByUserID(ctx context.Context, userId domain.UserID) ([]domain.PullRequest, error) {
	executor := r.exec.DefaultTxOrDB(ctx)

	rows, err := executor.QueryContext(ctx,
		`SELECT pr.pull_request_id, pr.author_id, pr.title, pr.created_at, pr.is_merged, pr.merged_at
		 FROM pull_requests pr
		 JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		 WHERE rev.user_id = $1`, userId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var prs []domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.AuthorID, &pr.Title, &pr.CreatedAt, &pr.IsMerged, &pr.MergedAt); err != nil {
			return nil, err
		}

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

// GetReviewers retrieves all reviewer IDs for a pull request from the database.
func (r *PRRepository) GetReviewers(ctx context.Context, pullRequestId domain.PRID) ([]domain.UserID, error) {
	executor := r.exec.DefaultTxOrDB(ctx)

	rows, err := executor.QueryContext(ctx,
		`SELECT user_id
		 FROM pr_reviewers
		 WHERE pull_request_id = $1`, pullRequestId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var reviewerIDs []domain.UserID

	for rows.Next() {
		var reviewerID domain.UserID
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}

		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviewerIDs, nil
}

// GetAssignmentStats retrieves statistics about reviewer assignments from the database.
func (r *PRRepository) GetAssignmentStats(ctx context.Context) (*domain.Stats, error) {
	executor := r.exec.DefaultTxOrDB(ctx)

	// Get assignments per user
	userRows, err := executor.QueryContext(ctx, `
		SELECT user_id, COUNT(*) as assignments
		FROM pr_reviewers
		GROUP BY user_id
		ORDER BY assignments DESC, user_id ASC
	`)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := userRows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var byUser []domain.UserAssignmentStat

	for userRows.Next() {
		var stat domain.UserAssignmentStat
		if err := userRows.Scan(&stat.UserID, &stat.Assignments); err != nil {
			return nil, err
		}

		byUser = append(byUser, stat)
	}

	if err := userRows.Err(); err != nil {
		return nil, err
	}

	// Get reviewers per PR
	prRows, err := executor.QueryContext(ctx, `
		SELECT pull_request_id, COUNT(*) as reviewers
		FROM pr_reviewers
		GROUP BY pull_request_id
		ORDER BY reviewers DESC, pull_request_id ASC
	`)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := prRows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var byPR []domain.PRReviewerStat

	for prRows.Next() {
		var stat domain.PRReviewerStat
		if err := prRows.Scan(&stat.PullRequestID, &stat.Reviewers); err != nil {
			return nil, err
		}

		byPR = append(byPR, stat)
	}

	if err := prRows.Err(); err != nil {
		return nil, err
	}

	return &domain.Stats{
		ByUser: byUser,
		ByPR:   byPR,
	}, nil
}
