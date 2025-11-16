package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Upsert(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		user := &domain.User{
			ID:       "user1",
			Username: "Test User",
			TeamName: "team1",
			IsActive: true,
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Username, user.TeamName, user.IsActive).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Upsert(context.Background(), user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		user := &domain.User{
			ID:       "user1",
			Username: "Test User",
			TeamName: "team1",
			IsActive: true,
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Username, user.TeamName, user.IsActive).
			WillReturnError(errors.New("db error"))

		err = repo.Upsert(context.Background(), user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_SetUserIsActive(t *testing.T) {
	t.Run("success - set to false", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")
		isActive := false

		rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
			AddRow(userID, "Test User", "team1", isActive)

		mock.ExpectQuery("UPDATE users SET is_active").
			WithArgs(isActive, userID).
			WillReturnRows(rows)

		user, err := repo.SetUserIsActive(context.Background(), userID, isActive)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, isActive, user.IsActive)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - set to true", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")
		isActive := true

		rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
			AddRow(userID, "Test User", "team1", isActive)

		mock.ExpectQuery("UPDATE users SET is_active").
			WithArgs(isActive, userID).
			WillReturnRows(rows)

		user, err := repo.SetUserIsActive(context.Background(), userID, isActive)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, isActive, user.IsActive)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("nonexistent")
		isActive := false

		mock.ExpectQuery("UPDATE users SET is_active").
			WithArgs(isActive, userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.SetUserIsActive(context.Background(), userID, isActive)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")
		isActive := false

		mock.ExpectQuery("UPDATE users SET is_active").
			WithArgs(isActive, userID).
			WillReturnError(errors.New("db error"))

		user, err := repo.SetUserIsActive(context.Background(), userID, isActive)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetActiveUsersByTeam(t *testing.T) {
	t.Run("success - multiple users", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		teamName := domain.TeamName("team1")

		rows := sqlmock.NewRows([]string{"user_id"}).
			AddRow("user1").
			AddRow("user2").
			AddRow("user3")

		mock.ExpectQuery("SELECT user_id FROM users WHERE team_name").
			WithArgs(teamName).
			WillReturnRows(rows)

		userIDs, err := repo.GetActiveUsersByTeam(context.Background(), teamName)

		assert.NoError(t, err)
		assert.Len(t, userIDs, 3)
		assert.Contains(t, userIDs, domain.UserID("user1"))
		assert.Contains(t, userIDs, domain.UserID("user2"))
		assert.Contains(t, userIDs, domain.UserID("user3"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - no users", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		teamName := domain.TeamName("team1")

		rows := sqlmock.NewRows([]string{"user_id"})

		mock.ExpectQuery("SELECT user_id FROM users WHERE team_name").
			WithArgs(teamName).
			WillReturnRows(rows)

		userIDs, err := repo.GetActiveUsersByTeam(context.Background(), teamName)

		assert.NoError(t, err)
		assert.Empty(t, userIDs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		teamName := domain.TeamName("team1")

		mock.ExpectQuery("SELECT user_id FROM users WHERE team_name").
			WithArgs(teamName).
			WillReturnError(errors.New("db error"))

		userIDs, err := repo.GetActiveUsersByTeam(context.Background(), teamName)

		assert.Error(t, err)
		assert.Nil(t, userIDs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Exists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(userID).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), userID)

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

		repo := NewUserRepository(db)
		userID := domain.UserID("nonexistent")

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(userID).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), userID)

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

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		exists, err := repo.Exists(context.Background(), userID)

		assert.Error(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")

		rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
			AddRow(userID, "Test User", "team1", true)

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id").
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(context.Background(), userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "Test User", user.Username)
		assert.Equal(t, domain.TeamName("team1"), user.TeamName)
		assert.True(t, user.IsActive)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("nonexistent")

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByID(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userID := domain.UserID("user1")

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		user, err := repo.GetUserByID(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetUsersByIDs(t *testing.T) {
	t.Run("success - multiple users", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userIDs := []domain.UserID{"user1", "user2"}

		rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
			AddRow("user1", "User 1", "team1", true).
			AddRow("user2", "User 2", "team1", true)

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id = ANY").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		users, err := repo.GetUsersByIDs(context.Background(), userIDs)

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - empty list", func(t *testing.T) {
		db, _, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userIDs := []domain.UserID{}

		users, err := repo.GetUsersByIDs(context.Background(), userIDs)

		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		userIDs := []domain.UserID{"user1", "user2"}

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id = ANY").
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))

		users, err := repo.GetUsersByIDs(context.Background(), userIDs)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetUsersByTeamName(t *testing.T) {
	t.Run("success - multiple users", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		teamName := domain.TeamName("team1")

		rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
			AddRow("user1", "User 1", teamName, true).
			AddRow("user2", "User 2", teamName, true)

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active").
			WithArgs(teamName).
			WillReturnRows(rows)

		users, err := repo.GetUsersByTeamName(context.Background(), teamName)

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, domain.UserID("user1"), users[0].ID)
		assert.Equal(t, domain.UserID("user2"), users[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - no users", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		teamName := domain.TeamName("team1")

		rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"})

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active").
			WithArgs(teamName).
			WillReturnRows(rows)

		users, err := repo.GetUsersByTeamName(context.Background(), teamName)

		assert.NoError(t, err)
		assert.Empty(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer func() {
			_ = db.Close()
		}()

		repo := NewUserRepository(db)
		teamName := domain.TeamName("team1")

		mock.ExpectQuery("SELECT user_id, username, team_name, is_active").
			WithArgs(teamName).
			WillReturnError(errors.New("db error"))

		users, err := repo.GetUsersByTeamName(context.Background(), teamName)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
