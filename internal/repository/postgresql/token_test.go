package postgresqlrepo_test

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/maksemen2/medods-task/internal/domain"
	"github.com/maksemen2/medods-task/internal/pkg/database"
	"github.com/maksemen2/medods-task/internal/repository"
	"github.com/maksemen2/medods-task/internal/repository/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

func getMockTokenRepo(t *testing.T) (repository.ITokenRepo, sqlmock.Sqlmock, func()) {
	logger := zap.NewNop()

	mockDB, mock, err := sqlmock.New()

	require.NoError(t, err, "creating mock db")

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	repo := postgresqlrepo.NewPostgresqlTokenRepo(sqlxDB, logger)

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestPostgresqlTokenRepo_Create(t *testing.T) {
	repo, mock, cleanup := getMockTokenRepo(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		id := uuid.New()
		jti := uuid.New()
		userID := uuid.New()
		token := "test_token"
		exp := time.Now().Add(1 * time.Hour)
		mock.ExpectExec("INSERT INTO tokens").
			WithArgs(id, jti, userID, token, exp).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), id, jti, userID, token, exp)
		assert.NoError(t, err)
	})

	t.Run("Token Exists", func(t *testing.T) {
		id := uuid.New()
		jti := uuid.New()
		userID := uuid.New()
		token := "test_token"
		exp := time.Now().Add(1 * time.Hour)
		mock.ExpectExec("INSERT INTO tokens").
			WithArgs(id, jti, userID, token, exp).
			WillReturnError(&pq.Error{Code: database.PGUniqueViolationCode})

		err := repo.Create(context.Background(), id, jti, userID, token, exp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrTokenExists)
	})
}

func TestPostgresqlTokenRepo_GetToken(t *testing.T) {
	repo, mock, cleanup := getMockTokenRepo(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		jti := uuid.New()
		notAfter := time.Now()
		expectedID := uuid.New()
		expectedToken := "test_token"

		mock.ExpectQuery("SELECT id, token FROM tokens").
			WithArgs(userID, jti, notAfter).
			WillReturnRows(sqlmock.NewRows([]string{"id", "token"}).
				AddRow(expectedID, expectedToken))

		tokenID, token, err := repo.GetToken(context.Background(), userID, jti, notAfter)
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		assert.Equal(t, expectedID, tokenID)
	})

	t.Run("Token not found", func(t *testing.T) {
		userID := uuid.New()
		jti := uuid.New()
		notAfter := time.Now()

		mock.ExpectQuery("SELECT id, token FROM tokens").
			WithArgs(userID, jti, notAfter).
			WillReturnError(sql.ErrNoRows)

		tokenID, token, err := repo.GetToken(context.Background(), userID, jti, notAfter)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, uuid.Nil, tokenID)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func TestPostgresqlTokenRepo_RotateToken(t *testing.T) {
	repo, mock, cleanup := getMockTokenRepo(t)

	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		oldID := uuid.New()
		id := uuid.New()
		jti := uuid.New()
		userID := uuid.New()
		token := "test_token"
		exp := time.Now().Add(1 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM tokens").
			WithArgs(oldID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO tokens").
			WithArgs(id, jti, userID, token, exp).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RotateToken(context.Background(), oldID, id, jti, userID, token, exp)
		assert.NoError(t, err)
	})

	t.Run("Token exists", func(t *testing.T) {
		oldID := uuid.New()
		id := uuid.New()
		jti := uuid.New()
		userID := uuid.New()
		token := "test_token"
		exp := time.Now().Add(1 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM tokens").
			WithArgs(oldID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO tokens").
			WithArgs(id, jti, userID, token, exp).
			WillReturnError(&pq.Error{Code: database.PGUniqueViolationCode})
		mock.ExpectRollback()

		err := repo.RotateToken(context.Background(), oldID, id, jti, userID, token, exp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrTokenExists)
	})
}
