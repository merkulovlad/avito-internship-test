package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

type TeamHandler struct {
	service domain.TeamServiceInterface
}

func NewTeamHandler(service domain.TeamServiceInterface) *TeamHandler {
	return &TeamHandler{
		service: service,
	}
}

func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	var req api.PostTeamAddJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}
	var members []domain.User
	var member domain.User

	for _, m := range req.Members {
		member = domain.User{
			ID:       domain.UserID(m.UserId),
			Username: m.Username,
			IsActive: m.IsActive,
		}
		members = append(members, member)
	}

	createdTeam, err := h.service.CreateTeam(c.Context(), domain.TeamName(req.TeamName), members)
	if errors.Is(err, domain.ErrTeamAlreadyExists) {
		return writeError(c, fiber.StatusBadRequest, domain.ErrTeamAlreadyExists.Code, "team_name already exists")
	} else if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	membersCurrent, err := h.service.GetTeamMembers(c.Context(), req.TeamName)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	res := TeamCreationResponse{
		Team: Team{
			TeamName: createdTeam.Name,
			Members:  membersCurrent,
		},
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *TeamHandler) GetTeam(c *fiber.Ctx) error {
	teamName := c.Query("team_name")
	if teamName == "" {
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "team_name is required")
	}

	team, err := h.service.GetTeamByName(c.Context(), domain.TeamName(teamName))
	if errors.Is(err, domain.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "team not found")
	} else if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	members, err := h.service.GetTeamMembers(c.Context(), teamName)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	var apiMembers []api.TeamMember
	for _, m := range members {
		apiMembers = append(apiMembers, api.TeamMember{
			UserId:   string(m.ID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	res := api.Team{
		TeamName: string(team.Name),
		Members:  apiMembers,
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
