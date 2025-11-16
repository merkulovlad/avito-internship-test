package domain

import "context"

// UserServiceInterface defines methods for user business logic.
type UserServiceInterface interface {
	// SetIsActive updates the active status of a user.
	SetIsActive(ctx context.Context, id UserID, isActive bool) (*User, error)
	// GetPrOfUser retrieves all pull requests assigned to a user.
	GetPrOfUser(ctx context.Context, userId UserID) ([]PullRequest, error)
}

// TeamServiceInterface defines methods for team business logic.
type TeamServiceInterface interface {
	// CreateTeam creates a new team with the given members.
	CreateTeam(ctx context.Context, t TeamName, members []User) (*Team, error)
	// GetTeamByName retrieves a team by its name.
	GetTeamByName(ctx context.Context, name TeamName) (*Team, error)
	// GetTeamMembers retrieves all members of a team.
	GetTeamMembers(ctx context.Context, teamName string) ([]User, error)
}

// PRServiceInterface defines methods for pull request business logic.
type PRServiceInterface interface {
	// CreatePr creates a new pull request and assigns reviewers.
	CreatePr(ctx context.Context, pullRequestId PRID, authorID UserID, title string) (*PullRequest, []User, error)
	// MergePr merges a pull request.
	MergePr(ctx context.Context, pullRequestId PRID) (*PullRequest, []User, error)
	// ReassignReviewer replaces an assigned reviewer with a new one.
	ReassignReviewer(ctx context.Context, pullRequestId PRID, oldReviewerId UserID) (*ReassignResult, error)
	// GetReviewers retrieves all reviewers for a pull request.
	GetReviewers(ctx context.Context, pullRequestId PRID) ([]User, error)
	// GetAssignmentStats retrieves statistics about reviewer assignments.
	GetAssignmentStats(ctx context.Context) (*Stats, error)
}
