package team

import (
	"context"
	"database/sql"

	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
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
	logger         logger.InterfaceLogger
}

func NewTeamService(db *sql.DB, logger logger.InterfaceLogger) *TeamService {
	return &TeamService{
		teamRepository: NewTeamRepository(db),
		userRepository: user.NewUserRepository(db),
		txManager:      tx.NewManager(db),
		logger:         logger,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, t domain.TeamName, members []domain.User) (*domain.Team, error) {
	var res *domain.Team

	s.logger.Infof("Creating team %s", t)

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		exists, err := s.teamRepository.Exists(txCtx, t)
		if err != nil {
			s.logger.Errorf("Failed to check existence of team %s: %v", t, err)
			return err
		}

		if exists {
			s.logger.Errorf("Team %s already exists", t)
			return domain.ErrTeamAlreadyExists
		}

		if err := s.teamRepository.CreateTeam(txCtx, t); err != nil {
			s.logger.Errorf("Failed to create team %s: %v", t, err)
			return err
		}

		for _, member := range members {
			member.TeamName = t

			if err := s.userRepository.Upsert(txCtx, &member); err != nil {
				s.logger.Errorf("Failed to add member %s to team %s: %v", member.ID, t, err)
				return err
			}
		}

		s.logger.Infof("Team %s created successfully", t)
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
	s.logger.Infof("Getting team by name %s", name)
	return s.teamRepository.GetTeamByName(ctx, name)
}

func (s *TeamService) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	s.logger.Infof("Getting members of team %s", teamName)
	// Check if team exists first
	exists, err := s.teamRepository.Exists(ctx, domain.TeamName(teamName))
	if err != nil {
		s.logger.Errorf("Failed to check existence of team %s: %v", teamName, err)
		return nil, err
	}

	if !exists {
		s.logger.Errorf("Team %s not found", teamName)
		return nil, domain.ErrNotFound
	}

	s.logger.Infof("Team %s exists, fetching members", teamName)

	return s.userRepository.GetUsersByTeamName(ctx, domain.TeamName(teamName))
}
