package team

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestTeamRepository_CreateTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("team1")

		mock.ExpectExec("INSERT INTO teams").
			WithArgs(teamName).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.CreateTeam(context.Background(), teamName)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("team1")

		mock.ExpectExec("INSERT INTO teams").
			WithArgs(teamName).
			WillReturnError(errors.New("db error"))

		err = repo.CreateTeam(context.Background(), teamName)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamRepository_Exists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("team1")

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(teamName).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), teamName)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("does not exist", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("nonexistent")

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(teamName).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), teamName)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("team1")

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(teamName).
			WillReturnError(errors.New("db error"))

		exists, err := repo.Exists(context.Background(), teamName)

		assert.Error(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamRepository_GetTeamByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("team1")

		rows := sqlmock.NewRows([]string{"team_name"}).AddRow(teamName)
		mock.ExpectQuery("SELECT team_name FROM teams WHERE team_name").
			WithArgs(teamName).
			WillReturnRows(rows)

		team, err := repo.GetTeamByName(context.Background(), teamName)

		assert.NoError(t, err)
		assert.NotNil(t, team)
		assert.Equal(t, teamName, team.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("nonexistent")

		mock.ExpectQuery("SELECT team_name FROM teams WHERE team_name").
			WithArgs(teamName).
			WillReturnError(sql.ErrNoRows)

		team, err := repo.GetTeamByName(context.Background(), teamName)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewTeamRepository(db)
		teamName := domain.TeamName("team1")

		mock.ExpectQuery("SELECT team_name FROM teams WHERE team_name").
			WithArgs(teamName).
			WillReturnError(errors.New("db error"))

		team, err := repo.GetTeamByName(context.Background(), teamName)

		assert.Error(t, err)
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
