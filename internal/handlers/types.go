package handlers

import (
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

// Error code constants.
const (
	// ErrorCodeInternal represents an internal server error.
	ErrorCodeInternal api.ErrorResponseErrorCode = "INTERNAL_ERROR"
	// ErrorBadRequest represents a bad request error.
	ErrorBadRequest api.ErrorResponseErrorCode = "BAD_REQUEST"
)

// TeamCreationResponse represents the response for team creation.
type TeamCreationResponse struct {
	Team `json:"team"`
}

// Team represents a team with its members.
type Team struct {
	TeamName domain.TeamName `json:"team_name"`
	Members  []domain.User   `json:"members"`
}

// PRCreationResponse represents the response for pull request creation.
type PRCreationResponse struct {
	PR `json:"pr"`
}

type PostPullRequestCreateResponse struct {
	PR `json:"pr"`
}

type PR struct {
	PullRequestId     string                     `json:"pull_request_id"`
	PullRequestName   string                     `json:"pull_request_name"`
	AuthorId          string                     `json:"author_id"`
	Status            api.PullRequestShortStatus `json:"status"`
	AssignedReviewers []string                   `json:"assigned_reviewers"`
	CreatedAt         *string                    `json:"createdAt,omitempty"`
	MergedAt          *string                    `json:"mergedAt,omitempty"`
}

// PostPullRequestMergeResponse represents the response for merging a pull request.
type PostPullRequestMergeResponse struct {
	PR `json:"pr"`
}

// PostPullRequestReassignResponse represents the response for reassigning a reviewer.
type PostPullRequestReassignResponse struct {
	PR         PR     `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

// UserAssignmentStat represents statistics for a single user's assignments.
type UserAssignmentStat struct {
	UserID      string `json:"user_id"`
	Assignments int    `json:"assignments"`
}

// PRReviewerStat represents statistics for a single pull request's reviewers.
type PRReviewerStat struct {
	PullRequestID string `json:"pull_request_id"`
	Reviewers     int    `json:"reviewers"`
}

// GetStatsAssignmentsResponse represents the response for assignment statistics.
type GetStatsAssignmentsResponse struct {
	ByUser []UserAssignmentStat `json:"by_user"`
	ByPR   []PRReviewerStat     `json:"by_pr"`
}
