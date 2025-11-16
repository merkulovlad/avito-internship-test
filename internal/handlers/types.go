package handlers

import (
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

const (
	ErrorCodeInternal api.ErrorResponseErrorCode = "INTERNAL_ERROR"
	ErrorBadRequest   api.ErrorResponseErrorCode = "BAD_REQUEST"
)

type TeamCreationResponse struct {
	Team `json:"team"`
}

type Team struct {
	TeamName domain.TeamName `json:"team_name"`
	Members  []domain.User   `json:"members"`
}

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

type PostPullRequestMergeResponse struct {
	PR `json:"pr"`
}

type PostPullRequestReassignResponse struct {
	PR         PR     `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}
