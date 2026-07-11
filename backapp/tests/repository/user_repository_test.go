package repository_test

import (
	"backapp/internal/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func setupUser(t *testing.T) (repository.UserRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return repository.NewUserRepository(db), mock, func() { db.Close() }
}
