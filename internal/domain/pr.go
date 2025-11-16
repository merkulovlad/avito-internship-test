package domain

import (
	"context"
	"time"
)

// PRID represents a unique pull request identifier.
type PRID string

// PullRequestRepositoryInterface defines methods for pull request data persistence.
type PullRequestRepositoryInterface interface {
	// CreatePr creates a new pull request.
	CreatePr(ctx context.Context, pullRequestId PRID, authorId UserID, title string) error
	// AssignReviewer assigns a reviewer to a pull request.
	AssignReviewer(ctx context.Context, pullRequestId PRID, reviewer UserID) error
	// MergePr marks a pull request as merged.
	MergePr(ctx context.Context, pullRequestId PRID) (*PullRequest, error)

	// Exists checks if a pull request with the given ID exists.
	Exists(ctx context.Context, pullRequestId PRID) (bool, error)
	// CheckIsMerged checks if a pull request is merged.
	CheckIsMerged(ctx context.Context, pullRequestId PRID) (bool, error)

	// IsReviewerAssigned checks if a reviewer is assigned to a pull request.
	IsReviewerAssigned(ctx context.Context, pullRequestId PRID, reviewerId UserID) (bool, error)
	// UnassignReviewer removes a reviewer from a pull request.
	UnassignReviewer(ctx context.Context, pullRequestId PRID, reviewerId UserID) error

	// GetPrByPrID retrieves a pull request by its ID.
	GetPrByPrID(ctx context.Context, pullRequestId PRID) (*PullRequest, error)
	// GetPrByUserID retrieves all pull requests assigned to a user.
	GetPrByUserID(ctx context.Context, userId UserID) ([]PullRequest, error)
	// GetReviewers retrieves all reviewer IDs for a pull request.
	GetReviewers(ctx context.Context, pullRequestId PRID) ([]UserID, error)

	// GetAssignmentStats retrieves statistics about reviewer assignments.
	GetAssignmentStats(ctx context.Context) (*Stats, error)
}

// PullRequest represents a pull request entity.
type PullRequest struct {
	// ID is the unique identifier for the pull request.
	ID PRID
	// AuthorID is the user who created the pull request.
	AuthorID UserID
	// Title is the title of the pull request.
	Title string
	// CreatedAt is when the pull request was created.
	CreatedAt time.Time
	// IsMerged indicates whether the pull request has been merged.
	IsMerged bool
	// AssignedReviewers is the list of users assigned as reviewers.
	AssignedReviewers []UserID
	// MergedAt is when the pull request was merged, nil if not merged.
	MergedAt *time.Time
}

// ReassignResult contains the result of a reviewer reassignment.
type ReassignResult struct {
	// PR is the updated pull request.
	PR *PullRequest
	// ReplacedBy is the ID of the new reviewer.
	ReplacedBy UserID
}

// UserAssignmentStat represents statistics for a single user's assignments.
type UserAssignmentStat struct {
	// UserID is the user identifier.
	UserID UserID
	// Assignments is the number of pull requests assigned to this user.
	Assignments int
}

// PRReviewerStat represents statistics for a single pull request's reviewers.
type PRReviewerStat struct {
	// PullRequestID is the pull request identifier.
	PullRequestID PRID
	// Reviewers is the number of reviewers assigned to this pull request.
	Reviewers int
}

// Stats contains assignment statistics.
type Stats struct {
	// ByUser contains per-user assignment statistics.
	ByUser []UserAssignmentStat
	// ByPR contains per-pull-request reviewer statistics.
	ByPR []PRReviewerStat
}
