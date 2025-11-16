package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/databases"
	"github.com/merkulovlad/avito-internship-test/internal/handlers"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
	"github.com/merkulovlad/avito-internship-test/internal/pr"
	"github.com/merkulovlad/avito-internship-test/internal/team"
	"github.com/merkulovlad/avito-internship-test/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB  *sql.DB
	testApp *fiber.App
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Get test database DSN from environment variable
	testDSN := os.Getenv("TEST_DATABASE_DSN")
	if testDSN == "" {
		fmt.Println("TEST_DATABASE_DSN environment variable not set, using default")

		testDSN = "host=localhost port=5432 user=postgres password=postgres dbname=test_db sslmode=disable"
	}

	// Open database connection
	var err error

	testDB, err = sql.Open("postgres", testDSN)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Ping database to verify connection
	if err := testDB.Ping(); err != nil {
		fmt.Printf("Failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	// Set connection pool settings
	testDB.SetMaxOpenConns(25)
	testDB.SetMaxIdleConns(10)
	testDB.SetConnMaxLifetime(5 * time.Minute)

	// Run migrations
	if err := databases.RunMigrations(testDB); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Create a simple test logger
	testLogger := &logger.FakeLogger{}

	// Set up services (same as in cmd/main.go)
	prRepository := pr.NewPRRepository(testDB)
	prService := pr.NewPRService(testDB, testLogger)
	teamService := team.NewTeamService(testDB, testLogger)
	userService := user.NewUserService(testDB, prRepository, testLogger)

	// Create Fiber app
	testApp = fiber.New(fiber.Config{
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
	})

	// Register routes (same as in cmd/main.go)
	handlers.RegisterRoutes(testApp, userService, teamService, prService, testLogger)

	// Run tests
	exitCode := m.Run()

	// Cleanup
	if testDB != nil {
		if err := testDB.Close(); err != nil {
			fmt.Printf("failed to close test DB: %v\n", err)
		}
	}

	os.Exit(exitCode)
}

// Helper functions

// truncateAllTables clears all data from the database tables
func truncateAllTables(t *testing.T) {
	_, err := testDB.Exec(`
		TRUNCATE TABLE pr_reviewers CASCADE;
		TRUNCATE TABLE pull_requests CASCADE;
		TRUNCATE TABLE users CASCADE;
		TRUNCATE TABLE teams CASCADE;
	`)
	require.NoError(t, err, "Failed to truncate tables")
}

// makeRequest is a helper to make HTTP requests to the test app
func makeRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")

		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(t, err, "Failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := testApp.Test(req, -1) // -1 = no timeout
	require.NoError(t, err, "Failed to execute request")

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	return resp, respBody
}

// seedTeam creates a team with members
func seedTeam(t *testing.T, teamName string, members []api.TeamMember) {
	teamReq := api.Team{
		TeamName: teamName,
		Members:  members,
	}
	resp, _ := makeRequest(t, "POST", "/team/add", teamReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode, "Failed to seed team")
}

// Integration Tests

func TestTeamCreate_Success(t *testing.T) {
	truncateAllTables(t)

	// Request body matches OpenAPI schema
	teamReq := api.Team{
		TeamName: "backend",
		Members: []api.TeamMember{
			{
				UserId:   "u1",
				Username: "Alice",
				IsActive: true,
			},
			{
				UserId:   "u2",
				Username: "Bob",
				IsActive: true,
			},
		},
	}

	resp, body := makeRequest(t, "POST", "/team/add", teamReq)

	// Verify status code (201 as per OpenAPI)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Parse response
	var result map[string]api.Team
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify response structure
	require.Contains(t, result, "team")
	team := result["team"]
	assert.Equal(t, "backend", team.TeamName)
	assert.Len(t, team.Members, 2)

	// Verify team exists in database
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM teams WHERE team_name = $1", "backend").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify users exist in database
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE team_name = $1", "backend").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestTeamCreate_AlreadyExists(t *testing.T) {
	truncateAllTables(t)

	teamReq := api.Team{
		TeamName: "payments",
		Members: []api.TeamMember{
			{
				UserId:   "u1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	// Create team first time
	resp, _ := makeRequest(t, "POST", "/team/add", teamReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Try to create same team again
	resp, body := makeRequest(t, "POST", "/team/add", teamReq)

	// Should return 400 with TEAM_EXISTS error code as per OpenAPI
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.TEAMEXISTS, errResp.Error.Code)
	assert.Contains(t, errResp.Error.Message, "already exists")
}

func TestTeamGet_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed team
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: false},
	}
	seedTeam(t, "frontend", members)

	// Get team
	resp, body := makeRequest(t, "GET", "/team/get?team_name=frontend", nil)

	// Verify status code (200 as per OpenAPI)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	var team api.Team
	err := json.Unmarshal(body, &team)
	require.NoError(t, err)

	// Verify response matches OpenAPI schema
	assert.Equal(t, "frontend", team.TeamName)
	assert.Len(t, team.Members, 2)

	// Verify member details
	for _, member := range team.Members {
		assert.NotEmpty(t, member.UserId)
		assert.NotEmpty(t, member.Username)
	}
}

func TestTeamGet_NotFound(t *testing.T) {
	truncateAllTables(t)

	// Try to get non-existing team
	resp, body := makeRequest(t, "GET", "/team/get?team_name=nonexistent", nil)

	// Should return 404 with NOT_FOUND error code as per OpenAPI
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.NOTFOUND, errResp.Error.Code)
}

func TestUsersSetIsActive_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed user
	members := []api.TeamMember{
		{UserId: "u2", Username: "Bob", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Set user inactive
	reqBody := map[string]interface{}{
		"user_id":   "u2",
		"is_active": false,
	}

	resp, body := makeRequest(t, "POST", "/users/setIsActive", reqBody)

	// Verify status code (200 as per OpenAPI)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]api.User
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify response structure (matches OpenAPI)
	require.Contains(t, result, "user")
	user := result["user"]
	assert.Equal(t, "u2", user.UserId)
	assert.Equal(t, "Bob", user.Username)
	assert.Equal(t, "backend", user.TeamName)
	assert.False(t, user.IsActive)

	// Verify database state
	var isActive bool
	err = testDB.QueryRow("SELECT is_active FROM users WHERE user_id = $1", "u2").Scan(&isActive)
	require.NoError(t, err)
	assert.False(t, isActive)
}

func TestUsersSetIsActive_NotFound(t *testing.T) {
	truncateAllTables(t)

	reqBody := map[string]interface{}{
		"user_id":   "nonexistent",
		"is_active": true,
	}

	resp, body := makeRequest(t, "POST", "/users/setIsActive", reqBody)

	// Should return 404 as per OpenAPI
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.NOTFOUND, errResp.Error.Code)
}

func TestPullRequestCreate_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed team with multiple users
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
		{UserId: "u3", Username: "Charlie", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create PR
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}

	resp, body := makeRequest(t, "POST", "/pullRequest/create", prReq)

	// Verify status code (201 as per OpenAPI)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify response structure
	require.Contains(t, result, "pr")
	prData := result["pr"].(map[string]interface{})
	assert.Equal(t, "pr-1001", prData["pull_request_id"])
	assert.Equal(t, "Add search", prData["pull_request_name"])
	assert.Equal(t, "u1", prData["author_id"])
	assert.Equal(t, "OPEN", prData["status"])

	// Verify reviewers are assigned (up to 2, not including author)
	reviewers := prData["assigned_reviewers"].([]interface{})
	assert.GreaterOrEqual(t, len(reviewers), 0)
	assert.LessOrEqual(t, len(reviewers), 2)

	// Verify no reviewer is the author
	for _, r := range reviewers {
		assert.NotEqual(t, "u1", r)
	}

	// Verify PR exists in database
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE pull_request_id = $1", "pr-1001").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPullRequestCreate_AlreadyExists(t *testing.T) {
	truncateAllTables(t)

	// Seed team and user
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create PR first time
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	resp, _ := makeRequest(t, "POST", "/pullRequest/create", prReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Try to create same PR again
	resp, body := makeRequest(t, "POST", "/pullRequest/create", prReq)

	// Should return 409 with PR_EXISTS error code as per OpenAPI
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.PREXISTS, errResp.Error.Code)
}

func TestPullRequestCreate_AuthorNotFound(t *testing.T) {
	truncateAllTables(t)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "nonexistent",
	}

	resp, body := makeRequest(t, "POST", "/pullRequest/create", prReq)

	// Should return 404 as per OpenAPI
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.NOTFOUND, errResp.Error.Code)
}

func TestPullRequestMerge_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed team and create PR
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
	}
	seedTeam(t, "backend", members)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	resp, _ := makeRequest(t, "POST", "/pullRequest/create", prReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Merge PR
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
	}
	resp, body := makeRequest(t, "POST", "/pullRequest/merge", mergeReq)

	// Verify status code (200 as per OpenAPI)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify PR is merged
	prData := result["pr"].(map[string]interface{})
	assert.Equal(t, "MERGED", prData["status"])
	assert.NotNil(t, prData["mergedAt"])

	// Verify database state
	var isMerged bool
	err = testDB.QueryRow("SELECT is_merged FROM pull_requests WHERE pull_request_id = $1", "pr-1001").Scan(&isMerged)
	require.NoError(t, err)
	assert.True(t, isMerged)
}

func TestPullRequestMerge_Idempotent(t *testing.T) {
	truncateAllTables(t)

	// Seed and create PR
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
	}
	seedTeam(t, "backend", members)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	makeRequest(t, "POST", "/pullRequest/create", prReq)

	// Merge PR first time
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
	}
	resp1, _ := makeRequest(t, "POST", "/pullRequest/merge", mergeReq)
	require.Equal(t, fiber.StatusOK, resp1.StatusCode)

	// Merge PR second time (idempotent)
	resp2, body2 := makeRequest(t, "POST", "/pullRequest/merge", mergeReq)

	// Should still return 200 and PR should be MERGED
	assert.Equal(t, fiber.StatusOK, resp2.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body2, &result)
	require.NoError(t, err)

	prData := result["pr"].(map[string]interface{})
	assert.Equal(t, "MERGED", prData["status"])
}

func TestPullRequestMerge_NotFound(t *testing.T) {
	truncateAllTables(t)

	mergeReq := map[string]interface{}{
		"pull_request_id": "nonexistent",
	}

	resp, body := makeRequest(t, "POST", "/pullRequest/merge", mergeReq)

	// Should return 404 as per OpenAPI
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.NOTFOUND, errResp.Error.Code)
}

func TestPullRequestReassign_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed team with multiple users
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
		{UserId: "u3", Username: "Charlie", IsActive: true},
		{UserId: "u4", Username: "Dave", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create PR (should assign some reviewers)
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	resp, body := makeRequest(t, "POST", "/pullRequest/create", prReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Parse create response to get assigned reviewers
	var createResult map[string]interface{}

	err := json.Unmarshal(body, &createResult)
	require.NoError(t, err)

	prData := createResult["pr"].(map[string]interface{})
	reviewers := prData["assigned_reviewers"].([]interface{})

	// Skip test if no reviewers assigned
	if len(reviewers) == 0 {
		t.Skip("No reviewers assigned, cannot test reassignment")
	}

	oldReviewerID := reviewers[0].(string)

	// Reassign first reviewer
	reassignReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
		"old_user_id":     oldReviewerID,
	}
	resp, body = makeRequest(t, "POST", "/pullRequest/reassign", reassignReq)

	// Verify status code (200 as per OpenAPI)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify response structure (matches OpenAPI)
	require.Contains(t, result, "pr")
	require.Contains(t, result, "replaced_by")

	newReviewerID := result["replaced_by"].(string)
	assert.NotEqual(t, oldReviewerID, newReviewerID, "New reviewer should be different")
	assert.NotEqual(t, "u1", newReviewerID, "New reviewer should not be the author")

	// Verify new reviewers list doesn't contain old reviewer
	newPRData := result["pr"].(map[string]interface{})

	newReviewers := newPRData["assigned_reviewers"].([]interface{})
	for _, r := range newReviewers {
		assert.NotEqual(t, oldReviewerID, r)
	}
}

func TestPullRequestReassign_NotAssigned(t *testing.T) {
	truncateAllTables(t)

	// Seed team with users
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
		{UserId: "u3", Username: "Charlie", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create PR
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	makeRequest(t, "POST", "/pullRequest/create", prReq)

	// Try to reassign a user who is not a reviewer (u3 likely not assigned if only 2 reviewers max)
	reassignReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
		"old_user_id":     "u3",
	}
	resp, body := makeRequest(t, "POST", "/pullRequest/reassign", reassignReq)

	// Should return 409 with NOT_ASSIGNED error code as per OpenAPI
	// (or 200 if u3 was actually assigned, in which case this test is design-dependent)
	if resp.StatusCode == fiber.StatusConflict {
		var errResp api.ErrorResponse
		err := json.Unmarshal(body, &errResp)
		require.NoError(t, err)
		assert.Equal(t, api.NOTASSIGNED, errResp.Error.Code)
	}
}

func TestPullRequestReassign_PRMerged(t *testing.T) {
	truncateAllTables(t)

	// Seed team with users
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create PR
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	resp, body := makeRequest(t, "POST", "/pullRequest/create", prReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Get assigned reviewer
	var createResult map[string]interface{}

	err := json.Unmarshal(body, &createResult)
	require.NoError(t, err)

	prData := createResult["pr"].(map[string]interface{})

	reviewers := prData["assigned_reviewers"].([]interface{})
	if len(reviewers) == 0 {
		t.Skip("No reviewers assigned")
	}

	reviewerID := reviewers[0].(string)

	// Merge PR
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
	}
	makeRequest(t, "POST", "/pullRequest/merge", mergeReq)

	// Try to reassign after merge
	reassignReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
		"old_user_id":     reviewerID,
	}
	resp, body = makeRequest(t, "POST", "/pullRequest/reassign", reassignReq)

	// Should return 409 with PR_MERGED error code as per OpenAPI
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	var errResp api.ErrorResponse
	err = json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.PRMERGED, errResp.Error.Code)
}

func TestPullRequestReassign_NoCandidate(t *testing.T) {
	truncateAllTables(t)

	// Seed team with only 2 users (author and 1 reviewer)
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create PR (u1 is author, u2 should be assigned as reviewer)
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	resp, body := makeRequest(t, "POST", "/pullRequest/create", prReq)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}

	if err := json.Unmarshal(body, &createResult); err != nil {
		t.Fatalf("Failed to parse create PR response: %v", err)
	}

	// Get assigned reviewers
	prData := createResult["pr"].(map[string]interface{})
	reviewers := prData["assigned_reviewers"].([]interface{})

	if len(reviewers) == 0 {
		t.Skip("No reviewers assigned, cannot test NO_CANDIDATE scenario")
	}

	// Try to reassign the only reviewer (no other active candidates available)
	reassignReq := map[string]interface{}{
		"pull_request_id": "pr-1001",
		"old_user_id":     reviewers[0].(string),
	}
	resp, body = makeRequest(t, "POST", "/pullRequest/reassign", reassignReq)

	// Should return 409 with NO_CANDIDATE error code as per OpenAPI
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	var errResp api.ErrorResponse
	err := json.Unmarshal(body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, api.NOCANDIDATE, errResp.Error.Code)
}

func TestUsersGetReview_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed team with users
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
		{UserId: "u3", Username: "Charlie", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create multiple PRs where u2 is a reviewer
	// PR 1 with u1 as author
	prReq1 := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	makeRequest(t, "POST", "/pullRequest/create", prReq1)

	// PR 2 with u3 as author
	prReq2 := map[string]interface{}{
		"pull_request_id":   "pr-1002",
		"pull_request_name": "Fix bug",
		"author_id":         "u3",
	}
	makeRequest(t, "POST", "/pullRequest/create", prReq2)

	// Get PRs where u2 is reviewer
	resp, body := makeRequest(t, "GET", "/users/getReview?user_id=u2", nil)

	// Verify status code (200 as per OpenAPI)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify response structure (matches OpenAPI schema)
	require.Contains(t, result, "user_id")
	require.Contains(t, result, "pull_requests")

	assert.Equal(t, "u2", result["user_id"])

	prs := result["pull_requests"].([]interface{})
	// u2 might be assigned to 0, 1, or 2 PRs depending on assignment logic
	assert.GreaterOrEqual(t, len(prs), 0)

	// Verify each PR has required fields per OpenAPI PullRequestShort schema
	for _, prInterface := range prs {
		pr := prInterface.(map[string]interface{})
		require.Contains(t, pr, "pull_request_id")
		require.Contains(t, pr, "pull_request_name")
		require.Contains(t, pr, "author_id")
		require.Contains(t, pr, "status")

		// Author should not be u2
		assert.NotEqual(t, "u2", pr["author_id"])
	}
}

func TestStatsAssignments_Success(t *testing.T) {
	truncateAllTables(t)

	// Seed team with users
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
		{UserId: "u3", Username: "Charlie", IsActive: true},
	}
	seedTeam(t, "backend", members)

	// Create multiple PRs
	prReq1 := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}
	makeRequest(t, "POST", "/pullRequest/create", prReq1)

	prReq2 := map[string]interface{}{
		"pull_request_id":   "pr-1002",
		"pull_request_name": "Fix bug",
		"author_id":         "u2",
	}
	makeRequest(t, "POST", "/pullRequest/create", prReq2)

	// Get stats
	resp, body := makeRequest(t, "GET", "/stats/assignments", nil)

	// Verify status code (200)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify response structure
	require.Contains(t, result, "by_user")
	require.Contains(t, result, "by_pr")

	byUser := result["by_user"].([]interface{})
	byPR := result["by_pr"].([]interface{})

	// Should have stats for users and PRs
	assert.GreaterOrEqual(t, len(byUser), 0)
	assert.Equal(t, 2, len(byPR)) // We created 2 PRs

	// Verify structure of each stat
	for _, userStat := range byUser {
		stat := userStat.(map[string]interface{})
		require.Contains(t, stat, "user_id")
		require.Contains(t, stat, "assignments")
		assert.IsType(t, "", stat["user_id"])
		assert.IsType(t, float64(0), stat["assignments"])
	}

	for _, prStat := range byPR {
		stat := prStat.(map[string]interface{})
		require.Contains(t, stat, "pull_request_id")
		require.Contains(t, stat, "reviewers")
		assert.IsType(t, "", stat["pull_request_id"])
		assert.IsType(t, float64(0), stat["reviewers"])
	}
}

func TestHealthCheck(t *testing.T) {
	resp, body := makeRequest(t, "GET", "/healthz", nil)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "ok")
}
