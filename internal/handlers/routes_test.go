package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	app.Get("/healthz", HealthCheck)

	req := httptest.NewRequest("GET", "/healthz", nil)

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

	if got["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", got["status"])
	}
}

func TestRegisterRoutes_TeamRoutes(t *testing.T) {
	app := fiber.New()
	userSvc := &fakeUserService{}
	teamSvc := &fakeTeamService{
		teamToReturn: &domain.Team{Name: "backend"},
		Members: []domain.User{
			{ID: "user1", Username: "alice", TeamName: "backend", IsActive: true},
		},
	}
	prSvc := &fakePRService{}
	log := &logger.FakeLogger{}

	RegisterRoutes(app, userSvc, teamSvc, prSvc, log)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST /team/add - success",
			method:         "POST",
			path:           "/team/add",
			body:           `{"team_name":"backend","members":[{"user_id":"user1","username":"alice","is_active":true}]}`,
			expectedStatus: fiber.StatusCreated,
		},
		{
			name:           "GET /team/get - success",
			method:         "GET",
			path:           "/team/get?team_name=backend",
			body:           "",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRegisterRoutes_UserRoutes(t *testing.T) {
	app := fiber.New()
	userSvc := &fakeUserService{
		userToReturn: &domain.User{
			ID:       "user1",
			Username: "alice",
			TeamName: "backend",
			IsActive: true,
		},
		prsToReturn: []domain.PullRequest{
			{ID: "pr1", Title: "Add feature", AuthorID: "user1", IsMerged: false},
		},
	}
	teamSvc := &fakeTeamService{}
	prSvc := &fakePRService{}
	log := &logger.FakeLogger{}

	RegisterRoutes(app, userSvc, teamSvc, prSvc, log)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST /users/setIsActive - success",
			method:         "POST",
			path:           "/users/setIsActive",
			body:           `{"user_id":"user1","is_active":true}`,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "GET /users/getReview - success",
			method:         "GET",
			path:           "/users/getReview?user_id=user1",
			body:           "",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRegisterRoutes_PRRoutes(t *testing.T) {
	app := fiber.New()
	userSvc := &fakeUserService{}
	teamSvc := &fakeTeamService{}
	prSvc := &fakePRService{
		prToReturn: &domain.PullRequest{
			ID:                "pr1",
			AuthorID:          "user1",
			Title:             "Add feature",
			IsMerged:          false,
			AssignedReviewers: []domain.UserID{"user2"},
		},
		reviewersToReturn: []domain.User{
			{ID: "user2", Username: "bob", TeamName: "backend", IsActive: true},
		},
		reassignToReturn: &domain.ReassignResult{
			PR: &domain.PullRequest{
				ID:                "pr1",
				AuthorID:          "user1",
				Title:             "Add feature",
				IsMerged:          false,
				AssignedReviewers: []domain.UserID{"user3"},
			},
			ReplacedBy: "user3",
		},
	}
	log := &logger.FakeLogger{}

	RegisterRoutes(app, userSvc, teamSvc, prSvc, log)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST /pullRequest/create - success",
			method:         "POST",
			path:           "/pullRequest/create",
			body:           `{"pull_request_id":"pr1","pull_request_name":"Add feature","author_id":"user1"}`,
			expectedStatus: fiber.StatusCreated,
		},
		{
			name:           "POST /pullRequest/merge - success",
			method:         "POST",
			path:           "/pullRequest/merge",
			body:           `{"pull_request_id":"pr1"}`,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "POST /pullRequest/reassign - success",
			method:         "POST",
			path:           "/pullRequest/reassign",
			body:           `{"pull_request_id":"pr1","old_user_id":"user2"}`,
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRegisterRoutes_HealthCheck(t *testing.T) {
	app := fiber.New()
	userSvc := &fakeUserService{}
	teamSvc := &fakeTeamService{}
	prSvc := &fakePRService{}
	log := &logger.FakeLogger{}

	RegisterRoutes(app, userSvc, teamSvc, prSvc, log)

	req := httptest.NewRequest("GET", "/healthz", nil)

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

	if got["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", got["status"])
	}
}

func TestRegisterRoutes_404NotFound(t *testing.T) {
	app := fiber.New()
	userSvc := &fakeUserService{}
	teamSvc := &fakeTeamService{}
	prSvc := &fakePRService{}
	log := &logger.FakeLogger{}

	RegisterRoutes(app, userSvc, teamSvc, prSvc, log)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "GET /nonexistent",
			method: "GET",
			path:   "/nonexistent",
		},
		{
			name:   "POST /invalid/route",
			method: "POST",
			path:   "/invalid/route",
		},
		{
			name:   "GET /team/invalid",
			method: "GET",
			path:   "/team/invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}

			if resp.StatusCode != fiber.StatusNotFound {
				t.Fatalf("expected status 404, got %d", resp.StatusCode)
			}
		})
	}
}

func TestRegisterRoutes_WrongMethod(t *testing.T) {
	app := fiber.New()
	userSvc := &fakeUserService{}
	teamSvc := &fakeTeamService{}
	prSvc := &fakePRService{}
	log := &logger.FakeLogger{}

	RegisterRoutes(app, userSvc, teamSvc, prSvc, log)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST /team/get - wrong method",
			method:         "POST",
			path:           "/team/get",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "GET /team/add - wrong method",
			method:         "GET",
			path:           "/team/add",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "GET /pullRequest/create - wrong method",
			method:         "GET",
			path:           "/pullRequest/create",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
