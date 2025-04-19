package postgresqlrepo

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func getMockUserRepo(t *testing.T) (repository.IUserRepo, sqlmock.Sqlmock, func()) {
	logger := zap.NewNop()

	mockDB, mock, err := sqlmock.New()

	require.NoError(t, err, "creating mock db")

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	repo := NewPostgresqlUserRepo(sqlxDB, logger)

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestPostgresqlUserRepo_Create(t *testing.T) {
	repo, mock, cleanup := getMockUserRepo(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		guid := uuid.New()
		email := "test@test.ru"

		mock.ExpectExec("INSERT INTO users").
			WithArgs(guid, email).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), guid, email)
		assert.NoError(t, err)
	})

	t.Run("User exists", func(t *testing.T) {
		guid := uuid.New()
		email := "test@test.ru"

		mock.ExpectExec("INSERT INTO users").
			WithArgs(guid, email).
			WillReturnError(&pq.Error{Code: database.PGUniqueViolationCode})

		err := repo.Create(context.Background(), guid, email)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserExists)
	})

}

func TestPostgresqlUserRepo_GetEmail(t *testing.T) {
	repo, mock, cleanup := getMockUserRepo(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		guid := uuid.New()
		mock.ExpectQuery("SELECT email FROM users").
			WithArgs(guid).
			WillReturnRows(sqlmock.NewRows([]string{"email"}).
				AddRow("test@test.ru"))

		email, err := repo.GetEmail(context.Background(), guid)
		assert.NoError(t, err)
		assert.Equal(t, "test@test.ru", email)
	})

	t.Run("User not found", func(t *testing.T) {
		guid := uuid.New()
		mock.ExpectQuery("SELECT email FROM users").
			WithArgs(guid).
			WillReturnError(sql.ErrNoRows)

		email, err := repo.GetEmail(context.Background(), guid)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Empty(t, email)
	})
}
