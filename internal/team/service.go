package team

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/tx"
	"github.com/merkulovlad/avito-internship-test/internal/user"
)

// TeamService provides methods to manage teams and their members.

type TeamService struct {
	teamRepository *TeamRepository
	userRepository *user.UserRepository
	txManager      *tx.Manager
}

func NewTeamService(db *sql.DB) *TeamService {
	return &TeamService{
		teamRepository: NewTeamRepository(db),
		userRepository: user.NewUserRepository(db),
		txManager:      tx.NewManager(db),
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, t *domain.Team, members []*domain.User) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		exists, err := s.teamRepository.Exists(txCtx, t.Name)
		if err != nil {
			return err
		}

		if exists {
			return domain.ErrTeamAlreadyExists
		}

		if err := s.teamRepository.CreateTeam(txCtx, *t); err != nil {
			return err
		}

		for _, member := range members {
			member.TeamName = t.Name

			if err := s.userRepository.Upsert(txCtx, member); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *TeamService) GetTeamByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	return s.teamRepository.GetTeamByName(ctx, name)
}
