package domain

import "time"

type PRID string

type PullRequest struct {
	ID                PRID
	AuthorID          UserID
	Title             string
	CreatedAt         time.Time
	IsMerged          bool
	AssignedReviewers []UserID
	MergedAt		 *time.Time
}

type ReassignResult struct {
	PR         *PullRequest
	ReplacedBy UserID
}
