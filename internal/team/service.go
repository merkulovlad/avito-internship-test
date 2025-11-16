package team

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
	"github.com/merkulovlad/avito-internship-test/internal/user"
)

// Compile-time interface check
var _ domain.TeamServiceInterface = (*TeamService)(nil)

// TeamService provides methods to manage teams and their members.

type TeamService struct {
	teamRepository domain.TeamRepositoryInterface
	userRepository domain.UserRepositoryInterface
	txManager      *tx.Manager
}

func NewTeamService(db *sql.DB) *TeamService {
	return &TeamService{
		teamRepository: NewTeamRepository(db),
		userRepository: user.NewUserRepository(db),
		txManager:      tx.NewManager(db),
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, t domain.TeamName, members []domain.User) (*domain.Team, error) {
	var res *domain.Team
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		exists, err := s.teamRepository.Exists(txCtx, t)
		if err != nil {
			return err
		}

		if exists {
			return domain.ErrTeamAlreadyExists
		}

		if err := s.teamRepository.CreateTeam(txCtx, t); err != nil {
			return err
		}

		for _, member := range members {
			member.TeamName = t

			if err := s.userRepository.Upsert(txCtx, &member); err != nil {
				return err
			}
		}
		res = &domain.Team{
			Name: t,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *TeamService) GetTeamByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	return s.teamRepository.GetTeamByName(ctx, name)
}

func (s *TeamService) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	// Check if team exists first
	exists, err := s.teamRepository.Exists(ctx, domain.TeamName(teamName))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrNotFound
	}

	return s.userRepository.GetUsersByTeamName(ctx, domain.TeamName(teamName))
}
