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

func NewPRService(db *sql.DB, logger logger.InterfaceLogger) *PRService {
	return &PRService{
		prRepo:    NewPRRepository(db),
		userRepo:  user.NewUserRepository(db),
		txManager: tx.NewManager(db),
		logger:    logger,
	}
}

func (s *PRService) CreatePr(
	ctx context.Context,
	pullRequestId domain.PRID,
	authorID domain.UserID,
	title string,
) (*domain.PullRequest, []domain.User, error) {
	var result *domain.PullRequest

	s.logger.Infof("Creating pull request %s by author %s with title %s", pullRequestId, authorID, title)

	var reviewers []domain.User

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		// 1. Check if PR with the same ID already exists
		exists, err := s.prRepo.Exists(txCtx, pullRequestId)
		if err != nil {
			s.logger.Errorf("Failed to check existence of pull request %s: %v", pullRequestId, err)
			return err
		}

		if exists {
			s.logger.Errorf("Pull request %s already exists", pullRequestId)
			return domain.ErrPrAlreadyExists
		}
		// Get author info
		author, err := s.userRepo.GetUserByID(txCtx, authorID)
		if err != nil {
			s.logger.Errorf("Failed to get author %s: %v", authorID, err)
			return err
		}
		// 2. Create PR
		if err := s.prRepo.CreatePr(txCtx, pullRequestId, author.ID, title); err != nil {
			s.logger.Errorf("Failed to create pull request %s: %v", pullRequestId, err)
			return err
		}

		// 3. Get active users of the team
		active, err := s.userRepo.GetActiveUsersByTeam(txCtx, author.TeamName)
		if err != nil {
			s.logger.Errorf("Failed to get active users of team %s: %v", author.TeamName, err)
			return err
		}

		// 4. Remove author from candidates
		candidates := make([]domain.UserID, 0, len(active))

		for _, uid := range active {
			if uid == author.ID {
				continue
			}

			candidates = append(candidates, uid)
		}

		// 5. Pick up to 2 reviewers
		selected := s.pickReviewers(candidates, MaxReviewersPerPr)

		// 6. Save reviewers to the pr_reviewers table (if any selected)
		if len(selected) > 0 {
			for _, reviewerID := range selected {
				if err := s.prRepo.AssignReviewer(txCtx, pullRequestId, reviewerID); err != nil {
					s.logger.Errorf("Failed to assign reviewer %s to pull request %s: %v", reviewerID, pullRequestId, err)
					return err
				}
			}

			s.logger.Infof("Assigned reviewers %v to pull request %s", selected, pullRequestId)
			// Fetch full reviewer details
			s.logger.Infof("Fetching full details for reviewers %v of pull request %s", selected, pullRequestId)

			reviewers, err = s.userRepo.GetUsersByIDs(txCtx, selected)
			if err != nil {
				return err
			}
		}

		// 7. Assemble the domain PR that will be returned from the service
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

	s.logger.Infof("Pull request %s created successfully", pullRequestId)

	return result, reviewers, nil
}

func (s *PRService) pickReviewers(candidates []domain.UserID, max int) []domain.UserID {
	if len(candidates) == 0 {
		return nil
	}

	if len(candidates) <= max {
		// Create a copy to avoid modifying the original slice
		s.logger.Infof("Number of candidates (%d) is less than or equal to max reviewers (%d), returning all candidates", len(candidates), max)
		result := make([]domain.UserID, len(candidates))
		copy(result, candidates)

		return result
	}

	// Create a copy to avoid modifying the original slice
	shuffled := make([]domain.UserID, len(candidates))
	copy(shuffled, candidates)

	// Shuffle using Fisher-Yates algorithm with crypto/rand
	for i := len(shuffled) - 1; i > 0; i-- {
		// Generate random index from 0 to i (inclusive)
		var b [8]byte
		if _, err := rand.Read(b[:]); err != nil {
			// Fallback to first N candidates if crypto/rand fails
			return shuffled[:max]
		}
		// Convert bytes to uint64 and get random index
		randomValue := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
			uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
		// #nosec G115 - i+1 is always positive and bounded by slice length
		j := randomValue % uint64(i+1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	s.logger.Infof("Selected reviewers after shuffling: %v", shuffled[:max])

	return shuffled[:max]
}

func (s *PRService) MergePr(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, []domain.User, error) {
	var result *domain.PullRequest

	var reviewers []domain.User

	s.logger.Infof("Merging pull request %s", pullRequestId)

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		s.logger.Infof("Checking existence of pull request %s", pullRequestId)

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

		// Fetch reviewer details
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

	s.logger.Infof("Pull request %s merged successfully", pullRequestId)

	return result, reviewers, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, pullRequestId domain.PRID, oldReviewerId domain.UserID) (*domain.ReassignResult, error) {
	// 1. Предпроверки (PR и пользователь существуют)
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

	// 2. PR не должен быть MERGED
	merged, err := s.prRepo.CheckIsMerged(ctx, pullRequestId)
	if err != nil {
		s.logger.Errorf("Failed to check if pull request %s is merged: %v", pullRequestId, err)
		return nil, err
	}

	if merged {
		s.logger.Errorf("Pull request %s is already merged", pullRequestId)
		return nil, domain.ErrPrMerged
	}

	// 3. Проверяем, что этот ревьювер назначен именно на этот PR
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

	// 4. Транзакция: вся модификация состояния
	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		s.logger.Infof("Getting pull request %s", pullRequestId)
		// 4.1. Получаем PR
		pr, err := s.prRepo.GetPrByPrID(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		// 4.2. Автор
		s.logger.Infof("Getting old reviewer %s", oldReviewerId)

		oldReviewer, err := s.userRepo.GetUserByID(txCtx, oldReviewerId)
		if err != nil {
			return err
		}

		// 4.3. Активные юзеры команды
		s.logger.Infof("Getting active users for team %s", oldReviewer.TeamName)

		activeUsers, err := s.userRepo.GetActiveUsersByTeam(txCtx, oldReviewer.TeamName)
		if err != nil {
			s.logger.Errorf("Failed to get active users for team %s: %v", oldReviewer.TeamName, err)
			return err
		}

		// 4.4. Кандидаты = активные минус автор и старый ревьювер
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

		// 4.5. Выбираем нового ревьювера
		picked := s.pickReviewers(candidates, 1)
		newID := picked[0]

		// 4.6. Разруливаем в БД: снимаем старого, назначаем нового
		if err := s.prRepo.UnassignReviewer(txCtx, pullRequestId, oldReviewerId); err != nil {
			return err
		}

		if err := s.prRepo.AssignReviewer(txCtx, pullRequestId, newID); err != nil {
			return err
		}

		// 4.7. Обновляем список ревьюверов в памяти
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
