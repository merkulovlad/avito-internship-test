package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

type TeamHandler struct {
	service domain.TeamServiceInterface
	logger  logger.InterfaceLogger
}

func NewTeamHandler(service domain.TeamServiceInterface, logger logger.InterfaceLogger) *TeamHandler {
	return &TeamHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	var req api.PostTeamAddJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		h.logger.Errorf("Failed to parse request body: %v", err)
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "Invalid request body")
	}

	h.logger.Infof("getting request to create team %s", req.TeamName)
	// Validate team name is not empty
	if req.TeamName == "" {
		h.logger.Errorf("team_name is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "team_name is required and cannot be empty")
	}

	// Validate members
	if len(req.Members) == 0 {
		h.logger.Errorf("members list cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "members list cannot be empty")
	}

	var members []domain.User

	var member domain.User

	for _, m := range req.Members {
		// Validate each member
		if m.UserId == "" {
			h.logger.Errorf("user_id is required and cannot be empty")
			return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "user_id is required and cannot be empty")
		}

		if m.Username == "" {
			h.logger.Errorf("username is required and cannot be empty")
			return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "username is required and cannot be empty")
		}

		member = domain.User{
			ID:       domain.UserID(m.UserId),
			Username: m.Username,
			IsActive: m.IsActive,
		}
		members = append(members, member)
	}

	createdTeam, err := h.service.CreateTeam(c.Context(), domain.TeamName(req.TeamName), members)
	if errors.Is(err, domain.ErrTeamAlreadyExists) {
		h.logger.Errorf("Team already exists: %v", err)
		return writeError(c, fiber.StatusBadRequest, domain.ErrTeamAlreadyExists.Code, "team_name already exists")
	} else if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	membersCurrent, err := h.service.GetTeamMembers(c.Context(), req.TeamName)
	if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	res := TeamCreationResponse{
		Team: Team{
			TeamName: createdTeam.Name,
			Members:  membersCurrent,
		},
	}
	h.logger.Infof("Created team %s with %d members", res.TeamName, len(res.Members))

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *TeamHandler) GetTeam(c *fiber.Ctx) error {
	teamName := c.Query("team_name")
	h.logger.Infof("getting request to get team %s", teamName)

	if teamName == "" {
		h.logger.Errorf("team_name is required and cannot be empty")
		return writeError(c, fiber.StatusBadRequest, ErrorBadRequest, "team_name is required")
	}

	team, err := h.service.GetTeamByName(c.Context(), domain.TeamName(teamName))
	if errors.Is(err, domain.ErrNotFound) {
		h.logger.Errorf("Team not found: %v", err)
		return writeError(c, fiber.StatusNotFound, domain.ErrNotFound.Code, "team not found")
	} else if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
		return writeError(c, fiber.StatusInternalServerError, ErrorCodeInternal, "internal server error")
	}

	members, err := h.service.GetTeamMembers(c.Context(), teamName)
	if err != nil {
		h.logger.Errorf("Internal server error: %v", err)
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
	h.logger.Infof("Retrieved team %s with %d members", res.TeamName, len(res.Members))

	return c.Status(fiber.StatusOK).JSON(res)
}
