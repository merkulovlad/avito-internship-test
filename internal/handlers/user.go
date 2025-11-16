package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

type UserHandler struct {
	service domain.UserServiceInterface
}

func NewUserHandler(service domain.UserServiceInterface) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) SetUserIsActive(c *fiber.Ctx) error {
	var req api.PostUsersSetIsActiveJSONBody
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	user, err := h.service.SetIsActive(c.Context(), domain.UserID(req.UserId), req.IsActive)
	if err != nil {
		if err == domain.ErrNotFound {
			return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "user not found")
		}
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")

	}
	res := api.User{
		UserId:   string(user.ID),
		Username: user.Username,
		TeamName: string(user.TeamName),
		IsActive: user.IsActive,
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *UserHandler) GetPrOfUser(c *fiber.Ctx) error {
	userId := c.Params("userId")
	prs, err := h.service.GetPrOfUser(c.Context(), domain.UserID(userId))
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	var res []api.PullRequestShort
	for _, pr := range prs {
		res = append(res, api.PullRequestShort{
			PullRequestId:   string(pr.ID),
			PullRequestName: pr.Title,
			AuthorId:        string(pr.AuthorID),
			Status:          BoolToStatus(pr.IsMerged),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
