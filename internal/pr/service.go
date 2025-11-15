package pr

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
	"github.com/merkulovlad/avito-internship-test/internal/user"
)

// PRService provides methods to manage pull requests.

type PRService struct {
	prRepo    *PRRepository
	userRepo  *user.UserRepository
	txManager *tx.Manager
}

func NewPRService(db *sql.DB) *PRService {
	return &PRService{
		prRepo:    NewPRRepository(db),
		userRepo:  user.NewUserRepository(db),
		txManager: tx.NewManager(db),
	}
}

func (s *PRService) CreatePr(
	ctx context.Context,
	pullRequestId domain.PRID,
	author *domain.User,
	title string,
) (*domain.PullRequest, error) {
	var result *domain.PullRequest

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		// 1. Check if PR with the same ID already exists
		exists, err := s.prRepo.Exists(txCtx, pullRequestId)
		if err != nil {
			return err
		}
		if exists {
			return domain.ErrPrAlreadyExists
		}

		// 2. Create PR
		if err := s.prRepo.CreatePr(txCtx, pullRequestId, author.ID, title); err != nil {
			return err
		}

		// 3. Get active users of the team
		active, err := s.userRepo.GetActiveUsersByTeam(txCtx, author.TeamName)
		if err != nil {
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
		selected := pickReviewers(candidates, 2)

		// 6. Save reviewers to the pr_reviewers table (if any selected)
		if len(selected) > 0 {
			for _, reviewerID := range selected {
				if err := s.prRepo.AssignReviewer(txCtx, pullRequestId, reviewerID); err != nil {
					return err
				}
			}
		}

		// 7. Assemble the domain PR that will be returned from the service
		result = &domain.PullRequest{
			ID:                domain.PRID(pullRequestId),
			AuthorID:          author.ID,
			Title:             title,
			CreatedAt:         time.Now(),
			IsMerged:          false,
			AssignedReviewers: selected,
			MergedAt:          nil,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func pickReviewers(candidates []domain.UserID, max int) []domain.UserID {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) <= max {
		return candidates
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	return candidates[:max]
}

func (s *PRService) MergePr(ctx context.Context, pullRequestId domain.PRID) (*domain.PullRequest, error) {
	if res, err := s.prRepo.Exists(ctx, pullRequestId); err != nil {
		return nil, err
	} else if !res {
		return nil, domain.ErrNotFound
	}

	pullReq, err := s.prRepo.MergePr(ctx, pullRequestId)
	if err != nil {
		return nil, err
	}

	return pullReq, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, pullRequestId domain.PRID, oldReviewerId domain.UserID) (*domain.ReassignResult, error) {
	// 1. Предпроверки (PR и пользователь существуют)
	if exists, err := s.prRepo.Exists(ctx, pullRequestId); err != nil {
		return nil, err
	} else if !exists {
		return nil, domain.ErrNotFound
	}

	if exists, err := s.userRepo.Exists(ctx, oldReviewerId); err != nil {
		return nil, err
	} else if !exists {
		return nil, domain.ErrNotFound
	}

	// 2. PR не должен быть MERGED
	merged, err := s.prRepo.CheckIsMerged(ctx, pullRequestId)
	if err != nil {
		return nil, err
	}
	if merged {
		return nil, domain.ErrPrMerged
	}

	// 3. Проверяем, что этот ревьювер назначен именно на этот PR
	assigned, err := s.prRepo.IsReviewerAssigned(ctx, pullRequestId, oldReviewerId)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, domain.ErrNotAssigned
	}

	var (
		updatedPR     *domain.PullRequest
		newReviewerID domain.UserID
	)

	// 4. Транзакция: вся модификация состояния
	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		// 4.1. Получаем PR
		pr, err := s.prRepo.GetPrByID(txCtx, pullRequestId)
		if err != nil {
			return err
		}

		// 4.2. Автор
		author, err := s.userRepo.GetUserByID(txCtx, pr.AuthorID)
		if err != nil {
			return err
		}

		// 4.3. Активные юзеры команды
		activeUsers, err := s.userRepo.GetActiveUsersByTeam(txCtx, author.TeamName)
		if err != nil {
			return err
		}

		// 4.4. Кандидаты = активные минус автор и старый ревьювер
		candidates := make([]domain.UserID, 0, len(activeUsers))
		for _, u := range activeUsers {
			id := u
			if id == author.ID || id == oldReviewerId {
				continue
			}
			candidates = append(candidates, id)
		}

		if len(candidates) == 0 {
			return domain.ErrNoCandidate
		}

		// 4.5. Выбираем нового ревьювера
		picked := pickReviewers(candidates, 1)
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
