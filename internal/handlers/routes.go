package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

// RegisterRoutes registers all HTTP routes for the application
func RegisterRoutes(
	app *fiber.App,
	userService domain.UserServiceInterface,
	teamService domain.TeamServiceInterface,
	prService domain.PRServiceInterface,
	logger logger.InterfaceLogger,
) {
	// Initialize handlers
	userHandler := NewUserHandler(userService, logger)
	teamHandler := NewTeamHandler(teamService, logger)
	prHandler := NewPRHandler(prService, logger)

	// Health check
	app.Get("/healthz", HealthCheck)

	// Team routes
	app.Post("/team/add", teamHandler.CreateTeam)
	app.Get("/team/get", teamHandler.GetTeam)

	// User routes
	app.Post("/users/setIsActive", userHandler.SetUserIsActive)
	app.Get("/users/getReview", userHandler.GetPrOfUser)

	// Pull Request routes
	app.Post("/pullRequest/create", prHandler.CreatePr)
	app.Post("/pullRequest/merge", prHandler.MergePr)
	app.Post("/pullRequest/reassign", prHandler.ReassignReviewer)
}
