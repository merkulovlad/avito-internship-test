// Package pr implements pull request management business logic.
package pr

import (
	"context"
	"crypto/rand"
	"database/sql"
	"time"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
	"github.com/merkulovlad/avito-internship-test/internal/user"
)

// MaxReviewersPerPr is the maximum number of reviewers that can be assigned to a pull request.
const MaxReviewersPerPr = 2

// Compile-time interface check
var _ domain.PRServiceInterface = (*PRService)(nil)

// PRService provides methods to manage pull requests.
type PRService struct {
	prRepo    domain.PullRequestRepositoryInterface
	userRepo  domain.UserRepositoryInterface
	txManager *tx.Manager
	logger    logger.InterfaceLogger
}

// NewPRService creates a new PRService instance.
func NewPRService(db *sql.DB, logger logger.InterfaceLogger) *PRService {
	return &PRService{
		prRepo:    NewPRRepository(db),
		userRepo:  user.NewUserRepository(db),
		txManager: tx.NewManager(db),
		logger:    logger,
	}
}

// CreatePr creates a new pull request and assigns reviewers from the author's team.
func (s *PRService) CreatePr(
	ctx context.Context,
	pullRequestId domain.PRID,
	authorID domain.UserID,
	title string,
) (*domain.PullRequest, []domain.User, error) {
	var result *domain.PullRequest

	var reviewers []domain.User

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		exists, err := s.prRepo.Exists(txCtx, pullRequestId)
		if err != nil {
			s.logger.Errorf("Failed to check existence of pull request %s: %v", pullRequestId, err)
			return err
		}

		if exists {
			s.logger.Errorf("Pull request %s already exists", pullRequestId)
			return domain.ErrPrAlreadyExists
		}

		author, err := s.userRepo.GetUserByID(txCtx, authorID)
		if err != nil {
			s.logger.Errorf("Failed to get author %s: %v", authorID, err)
			return err
		}

		if err := s.prRepo.CreatePr(txCtx, pullRequestId, author.ID, title); err != nil {
			s.logger.Errorf("Failed to create pull request %s: %v", pullRequestId, err)
			return err
		}

		active, err := s.userRepo.GetActiveUsersByTeam(txCtx, author.TeamName)
		if err != nil {
			s.logger.Errorf("Failed to get active users of team %s: %v", author.TeamName, err)
			return err
		}

		candidates := make([]domain.UserID, 0, len(active))

		for _, uid := range active {
			if uid == author.ID {
				continue
			}

			candidates = append(candidates, uid)
		}

		selected := s.pickReviewers(candidates, MaxReviewersPerPr)

		if len(selected) > 0 {
			for _, reviewerID := range selected {
				if err := s.prRepo.AssignReviewer(txCtx, pullRequestId, reviewerID); err != nil {
					s.logger.Errorf("Failed to assign reviewer %s to pull request %s: %v", reviewerID, pullRequestId, err)
					return err
				}
			}

			reviewers, err = s.userRepo.GetUsersByIDs(txCtx, selected)
			if err != nil {
				return err
			}
		}

		result = &domain.PullRequest{
			ID:                domain.PRID(pullRequestId),
			AuthorID:          author.ID,
			Title:             title,
			CreatedAt:         time.Now().UTC(),
			IsMerged:          false,
			AssignedReviewers: selected,
			MergedAt:          nil,
		}

		return nil
	})
	if err != nil {
		s.logger.Errorf("Failed to create pull request %s: %v", pullRequestId, err)
		return nil, nil, err
	}

	return result, reviewers, nil
}

// pickReviewers randomly selects up to max reviewers from the candidates list.
func (s *PRService) pickReviewers(candidates []domain.UserID, max int) []domain.UserID {
	if len(candidates) == 0 {
		return nil
	}

	if len(candidates) <= max {
		result := make([]domain.UserID, len(candidates))
		copy(result, candidates)

		return result
	}

	shuffled := make([]domain.UserID, len(candidates))
	copy(shuffled, candidates)

	for i := len(shuffled) - 1; i > 0; i-- {
		var b [8]byte
		if _, err := rand.Read(b[:]); err != nil {
			return shuffled[:max]
		}

		randomValue := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
			uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
		// #nosec G115 - value always positive and guaranteed to fit into int
		j := randomValue % uint64(i+1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:max]
}

// MergePr marks a pull request as merged.
func (s *PRService) MergePr(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, []domain.User, error) {
	var result *domain.PullRequest

	var reviewers []domain.User

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		exists, err := s.prRepo.Exists(txCtx, pullRequestId)
		if err != nil {
			s.logger.Errorf("Failed to check existence of pull request %s: %v", pullRequestId, err)
			return err
		}

		if !exists {
			s.logger.Errorf("Pull request %s not found", pullRequestId)
			return domain.ErrNotFound
		}

		pullReq, err := s.prRepo.MergePr(txCtx, pullRequestId)
		if err != nil {
			s.logger.Errorf("Failed to merge pull request %s: %v", pullRequestId, err)
			return err
		}

		reviewerIds, err := s.prRepo.GetReviewers(txCtx, pullRequestId)
		if err != nil {
			s.logger.Errorf("Failed to get reviewers for pull request %s: %v", pullRequestId, err)
			return err
		}

		if len(reviewerIds) > 0 {
			reviewers, err = s.userRepo.GetUsersByIDs(txCtx, reviewerIds)
			if err != nil {
				s.logger.Errorf("Failed to get reviewer details for pull request %s: %v", pullRequestId, err)
				return err
			}
		}

		result = pullReq

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return result, reviewers, nil
}

// ReassignReviewer replaces an assigned reviewer with a new one from the same team.
func (s *PRService) ReassignReviewer(ctx context.Context, pullRequestId domain.PRID, oldReviewerId domain.UserID) (*domain.ReassignResult, error) {
	if exists, err := s.prRepo.Exists(ctx, pullRequestId); err != nil {
		return nil, err
	} else if !exists {
		s.logger.Errorf("Pull request %s not found", pullRequestId)
		return nil, domain.ErrNotFound
	}

	if exists, err := s.userRepo.Exists(ctx, oldReviewerId); err != nil {
		s.logger.Errorf("Failed to check existence of user %s: %v", oldReviewerId, err)
		return nil, err
	} else if !exists {
		s.logger.Errorf("User %s not found", oldReviewerId)
		return nil, domain.ErrNotFound
	}

	merged, err := s.prRepo.CheckIsMerged(ctx, pullRequestId)
	if err != nil {
		s.logger.Errorf("Failed to check if pull request %s is merged: %v", pullRequestId, err)
		return nil, err
	}

	if merged {
		s.logger.Errorf("Pull request %s is already merged", pullRequestId)
		return nil, domain.ErrPrMerged
	}

	assigned, err := s.prRepo.IsReviewerAssigned(ctx, pullRequestId, oldReviewerId)
	if err != nil {
		s.logger.Errorf("Failed to check if reviewer %s is assigned to pull request %s: %v", oldReviewerId, pullRequestId, err)
		return nil, err
	}

	if !assigned {
		s.logger.Errorf("Reviewer %s is not assigned to pull request %s", oldReviewerId, pullRequestId)
		return nil, domain.ErrNotAssigned
	}

	var (
		updatedPR     *domain.PullRequest
		newReviewerID domain.UserID
	)

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		pr, err := s.prRepo.GetPrByPrID(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		oldReviewer, err := s.userRepo.GetUserByID(txCtx, oldReviewerId)
		if err != nil {
			return err
		}

		activeUsers, err := s.userRepo.GetActiveUsersByTeam(txCtx, oldReviewer.TeamName)
		if err != nil {
			s.logger.Errorf("Failed to get active users for team %s: %v", oldReviewer.TeamName, err)
			return err
		}

		candidates := make([]domain.UserID, 0, len(activeUsers))

		for _, u := range activeUsers {
			id := u
			if id == oldReviewer.ID || id == pr.AuthorID {
				continue
			}

			candidates = append(candidates, id)
		}

		if len(candidates) == 0 {
			return domain.ErrNoCandidate
		}

		picked := s.pickReviewers(candidates, 1)
		newID := picked[0]

		if err := s.prRepo.UnassignReviewer(txCtx, pullRequestId, oldReviewerId); err != nil {
			return err
		}

		if err := s.prRepo.AssignReviewer(txCtx, pullRequestId, newID); err != nil {
			return err
		}

		newReviewers := make([]domain.UserID, 0, len(pr.AssignedReviewers))

		for _, r := range pr.AssignedReviewers {
			if r == oldReviewerId {
				continue
			}

			newReviewers = append(newReviewers, r)
		}

		newReviewers = append(newReviewers, newID)
		pr.AssignedReviewers = newReviewers

		updatedPR = pr
		newReviewerID = newID

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.ReassignResult{
		PR:         updatedPR,
		ReplacedBy: newReviewerID,
	}, nil
}

func (s *PRService) GetReviewers(ctx context.Context, pullRequestId domain.PRID) ([]domain.User, error) {
	// Check if PR exists first
	s.logger.Infof("Checking existence of pull request %s", pullRequestId)

	exists, err := s.prRepo.Exists(ctx, pullRequestId)
	if err != nil {
		s.logger.Errorf("Failed to check existence of pull request %s: %v", pullRequestId, err)
		return nil, err
	}

	if !exists {
		s.logger.Errorf("Pull request %s not found", pullRequestId)
		return nil, domain.ErrNotFound
	}

	reviewerIds, err := s.prRepo.GetReviewers(ctx, pullRequestId)
	if err != nil {
		s.logger.Errorf("Failed to get reviewers for pull request %s: %v", pullRequestId, err)
		return nil, err
	}

	reviewers, err := s.userRepo.GetUsersByIDs(ctx, reviewerIds)
	if err != nil {
		s.logger.Errorf("Failed to get users by IDs for pull request %s: %v", pullRequestId, err)
		return nil, err
	}

	s.logger.Infof("Fetched reviewers for pull request %s successfully", pullRequestId)

	return reviewers, nil
}

func (s *PRService) GetAssignmentStats(ctx context.Context) (*domain.Stats, error) {
	s.logger.Infof("Fetching assignment statistics")

	stats, err := s.prRepo.GetAssignmentStats(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get assignment statistics: %v", err)
		return nil, err
	}

	s.logger.Infof("Successfully fetched assignment statistics")

	return stats, nil
}
