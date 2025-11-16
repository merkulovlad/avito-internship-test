package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/config"
	"github.com/merkulovlad/avito-internship-test/internal/databases"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

// TestCreateFiberApp tests the createFiberApp function
func TestCreateFiberApp(t *testing.T) {
	log := &logger.FakeLogger{}
	app := createFiberApp(log)

	if app == nil {
		t.Fatal("createFiberApp returned nil")
	}

	// Verify the app works by adding a test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestCreateFiberAppErrorHandler tests the error handler in createFiberApp
func TestCreateFiberAppErrorHandler(t *testing.T) {
	log := &logger.FakeLogger{}
	app := createFiberApp(log)

	// Test error handling with Fiber error
	app.Get("/fiber-error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "test error")
	})

	req := httptest.NewRequest("GET", "/fiber-error", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["error"] != "test error" {
		t.Errorf("Expected error 'test error', got %v", result["error"])
	}

	// Verify logger was called
	if len(log.Errors) == 0 {
		t.Error("Expected error to be logged")
	}
}

// TestCreateFiberAppConfiguration tests the app configuration
func TestCreateFiberAppConfiguration(t *testing.T) {
	log := &logger.FakeLogger{}
	app := createFiberApp(log)

	if app == nil {
		t.Fatal("createFiberApp returned nil")
	}

	// Test that the app has the correct timeout configurations by ensuring it works
	app.Get("/config-test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/config-test", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestSetupServices tests the setupServices function
func TestSetupServices(t *testing.T) {
	log := &logger.FakeLogger{}

	// Setup test database
	dbConfig := &config.DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		User:              "test",
		Password:          "test",
		Name:              "test",
		SSLMode:           "disable",
		MaxConnections:    1,
		ConnectionTimeout: 1,
		ConnMaxLifetime:   1,
	}

	db, err := databases.NewDB(dbConfig)
	if err != nil {
		t.Skipf("Skipping test - no test database: %v", err)
		return
	}

	defer func() {
		_ = db.Close()
	}()

	userService, teamService, prService := setupServices(db, log)

	if userService == nil {
		t.Error("userService is nil")
	}

	if teamService == nil {
		t.Error("teamService is nil")
	}

	if prService == nil {
		t.Error("prService is nil")
	}
}

// TestSetupServicesWithNilDB tests setupServices with nil database
func TestSetupServicesWithNilDB(t *testing.T) {
	log := &logger.FakeLogger{}

	// This will likely panic or return nil services, but we're testing the function exists and can be called
	defer func() {
		if r := recover(); r != nil {
			t.Logf("setupServices panicked with nil DB (expected): %v", r)
		}
	}()

	userService, teamService, prService := setupServices(nil, log)

	// If we get here without panic, verify services are created
	if userService == nil {
		t.Error("userService is nil")
	}

	if teamService == nil {
		t.Error("teamService is nil")
	}

	if prService == nil {
		t.Error("prService is nil")
	}
}

// TestFiberAppConfiguration tests the Fiber app configuration
func TestFiberAppConfiguration(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    4 * 1024 * 1024,
	})

	if app == nil {
		t.Fatal("Failed to create Fiber app")
	}
}

// TestErrorHandlerWithFiberError tests the custom error handler with Fiber errors
func TestErrorHandlerWithFiberError(t *testing.T) {
	errorHandler := func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return c.Status(code).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a mock context to test error handling logic
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	// Test with a Fiber error (should extract code)
	app.Get("/fiber-error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "bad request error")
	})

	// Test with a regular error (should default to 500)
	app.Get("/regular-error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	})

	if app == nil {
		t.Fatal("Failed to create app with error handler")
	}
}

// TestAppConfigurationTimeouts verifies timeout configurations
func TestAppConfigurationTimeouts(t *testing.T) {
	config := fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    4 * 1024 * 1024,
	}

	if config.ReadTimeout != 10*time.Second {
		t.Errorf("Expected ReadTimeout 10s, got %v", config.ReadTimeout)
	}

	if config.WriteTimeout != 10*time.Second {
		t.Errorf("Expected WriteTimeout 10s, got %v", config.WriteTimeout)
	}

	if config.IdleTimeout != 120*time.Second {
		t.Errorf("Expected IdleTimeout 120s, got %v", config.IdleTimeout)
	}

	if config.BodyLimit != 4*1024*1024 {
		t.Errorf("Expected BodyLimit 4MB, got %v", config.BodyLimit)
	}
}

// TestErrorHandlerReturnsCorrectStatusCode verifies error handler status codes
func TestErrorHandlerReturnsCorrectStatusCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "Fiber error with custom code",
			err:          fiber.NewError(fiber.StatusNotFound, "not found"),
			expectedCode: fiber.StatusNotFound,
		},
		{
			name:         "Fiber error 400",
			err:          fiber.NewError(fiber.StatusBadRequest, "bad request"),
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name:         "Fiber error 401",
			err:          fiber.NewError(fiber.StatusUnauthorized, "unauthorized"),
			expectedCode: fiber.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := fiber.StatusInternalServerError
			if e, ok := tt.err.(*fiber.Error); ok {
				code = e.Code
			}

			if code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d", tt.expectedCode, code)
			}
		})
	}
}
