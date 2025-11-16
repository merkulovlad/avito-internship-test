package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

// RegisterRoutes registers all HTTP routes for the application
func RegisterRoutes(
	app *fiber.App,
	userService domain.UserServiceInterface,
	teamService domain.TeamServiceInterface,
	prService domain.PRServiceInterface,
) {
	// Initialize handlers
	userHandler := NewUserHandler(userService)
	teamHandler := NewTeamHandler(teamService)
	prHandler := NewPRHandler(prService)

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
