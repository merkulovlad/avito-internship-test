package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

type UserHandler struct {
	service domain.UserServiceInterface
	logger  logger.InterfaceLogger
}

func NewUserHandler(service domain.UserServiceInterface, logger logger.InterfaceLogger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) SetUserIsActive(c *fiber.Ctx) error {
	var req api.PostUsersSetIsActiveJSONBody

	if err := c.BodyParser(&req); err != nil {
		h.logger.Errorf("Failed to parse request body: %v", err)
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	h.logger.Infof("getting request to set is_active=%v for user %s", req.IsActive, req.UserId)

	// Validate user_id is not empty
	if req.UserId == "" {
		h.logger.Errorf("user_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "user_id is required and cannot be empty")
	}

	user, err := h.service.SetIsActive(c.Context(), domain.UserID(req.UserId), req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.logger.Errorf("User not found: %v", err)
			return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "user not found")
		}

		h.logger.Errorf("Internal server error: %v", err)

		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	res := map[string]interface{}{
		"user": api.User{
			UserId:   string(user.ID),
			Username: user.Username,
			TeamName: string(user.TeamName),
			IsActive: user.IsActive,
		},
	}

	h.logger.Infof("set is_active=%v for user %s successfully", req.IsActive, req.UserId)

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *UserHandler) GetPrOfUser(c *fiber.Ctx) error {
	userId := c.Query("user_id")
	if userId == "" {
		h.logger.Errorf("user_id is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "user_id is required and cannot be empty")
	}

	prs, err := h.service.GetPrOfUser(c.Context(), domain.UserID(userId))
	if errors.Is(err, domain.ErrNotFound) {
		h.logger.Errorf("User not found: %v", err)
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "user not found")
	}

	if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	var pullRequests []api.PullRequestShort
	for _, pr := range prs {
		pullRequests = append(pullRequests, api.PullRequestShort{
			PullRequestId:   string(pr.ID),
			PullRequestName: pr.Title,
			AuthorId:        string(pr.AuthorID),
			Status:          BoolToStatus(pr.IsMerged),
		})
	}

	res := map[string]interface{}{
		"user_id":       userId,
		"pull_requests": pullRequests,
	}

	h.logger.Infof("retrieved %d pull requests for user %s", len(pullRequests), userId)

	return c.Status(fiber.StatusOK).JSON(res)
}
