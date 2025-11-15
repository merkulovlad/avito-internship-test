package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

type PR struct {
	PullRequestId   string                      `json:"pull_request_id"`
	PullRequestName string                      `json:"pull_request_name"`
	AuthorId        string                      `json:"author_id"`
	Status          api.PullRequestShortStatus `json:"status"`
	AssignedReviewers []string                  `json:"assigned_reviewers"`
	MergedAt        *string                     `json:"mergedAt,omitempty"`
}

type PostPullRequestMergeResponse struct {
	PR  `json:"pr"`
}

type PostPullRequestReassignResponse struct {
	PR         api.PullRequestShort `json:"pr"`
	ReplacedBy string               `json:"replaced_by"`
}

func BoolToStatus(isMerged bool) api.PullRequestShortStatus {
	if isMerged {
		return api.PullRequestShortStatusMERGED
	}
	return api.PullRequestShortStatusOPEN
}

func writeError(c *fiber.Ctx, status int, code api.ErrorResponseErrorCode, msg string) error {
	resp := api.ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = msg

	return c.Status(status).JSON(resp)
}

func PullRequestDomainToApi(pr *domain.PullRequest) *PR {
	var mergedAt *string
	if pr.MergedAt != nil {
		str := pr.MergedAt.Format("2006-01-02T15:04:05Z07:00")
		mergedAt = &str
	}
	return &PR{
		PullRequestId:   string(pr.ID),
		PullRequestName: pr.Title,
		AuthorId:        string(pr.AuthorID),
		Status:          BoolToStatus(pr.IsMerged),
		MergedAt:        mergedAt,
	}
}
 