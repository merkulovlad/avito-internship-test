// Package main implements the entry point for the application.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"github.com/merkulovlad/avito-internship-test/internal/config"
	"github.com/merkulovlad/avito-internship-test/internal/databases"
	"github.com/merkulovlad/avito-internship-test/internal/handlers"
	"github.com/merkulovlad/avito-internship-test/internal/logger"

	"github.com/merkulovlad/avito-internship-test/internal/pr"
	"github.com/merkulovlad/avito-internship-test/internal/team"
	"github.com/merkulovlad/avito-internship-test/internal/user"
)

// createFiberApp initializes and configures the Fiber application.
func createFiberApp(log logger.InterfaceLogger) *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Errorf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    4 * 1024 * 1024,
	})
}

// setupServices initializes the services with their respective repositories.

func setupServices(db *sql.DB, log logger.InterfaceLogger) (
	*user.UserService,
	*team.TeamService,
	*pr.PRService,
) {
	prRepository := pr.NewPRRepository(db)
	prService := pr.NewPRService(db, log)
	teamService := team.NewTeamService(db, log)
	userService := user.NewUserService(db, prRepository, log)

	return userService, teamService, prService
}

// runServer starts the Fiber server and handles graceful shutdown.
func runServer(app *fiber.App, addr string, log logger.InterfaceLogger) {
	log.Infof("Starting server on %s", addr)

	go func() {
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down...")

	if err := app.Shutdown(); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}
}

// entry point of the application.
func main() {
	cfg := config.MustLoad()
	options := &logger.Options{
		Level:     cfg.Log.Level,
		ToConsole: cfg.Log.ToConsole,
		Filename:  cfg.Log.Filename,
	}

	logger, err := logger.NewLogger(options)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("failed to sync logger: %v\n", err)
		}
	}()

	// Initialize database
	db, err := databases.NewDB(&cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("Failed to close database: %v", err)
		}
	}()

	logger.Info("Database connected successfully")
	// Run migrations
	err = databases.RunMigrations(db)
	if err != nil {
		logger.Fatalf("failed to run migrations: %v", err)
	}

	// Initialize repositories and services
	userService, teamService, prService := setupServices(db, logger)

	// Initialize Fiber app
	app := createFiberApp(logger)

	// Register routes
	handlers.RegisterRoutes(app, userService, teamService, prService, logger)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	runServer(app, addr, logger)
}
