// Package handlers implements HTTP request handlers for the application.
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

// RegisterRoutes registers all HTTP routes for the application.
func RegisterRoutes(
	app *fiber.App,
	userService domain.UserServiceInterface,
	teamService domain.TeamServiceInterface,
	prService domain.PRServiceInterface,
	logger logger.InterfaceLogger,
) {
	userHandler := NewUserHandler(userService, logger)
	teamHandler := NewTeamHandler(teamService, logger)
	prHandler := NewPRHandler(prService, logger)

	app.Get("/healthz", HealthCheck)

	app.Post("/team/add", teamHandler.CreateTeam)
	app.Get("/team/get", teamHandler.GetTeam)

	app.Post("/users/setIsActive", userHandler.SetUserIsActive)
	app.Get("/users/getReview", userHandler.GetPrOfUser)

	app.Post("/pullRequest/create", prHandler.CreatePr)
	app.Post("/pullRequest/merge", prHandler.MergePr)
	app.Post("/pullRequest/reassign", prHandler.ReassignReviewer)

	app.Get("/stats/assignments", prHandler.GetStats)
}
