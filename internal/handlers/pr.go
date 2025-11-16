package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

type PRHandler struct {
	service domain.PRServiceInterface
}

func NewPRHandler(service domain.PRServiceInterface) *PRHandler {
	return &PRHandler{
		service: service,
	}
}

func (h *PRHandler) ReassignReviewer(c *fiber.Ctx) error {
	var req api.PostPullRequestReassignJSONBody
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	updatedPR, err := h.service.ReassignReviewer(c.Context(), domain.PRID(req.PullRequestId), domain.UserID(req.OldUserId))
	if errors.Is(err, domain.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "resource not found")
	} else if errors.Is(err, domain.ErrNoCandidate) {
		return writeError(c, fiber.StatusConflict, domain.ErrNoCandidate.Code, "no active replacement candidate in team")
	} else if errors.Is(err, domain.ErrNotAssigned) {
		return writeError(c, fiber.StatusConflict, domain.ErrNotAssigned.Code, "reviewer is not assigned to this PR")
	} else if errors.Is(err, domain.ErrPrMerged) {
		return writeError(c, fiber.StatusConflict, domain.ErrPrMerged.Code, "cannot reassign on merged PR")
	} else if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	if updatedPR == nil {
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

func (h *PRHandler) MergePr(c *fiber.Ctx) error {
	var req api.PostPullRequestMergeJSONBody
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	mergedPr, err := h.service.MergePr(c.Context(), domain.PRID(req.PullRequestId))
	if errors.Is(err, domain.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "resource not found")
	} else if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	if mergedPr == nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}
	reviewers, err := h.service.GetReviewers(c.Context(), domain.PRID(req.PullRequestId))
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	res := PostPullRequestMergeResponse{
		PR: *PullRequestDomainToApi(mergedPr, reviewers),
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *PRHandler) CreatePr(c *fiber.Ctx) error {
	var req api.PostPullRequestCreateJSONBody
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}
	pr, err := h.service.CreatePr(c.Context(), domain.PRID(req.PullRequestId), domain.UserID(req.AuthorId), req.PullRequestName)
	if errors.Is(err, domain.ErrPrAlreadyExists) {
		return writeError(c, fiber.StatusConflict, domain.ErrPrAlreadyExists.Code, "pull request already exists")
	} else if errors.Is(err, domain.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "resource not found")
	} else if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}
	reviewers, err := h.service.GetReviewers(c.Context(), domain.PRID(req.PullRequestId))
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}
	res := PostPullRequestCreateResponse{
		PR: *PullRequestDomainToApi(pr, reviewers),
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}
