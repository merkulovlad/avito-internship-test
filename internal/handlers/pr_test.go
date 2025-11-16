package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

type fakePRService struct {
	prToReturn        *domain.PullRequest
	reviewersToReturn []domain.User
	reassignToReturn  *domain.ReassignResult
	errToReturn       error
	lastPRID          domain.PRID
	lastAuthorID      domain.UserID
	lastTitle         string
	lastOldReviewerID domain.UserID
}

func (f *fakePRService) CreatePr(ctx context.Context, pullRequestId domain.PRID, authorID domain.UserID, title string) (*domain.PullRequest, []domain.User, error) {
	f.lastPRID = pullRequestId
	f.lastAuthorID = authorID
	f.lastTitle = title

	if f.errToReturn != nil {
		return nil, nil, f.errToReturn
	}

	return f.prToReturn, f.reviewersToReturn, nil
}

func (f *fakePRService) MergePr(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, []domain.User, error) {
	f.lastPRID = pullRequestId
	if f.errToReturn != nil {
		return nil, nil, f.errToReturn
	}

	return f.prToReturn, f.reviewersToReturn, nil
}

func (f *fakePRService) ReassignReviewer(ctx context.Context, pullRequestId domain.PRID, oldReviewerId domain.UserID) (*domain.ReassignResult, error) {
	f.lastPRID = pullRequestId
	f.lastOldReviewerID = oldReviewerId

	if f.errToReturn != nil {
		return nil, f.errToReturn
	}

	return f.reassignToReturn, nil
}

func (f *fakePRService) GetReviewers(ctx context.Context, pullRequestId domain.PRID) ([]domain.User, error) {
	f.lastPRID = pullRequestId
	if f.errToReturn != nil {
		return nil, f.errToReturn
	}

	return f.reviewersToReturn, nil
}

func TestPRHandler_CreatePr_Success(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}

	createdAt := time.Now()
	svc := &fakePRService{
		prToReturn: &domain.PullRequest{
			ID:                "pr1",
			AuthorID:          "user1",
			Title:             "Add feature",
			CreatedAt:         createdAt,
			IsMerged:          false,
			AssignedReviewers: []domain.UserID{"user2", "user3"},
			MergedAt:          nil,
		},
		reviewersToReturn: []domain.User{
			{ID: "user2", Username: "bob", TeamName: "backend", IsActive: true},
			{ID: "user3", Username: "charlie", TeamName: "backend", IsActive: true},
		},
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{
		"pull_request_id": "pr1",
		"pull_request_name": "Add feature",
		"author_id": "user1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
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

	pr, ok := got["pr"].(map[string]any)
	if !ok {
		t.Fatalf("expected pr object in response")
	}

	if pr["pull_request_id"] != "pr1" {
		t.Fatalf("expected pull_request_id=pr1, got %v", pr["pull_request_id"])
	}

	if pr["pull_request_name"] != "Add feature" {
		t.Fatalf("expected pull_request_name='Add feature', got %v", pr["pull_request_name"])
	}

	if pr["author_id"] != "user1" {
		t.Fatalf("expected author_id=user1, got %v", pr["author_id"])
	}

	if svc.lastPRID != "pr1" {
		t.Fatalf("handler passed wrong pr_id to service: %v", svc.lastPRID)
	}

	if svc.lastAuthorID != "user1" {
		t.Fatalf("handler passed wrong author_id to service: %v", svc.lastAuthorID)
	}

	if svc.lastTitle != "Add feature" {
		t.Fatalf("handler passed wrong title to service: %v", svc.lastTitle)
	}
}

func TestPRHandler_CreatePr_MissingPullRequestId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{
		"pull_request_id": "",
		"pull_request_name": "Add feature",
		"author_id": "user1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_CreatePr_MissingPullRequestName(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{
		"pull_request_id": "pr1",
		"pull_request_name": "",
		"author_id": "user1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_CreatePr_MissingAuthorId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{
		"pull_request_id": "pr1",
		"pull_request_name": "Add feature",
		"author_id": ""
	}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_CreatePr_AlreadyExists(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrPrAlreadyExists,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{
		"pull_request_id": "pr1",
		"pull_request_name": "Add feature",
		"author_id": "user1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestPRHandler_CreatePr_NotFound(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrNotFound,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{
		"pull_request_id": "pr1",
		"pull_request_name": "Add feature",
		"author_id": "user1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestPRHandler_CreatePr_InvalidJSON(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/create", h.CreatePr)

	reqBody := `{invalid json}`

	req := httptest.NewRequest("POST", "/pull-request/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_MergePr_Success(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}

	mergedAt := time.Now()
	createdAt := time.Now().Add(-24 * time.Hour)
	svc := &fakePRService{
		prToReturn: &domain.PullRequest{
			ID:                "pr1",
			AuthorID:          "user1",
			Title:             "Add feature",
			CreatedAt:         createdAt,
			IsMerged:          true,
			AssignedReviewers: []domain.UserID{"user2"},
			MergedAt:          &mergedAt,
		},
		reviewersToReturn: []domain.User{
			{ID: "user2", Username: "bob", TeamName: "backend", IsActive: true},
		},
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/merge", h.MergePr)

	reqBody := `{
		"pull_request_id": "pr1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/merge", strings.NewReader(reqBody))
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

	pr, ok := got["pr"].(map[string]any)
	if !ok {
		t.Fatalf("expected pr object in response")
	}

	if pr["pull_request_id"] != "pr1" {
		t.Fatalf("expected pull_request_id=pr1, got %v", pr["pull_request_id"])
	}

	if pr["status"] != "MERGED" {
		t.Fatalf("expected status=MERGED, got %v", pr["status"])
	}

	if svc.lastPRID != "pr1" {
		t.Fatalf("handler passed wrong pr_id to service: %v", svc.lastPRID)
	}
}

func TestPRHandler_MergePr_MissingPullRequestId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/merge", h.MergePr)

	reqBody := `{
		"pull_request_id": ""
	}`

	req := httptest.NewRequest("POST", "/pull-request/merge", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_MergePr_NotFound(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrNotFound,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/merge", h.MergePr)

	reqBody := `{
		"pull_request_id": "pr1"
	}`

	req := httptest.NewRequest("POST", "/pull-request/merge", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestPRHandler_MergePr_InvalidJSON(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/merge", h.MergePr)

	reqBody := `{invalid json}`

	req := httptest.NewRequest("POST", "/pull-request/merge", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_Success(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}

	createdAt := time.Now()
	svc := &fakePRService{
		reassignToReturn: &domain.ReassignResult{
			PR: &domain.PullRequest{
				ID:                "pr1",
				AuthorID:          "user1",
				Title:             "Add feature",
				CreatedAt:         createdAt,
				IsMerged:          false,
				AssignedReviewers: []domain.UserID{"user3"},
				MergedAt:          nil,
			},
			ReplacedBy: "user3",
		},
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "pr1",
		"old_user_id": "user2"
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
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

	pr, ok := got["pr"].(map[string]any)
	if !ok {
		t.Fatalf("expected pr object in response")
	}

	if pr["pull_request_id"] != "pr1" {
		t.Fatalf("expected pull_request_id=pr1, got %v", pr["pull_request_id"])
	}

	if got["replaced_by"] != "user3" {
		t.Fatalf("expected replaced_by=user3, got %v", got["replaced_by"])
	}

	if svc.lastPRID != "pr1" {
		t.Fatalf("handler passed wrong pr_id to service: %v", svc.lastPRID)
	}

	if svc.lastOldReviewerID != "user2" {
		t.Fatalf("handler passed wrong old_user_id to service: %v", svc.lastOldReviewerID)
	}
}

func TestPRHandler_ReassignReviewer_MissingPullRequestId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "",
		"old_user_id": "user2"
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_MissingOldUserId(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "pr1",
		"old_user_id": ""
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_NotFound(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrNotFound,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "pr1",
		"old_user_id": "user2"
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_NoCandidate(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrNoCandidate,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "pr1",
		"old_user_id": "user2"
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_NotAssigned(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrNotAssigned,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "pr1",
		"old_user_id": "user2"
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_PrMerged(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{
		errToReturn: domain.ErrPrMerged,
	}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{
		"pull_request_id": "pr1",
		"old_user_id": "user2"
	}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestPRHandler_ReassignReviewer_InvalidJSON(t *testing.T) {
	app := fiber.New()
	log := &logger.FakeLogger{}
	svc := &fakePRService{}
	h := NewPRHandler(svc, log)

	app.Post("/pull-request/reassign", h.ReassignReviewer)

	reqBody := `{invalid json}`

	req := httptest.NewRequest("POST", "/pull-request/reassign", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}
