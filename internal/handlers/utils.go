package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

// BoolToStatus converts a boolean merge status to an API status type.
func BoolToStatus(isMerged bool) api.PullRequestShortStatus {
	if isMerged {
		return api.PullRequestShortStatusMERGED
	}

	return api.PullRequestShortStatusOPEN
}

// writeError writes an error response to the client.
func writeError(c *fiber.Ctx, status int, code api.ErrorResponseErrorCode, msg string) error {
	resp := api.ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = msg

	return c.Status(status).JSON(resp)
}

// PullRequestDomainToApi converts a domain PullRequest to an API PR.
func PullRequestDomainToApi(pr *domain.PullRequest, reviewers []domain.User) *PR {
	var mergedAt *string

	if pr.MergedAt != nil {
		str := pr.MergedAt.Format("2006-01-02T15:04:05Z07:00")
		mergedAt = &str
	}

	var createdAt *string

	if !pr.CreatedAt.IsZero() {
		str := pr.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		createdAt = &str
	}

	return &PR{
		PullRequestId:     string(pr.ID),
		PullRequestName:   pr.Title,
		AuthorId:          string(pr.AuthorID),
		Status:            BoolToStatus(pr.IsMerged),
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
		AssignedReviewers: sliceOfUsersToStr(reviewers),
	}
}

// sliceOfUsersToStr converts a slice of users to a slice of user ID strings.
func sliceOfUsersToStr(users []domain.User) []string {
	var ids []string
	for _, r := range users {
		ids = append(ids, string(r.ID))
	}

	return ids
}
