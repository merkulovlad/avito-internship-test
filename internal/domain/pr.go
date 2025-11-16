package domain

import (
	"context"
	"time"
)

type PRID string

type PullRequestRepositoryInterface interface {
	CreatePr(ctx context.Context, pullRequestId PRID, authorId UserID, title string) error
	AssignReviewer(ctx context.Context, pullRequestId PRID, reviewer UserID) error
	MergePr(ctx context.Context, pullRequestId PRID) (*PullRequest, error)

	Exists(ctx context.Context, pullRequestId PRID) (bool, error)
	CheckIsMerged(ctx context.Context, pullRequestId PRID) (bool, error)

	IsReviewerAssigned(ctx context.Context, pullRequestId PRID, reviewerId UserID) (bool, error)
	UnassignReviewer(ctx context.Context, pullRequestId PRID, reviewerId UserID) error

	GetPrByPrID(ctx context.Context, pullRequestId PRID) (*PullRequest, error)
	GetPrByUserID(ctx context.Context, userId UserID) ([]PullRequest, error)
	GetReviewers(ctx context.Context, pullRequestId PRID) ([]UserID, error)
}

type PullRequest struct {
	ID                PRID
	AuthorID          UserID
	Title             string
	CreatedAt         time.Time
	IsMerged          bool
	AssignedReviewers []UserID
	MergedAt          *time.Time
}

type ReassignResult struct {
	PR         *PullRequest
	ReplacedBy UserID
}
