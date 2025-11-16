package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

// PRHandler handles pull request-related HTTP requests.
type PRHandler struct {
	service domain.PRServiceInterface
	logger  logger.InterfaceLogger
}

// NewPRHandler creates a new PRHandler instance.
func NewPRHandler(service domain.PRServiceInterface, logger logger.InterfaceLogger) *PRHandler {
	return &PRHandler{
		service: service,
		logger:  logger,
	}
}

// ReassignReviewer handles the pull request reviewer reassignment endpoint.
func (h *PRHandler) ReassignReviewer(c *fiber.Ctx) error {
	var req api.PostPullRequestReassignJSONBody
	if err := c.BodyParser(&req); err != nil {
		h.logger.Errorf("Failed to parse request body: %v", err)
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.PullRequestId == "" {
		h.logger.Errorf("pull_request_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "pull_request_id is required and cannot be empty")
	}

	if req.OldUserId == "" {
		h.logger.Errorf("old_user_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "old_user_id is required and cannot be empty")
	}

	updatedPR, err := h.service.ReassignReviewer(c.Context(), domain.PRID(req.PullRequestId), domain.UserID(req.OldUserId))
	if errors.Is(err, domain.ErrNotFound) {
		h.logger.Errorf("Resource not found: %v", err)
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "resource not found")
	} else if errors.Is(err, domain.ErrNoCandidate) {
		h.logger.Errorf("No active replacement candidate in team: %v", err)
		return writeError(c, fiber.StatusConflict, domain.ErrNoCandidate.Code, "no active replacement candidate in team")
	} else if errors.Is(err, domain.ErrNotAssigned) {
		h.logger.Errorf("Reviewer is not assigned to this PR: %v", err)
		return writeError(c, fiber.StatusConflict, domain.ErrNotAssigned.Code, "reviewer is not assigned to this PR")
	} else if errors.Is(err, domain.ErrPrMerged) {
		h.logger.Errorf("Cannot reassign on merged PR: %v", err)
		return writeError(c, fiber.StatusConflict, domain.ErrPrMerged.Code, "cannot reassign on merged PR")
	} else if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	if updatedPR == nil {
		h.logger.Errorf("Updated PR is nil")
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	assigned := make([]string, len(updatedPR.PR.AssignedReviewers))
	for i, id := range updatedPR.PR.AssignedReviewers {
		assigned[i] = string(id)
	}

	var mergedAt *string

	if updatedPR.PR.MergedAt != nil {
		str := updatedPR.PR.MergedAt.Format("2006-01-02T15:04:05Z07:00")
		mergedAt = &str
	}

	var createdAt *string

	if !updatedPR.PR.CreatedAt.IsZero() {
		str := updatedPR.PR.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		createdAt = &str
	}

	res := &PostPullRequestReassignResponse{
		PR: PR{
			PullRequestId:     string(updatedPR.PR.ID),
			PullRequestName:   updatedPR.PR.Title,
			AuthorId:          string(updatedPR.PR.AuthorID),
			Status:            BoolToStatus(updatedPR.PR.IsMerged),
			AssignedReviewers: assigned,
			CreatedAt:         createdAt,
			MergedAt:          mergedAt,
		},
		ReplacedBy: string(updatedPR.ReplacedBy),
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// MergePr handles the pull request merge endpoint.
func (h *PRHandler) MergePr(c *fiber.Ctx) error {
	var req api.PostPullRequestMergeJSONBody
	if err := c.BodyParser(&req); err != nil {
		h.logger.Errorf("Failed to parse request body: %v", err)
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	// Validate pull_request_id
	if req.PullRequestId == "" {
		h.logger.Errorf("pull_request_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "pull_request_id is required and cannot be empty")
	}

	mergedPr, reviewers, err := h.service.MergePr(c.Context(), domain.PRID(req.PullRequestId))
	if errors.Is(err, domain.ErrNotFound) {
		h.logger.Errorf("Resource not found: %v", err)
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "resource not found")
	} else if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	if mergedPr == nil {
		h.logger.Errorf("Merged PR is nil")
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	res := PostPullRequestMergeResponse{
		PR: *PullRequestDomainToApi(mergedPr, reviewers),
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// CreatePr handles the pull request creation endpoint.
func (h *PRHandler) CreatePr(c *fiber.Ctx) error {
	var req api.PostPullRequestCreateJSONBody
	if err := c.BodyParser(&req); err != nil {
		h.logger.Errorf("Failed to parse request body: %v", err)
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.PullRequestId == "" {
		h.logger.Errorf("pull_request_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "pull_request_id is required and cannot be empty")
	}

	if req.PullRequestName == "" {
		h.logger.Errorf("pull_request_name is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "pull_request_name is required and cannot be empty")
	}

	if req.AuthorId == "" {
		h.logger.Errorf("author_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "author_id is required and cannot be empty")
	}

	pr, reviewers, err := h.service.CreatePr(c.Context(), domain.PRID(req.PullRequestId), domain.UserID(req.AuthorId), req.PullRequestName)
	if errors.Is(err, domain.ErrPrAlreadyExists) {
		h.logger.Errorf("Pull request already exists: %v", err)
		return writeError(c, fiber.StatusConflict, domain.ErrPrAlreadyExists.Code, "pull request already exists")
	} else if errors.Is(err, domain.ErrNotFound) {
		h.logger.Errorf("Resource not found: %v", err)
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "resource not found")
	} else if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	res := PostPullRequestCreateResponse{
		PR: *PullRequestDomainToApi(pr, reviewers),
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

// GetStats handles the assignment statistics endpoint.
func (h *PRHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.service.GetAssignmentStats(c.Context())
	if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	// Convert domain stats to API response
	byUser := make([]UserAssignmentStat, len(stats.ByUser))
	for i, stat := range stats.ByUser {
		byUser[i] = UserAssignmentStat{
			UserID:      string(stat.UserID),
			Assignments: stat.Assignments,
		}
	}

	byPR := make([]PRReviewerStat, len(stats.ByPR))
	for i, stat := range stats.ByPR {
		byPR[i] = PRReviewerStat{
			PullRequestID: string(stat.PullRequestID),
			Reviewers:     stat.Reviewers,
		}
	}

	res := GetStatsAssignmentsResponse{
		ByUser: byUser,
		ByPR:   byPR,
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
