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

type fakeTeamService struct {
	teamToReturn *domain.Team
	errToReturn  error
	Members      []domain.User
	lastName     domain.TeamName
}

func (f *fakeTeamService) GetTeamByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	f.lastName = name
	if name != f.teamToReturn.Name {
		return nil, domain.ErrNotFound
	}

	return f.teamToReturn, f.errToReturn
}

func (f *fakeTeamService) CreateTeam(ctx context.Context, t domain.TeamName, members []domain.User) (*domain.Team, error) {
	if f.errToReturn != nil {
		return nil, f.errToReturn
	}

	return &domain.Team{Name: t}, nil
}

func (f *fakeTeamService) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	return f.Members, nil
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakeTeamService{
		teamToReturn: &domain.Team{Name: "backend"},
		Members: []domain.User{
			{ID: "user1", Username: "alice", TeamName: "backend", IsActive: true},

			{ID: "user2", Username: "bob", TeamName: "backend", IsActive: false},
		},
	}
	h := NewTeamHandler(svc, log)

	app.Get("/team/get", h.GetTeam)

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", nil)

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

	if got["team_name"] != "backend" {
		t.Fatalf("expected team_name=backend, got %v", got["team_name"])
	}

	if svc.lastName != "backend" {
		t.Fatalf("handler passed wrong team name to service: %v", svc.lastName)
	}

	if svc.Members == nil {
		t.Fatalf("handler did not call GetTeamMembers")
	}
}

func TestTeamHandler_GetTeam_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{
		teamToReturn: &domain.Team{Name: "frontend"},
	}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Get("/team/get", h.GetTeam)

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestTeamHandler_GetTeam_MissingQuery(t *testing.T) {
	app := fiber.New()

	svc := &fakeTeamService{}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Get("/team/get", h.GetTeam)

	req := httptest.NewRequest("GET", "/team/get", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateTeamHandlerSuccess(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{
		teamToReturn: &domain.Team{Name: "backend"},
		Members: []domain.User{
			{ID: "user1", Username: "alice", TeamName: "backend", IsActive: true},
			{ID: "user2", Username: "bob", TeamName: "backend", IsActive: false},
		},
	}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Post("/team/add", h.CreateTeam)

	reqBody := `{
		"team_name": "backend",
		"members": [
			{"user_id": "user1", "username": "alice", "is_active": true},
			{"user_id": "user2", "username": "bob", "is_active": false}
		]
	}`

	req := httptest.NewRequest("POST", "/team/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	team, ok := got["team"].(map[string]any)
	if !ok {
		t.Fatalf("expected team object in response")
	}

	if team["team_name"] != "backend" {
		t.Fatalf("expected team_name=backend, got %v", team["team_name"])
	}
}

func TestCreateTeamHandler_MissingTeamName(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Post("/team/add", h.CreateTeam)

	reqBody := `{
		"team_name": "",
		"members": [
			{"user_id": "user1", "username": "alice", "is_active": true}
		]
	}`

	req := httptest.NewRequest("POST", "/team/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateTeamHandler_EmptyMembers(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Post("/team/add", h.CreateTeam)

	reqBody := `{
		"team_name": "backend",
		"members": []
	}`

	req := httptest.NewRequest("POST", "/team/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateTeamHandler_MissingUserID(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Post("/team/add", h.CreateTeam)

	reqBody := `{
		"team_name": "backend",
		"members": [
			{"user_id": "", "username": "alice", "is_active": true}
		]
	}`

	req := httptest.NewRequest("POST", "/team/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateTeamHandler_MissingUsername(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Post("/team/add", h.CreateTeam)

	reqBody := `{
		"team_name": "backend",
		"members": [
			{"user_id": "user1", "username": "", "is_active": true}
		]
	}`

	req := httptest.NewRequest("POST", "/team/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestTeamAlrExists(t *testing.T) {
	app := fiber.New()
	svc := &fakeTeamService{
		teamToReturn: &domain.Team{Name: "backend"},
		errToReturn:  domain.ErrTeamAlreadyExists,
	}
	log := &logger.FakeLogger{}
	h := NewTeamHandler(svc, log)
	app.Post("/team/add", h.CreateTeam)

	reqBody := `{
		"team_name": "backend",
		"members": [
			{"user_id": "user1", "username": "alice", "is_active": true}
		]
	}`

	req := httptest.NewRequest("POST", "/team/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}
