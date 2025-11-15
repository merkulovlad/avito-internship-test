package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/pr"
)

type PRHandler struct {
	service pr.PRService
}

func NewPRHandler(service pr.PRService) *PRHandler {
	return &PRHandler{
		service: service,
	}
}

func (h *PRHandler) ReassignReviewer (c *fiber.Ctx) error {
	var req api.PostPullRequestReassignJSONBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	
	updatedPR, err := h.service.ReassignReviewer(c.Context(), domain.PRID(req.PullRequestId), domain.UserID(req.OldUserId))
	if errors.Is(err, domain.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, err.Error())
	} else if errors.Is(err, domain.ErrNoCandidate) {
		return writeError(c, fiber.StatusConflict, domain.ErrNoCandidate.Code, err.Error())
	} else if errors.Is(err, domain.ErrNotAssigned) {
		return writeError(c, fiber.StatusConflict, domain.ErrNotAssigned.Code, err.Error())
	} else if errors.Is(err, domain.ErrPrMerged) {
		return writeError(c, fiber.StatusConflict, domain.ErrPrMerged.Code, err.Error())
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	if updatedPR == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	res := &PostPullRequestReassignResponse{
		PR: api.PullRequestShort{
			PullRequestId: string(updatedPR.PR.ID),
			PullRequestName:         updatedPR.PR.Title,
			AuthorId:      string(updatedPR.PR.AuthorID),
			Status:      BoolToStatus(updatedPR.PR.IsMerged),
		},
		ReplacedBy: string(updatedPR.ReplacedBy),
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *PRHandler) MergePr(c *fiber.Ctx) error {
	var req api.PostPullRequestMergeJSONBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	mergedPr, err := h.service.MergePr(c.Context(), domain.PRID(req.PullRequestId))
	if errors.Is(err, domain.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, err.Error())
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	if mergedPr == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	res := PostPullRequestMergeResponse{
		PR: *PullRequestDomainToApi(mergedPr),
	}
	return c.Status(fiber.StatusOK).JSON(res)
}




	
	




	

