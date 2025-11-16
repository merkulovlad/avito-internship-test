package domain

import "context"

// UserServiceInterface defines methods for user business logic
type UserServiceInterface interface {
	SetIsActive(ctx context.Context, id UserID, isActive bool) (*User, error)
	GetPrOfUser(ctx context.Context, userId UserID) ([]PullRequest, error)
}

// TeamServiceInterface defines methods for team business logic
type TeamServiceInterface interface {
	CreateTeam(ctx context.Context, t TeamName, members []User) (*Team, error)
	GetTeamByName(ctx context.Context, name TeamName) (*Team, error)
	GetTeamMembers(ctx context.Context, teamName string) ([]User, error)
}

// PRServiceInterface defines methods for pull request business logic
type PRServiceInterface interface {
	CreatePr(ctx context.Context, pullRequestId PRID, authorID UserID, title string) (*PullRequest, []User, error)
	MergePr(ctx context.Context, pullRequestId PRID) (*PullRequest, []User, error)
	ReassignReviewer(ctx context.Context, pullRequestId PRID, oldReviewerId UserID) (*ReassignResult, error)
	GetReviewers(ctx context.Context, pullRequestId PRID) ([]User, error)
}
