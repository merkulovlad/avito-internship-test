package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// Initialize repositories and services
	prRepository := pr.NewPRRepository(db)
	prService := pr.NewPRService(db)
	teamService := team.NewTeamService(db)
	userService := user.NewUserService(db, prRepository)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Errorf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Register routes
	handlers.RegisterRoutes(app, userService, teamService, prService)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Starting server on %s", addr)

	go func() {
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down...")
	if err := app.Shutdown(); err != nil {
		logger.Errorf("Error during shutdown: %v", err)
	}
}
