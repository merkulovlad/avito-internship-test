package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

type fakeUserService struct {
	userToReturn *domain.User
	prsToReturn  []domain.PullRequest
	errToReturn  error
	lastUserID   domain.UserID
	lastIsActive bool
}

func (f *fakeUserService) SetIsActive(ctx context.Context, id domain.UserID, isActive bool) (*domain.User, error) {
	f.lastUserID = id
	f.lastIsActive = isActive

	if f.errToReturn != nil {
		return nil, f.errToReturn
	}

	return f.userToReturn, nil
}

func (f *fakeUserService) GetPrOfUser(ctx context.Context, userId domain.UserID) ([]domain.PullRequest, error) {
	f.lastUserID = userId
	if f.errToReturn != nil {
		return nil, f.errToReturn
	}

	return f.prsToReturn, nil
}

func TestUserHandler_SetUserIsActive_Success(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{
		userToReturn: &domain.User{
			ID:       "user1",
			Username: "alice",
			TeamName: "backend",
			IsActive: true,
		},
	}
	h := NewUserHandler(svc, log)

	app.Post("/users/set-is-active", h.SetUserIsActive)

	reqBody := `{
		"user_id": "user1",
		"is_active": true
	}`

	req := httptest.NewRequest("POST", "/users/set-is-active", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	user, ok := got["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected user object in response")
	}

	if user["user_id"] != "user1" {
		t.Fatalf("expected user_id=user1, got %v", user["user_id"])
	}

	if user["is_active"] != true {
		t.Fatalf("expected is_active=true, got %v", user["is_active"])
	}

	if svc.lastUserID != "user1" {
		t.Fatalf("handler passed wrong user_id to service: %v", svc.lastUserID)
	}

	if svc.lastIsActive != true {
		t.Fatalf("handler passed wrong is_active to service: %v", svc.lastIsActive)
	}
}

func TestUserHandler_SetUserIsActive_MissingUserId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{}
	h := NewUserHandler(svc, log)

	app.Post("/users/set-is-active", h.SetUserIsActive)

	reqBody := `{
		"user_id": "",
		"is_active": true
	}`

	req := httptest.NewRequest("POST", "/users/set-is-active", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_SetUserIsActive_UserNotFound(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{
		errToReturn: domain.ErrNotFound,
	}
	h := NewUserHandler(svc, log)

	app.Post("/users/set-is-active", h.SetUserIsActive)

	reqBody := `{
		"user_id": "nonexistent",
		"is_active": true
	}`

	req := httptest.NewRequest("POST", "/users/set-is-active", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUserHandler_SetUserIsActive_InvalidJSON(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{}
	h := NewUserHandler(svc, log)

	app.Post("/users/set-is-active", h.SetUserIsActive)

	reqBody := `{invalid json}`

	req := httptest.NewRequest("POST", "/users/set-is-active", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_GetPrOfUser_Success(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{
		prsToReturn: []domain.PullRequest{
			{
				ID:       "pr1",
				Title:    "Add feature",
				AuthorID: "user1",
				IsMerged: false,
			},
			{
				ID:       "pr2",
				Title:    "Fix bug",
				AuthorID: "user1",
				IsMerged: true,
			},
		},
	}
	h := NewUserHandler(svc, log)

	app.Get("/users/prs", h.GetPrOfUser)

	req := httptest.NewRequest("GET", "/users/prs?user_id=user1", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["user_id"] != "user1" {
		t.Fatalf("expected user_id=user1, got %v", got["user_id"])
	}

	prs, ok := got["pull_requests"].([]any)
	if !ok {
		t.Fatalf("expected pull_requests array in response")
	}

	if len(prs) != 2 {
		t.Fatalf("expected 2 pull requests, got %d", len(prs))
	}

	if svc.lastUserID != "user1" {
		t.Fatalf("handler passed wrong user_id to service: %v", svc.lastUserID)
	}
}

func TestUserHandler_GetPrOfUser_MissingUserId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{}
	h := NewUserHandler(svc, log)

	app.Get("/users/prs", h.GetPrOfUser)

	req := httptest.NewRequest("GET", "/users/prs", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_GetPrOfUser_UserNotFound(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{
		errToReturn: domain.ErrNotFound,
	}
	h := NewUserHandler(svc, log)

	app.Get("/users/prs", h.GetPrOfUser)

	req := httptest.NewRequest("GET", "/users/prs?user_id=nonexistent", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUserHandler_GetPrOfUser_EmptyResult(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeUserService{
		prsToReturn: []domain.PullRequest{},
	}
	h := NewUserHandler(svc, log)

	app.Get("/users/prs", h.GetPrOfUser)

	req := httptest.NewRequest("GET", "/users/prs?user_id=user1", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["user_id"] != "user1" {
		t.Fatalf("expected user_id=user1, got %v", got["user_id"])
	}
}
