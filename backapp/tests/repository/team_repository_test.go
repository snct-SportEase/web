package repository_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTeam(t *testing.T) (repository.TeamRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return repository.NewTeamRepository(db), mock, func() { db.Close() }
}

// userCols lists the columns returned by team-member SELECT queries.
var userCols = []string{"id", "email", "display_name", "class_id", "is_profile_complete", "created_at", "updated_at"}

// ─── CreateTeam ────────────────────────────────────────────────────────────

func TestTeamRepository_CreateTeam(t *testing.T) {
	const q = "INSERT INTO teams (name, class_id, sport_id) VALUES (?, ?, ?)"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		team := &models.Team{Name: "IS3-A", ClassID: 10, SportID: 1}
		mock.ExpectExec(regexp.QuoteMeta(q)).
			WithArgs(team.Name, team.ClassID, team.SportID).
			WillReturnResult(sqlmock.NewResult(5, 1))

		id, err := repo.CreateTeam(team)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		id, err := repo.CreateTeam(&models.Team{Name: "X", ClassID: 1, SportID: 1})
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── DeleteTeamsByEventAndSportID ──────────────────────────────────────────

func TestTeamRepository_DeleteTeamsByEventAndSportID(t *testing.T) {
	const q = `
			DELETE t FROM teams t
			JOIN classes c ON t.class_id = c.id
			WHERE c.event_id = ? AND t.sport_id = ?
		`

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).
			WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 3))

		assert.NoError(t, repo.DeleteTeamsByEventAndSportID(1, 2))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.DeleteTeamsByEventAndSportID(1, 2))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamsByUserID ──────────────────────────────────────────────────────

func TestTeamRepository_GetTeamsByUserID(t *testing.T) {
	const q = `
			SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, s.name as sport_name
			FROM teams t
			INNER JOIN team_members tm ON t.id = tm.team_id
			INNER JOIN sports s ON t.sport_id = s.id
			INNER JOIN classes c ON t.class_id = c.id
			WHERE tm.user_id = ?
		`
	cols := []string{"id", "name", "class_id", "sport_id", "event_id", "sport_name"}

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("user-1").
			WillReturnRows(sqlmock.NewRows(cols).AddRow(10, "IS3-A", 5, 1, 1, "バスケ"))

		teams, err := repo.GetTeamsByUserID("user-1")
		require.NoError(t, err)
		assert.Len(t, teams, 1)
		assert.Equal(t, 10, teams[0].ID)
		assert.Equal(t, "バスケ", teams[0].SportName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("no-team-user").
			WillReturnRows(sqlmock.NewRows(cols))

		teams, err := repo.GetTeamsByUserID("no-team-user")
		assert.NoError(t, err)
		assert.Nil(t, teams)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		teams, err := repo.GetTeamsByUserID("user-1")
		assert.Error(t, err)
		assert.Nil(t, teams)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamsByClassID ─────────────────────────────────────────────────────

func TestTeamRepository_GetTeamsByClassID(t *testing.T) {
	const q = `
			SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, s.name as sport_name
			FROM teams t
			INNER JOIN sports s ON t.sport_id = s.id
			INNER JOIN classes c ON t.class_id = c.id
			WHERE t.class_id = ? AND c.event_id = ?
		`
	cols := []string{"id", "name", "class_id", "sport_id", "event_id", "sport_name"}

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(5, 1).
			WillReturnRows(sqlmock.NewRows(cols).
				AddRow(10, "IS3-A", 5, 1, 1, "バスケ").
				AddRow(11, "IS3-B", 5, 2, 1, "サッカー"))

		teams, err := repo.GetTeamsByClassID(5, 1)
		require.NoError(t, err)
		assert.Len(t, teams, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows(cols))

		teams, err := repo.GetTeamsByClassID(999, 1)
		assert.NoError(t, err)
		assert.Nil(t, teams)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		teams, err := repo.GetTeamsByClassID(5, 1)
		assert.Error(t, err)
		assert.Nil(t, teams)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamByClassAndSport ────────────────────────────────────────────────

func TestTeamRepository_GetTeamByClassAndSport(t *testing.T) {
	const q = "SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, t.min_capacity, t.max_capacity FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.class_id = ? AND t.sport_id = ? AND c.event_id = ?"

	t.Run("success with capacity", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(5, 1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id", "min_capacity", "max_capacity"}).
				AddRow(10, "IS3-A", 5, 1, 1, sql.NullInt64{Int64: 3, Valid: true}, sql.NullInt64{Int64: 8, Valid: true}))

		team, err := repo.GetTeamByClassAndSport(5, 1, 1)
		require.NoError(t, err)
		require.NotNil(t, team)
		assert.Equal(t, 10, team.ID)
		assert.Equal(t, 3, *team.MinCapacity)
		assert.Equal(t, 8, *team.MaxCapacity)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999, 1, 1).WillReturnError(sql.ErrNoRows)

		team, err := repo.GetTeamByClassAndSport(999, 1, 1)
		assert.NoError(t, err)
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		team, err := repo.GetTeamByClassAndSport(5, 1, 1)
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── AddTeamMember ─────────────────────────────────────────────────────────

func TestTeamRepository_AddTeamMember(t *testing.T) {
	const q = "INSERT INTO team_members (team_id, user_id) VALUES (?, ?)"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(10, "user-1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, repo.AddTeamMember(10, "user-1"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.AddTeamMember(10, "user-1"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamMembers ────────────────────────────────────────────────────────

func TestTeamRepository_GetTeamMembers(t *testing.T) {
	const q = `
			SELECT u.id, u.email, u.display_name, u.class_id, u.is_profile_complete, u.created_at, u.updated_at
			FROM users u
			INNER JOIN team_members tm ON u.id = tm.user_id
			WHERE tm.team_id = ?
			ORDER BY u.display_name, u.email
		`

	now := time.Now()

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(10).
			WillReturnRows(sqlmock.NewRows(userCols).
				AddRow("u1", "a@example.com", "Alice", sql.NullInt32{Int32: 5, Valid: true}, true, now, now).
				AddRow("u2", "b@example.com", sql.NullString{}, sql.NullInt32{}, true, now, now))

		users, err := repo.GetTeamMembers(10)
		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "Alice", *users[0].DisplayName)
		assert.Nil(t, users[1].DisplayName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999).
			WillReturnRows(sqlmock.NewRows(userCols))

		users, err := repo.GetTeamMembers(999)
		assert.NoError(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		users, err := repo.GetTeamMembers(10)
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamMembersByTeamIDs ───────────────────────────────────────────────

func TestTeamRepository_GetTeamMembersByTeamIDs(t *testing.T) {
	bulkCols := []string{"team_id", "id", "email", "display_name", "class_id", "is_profile_complete", "created_at", "updated_at"}
	now := time.Now()

	t.Run("success - multiple teams", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery("WHERE tm.team_id IN").WithArgs(10, 11).
			WillReturnRows(sqlmock.NewRows(bulkCols).
				AddRow(10, "u1", "a@example.com", "Alice", sql.NullInt32{}, true, now, now).
				AddRow(11, "u2", "b@example.com", "Bob", sql.NullInt32{}, true, now, now))

		result, err := repo.GetTeamMembersByTeamIDs([]int{10, 11})
		require.NoError(t, err)
		assert.Len(t, result[10], 1)
		assert.Len(t, result[11], 1)
		assert.Equal(t, "Alice", *result[10][0].DisplayName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty teamIDs returns empty map without querying DB", func(t *testing.T) {
		repo, _, close := setupTeam(t)
		defer close()

		result, err := repo.GetTeamMembersByTeamIDs([]int{})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery("WHERE tm.team_id IN").WillReturnError(errors.New("db error"))

		result, err := repo.GetTeamMembersByTeamIDs([]int{10})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── RemoveTeamMember ──────────────────────────────────────────────────────

func TestTeamRepository_RemoveTeamMember(t *testing.T) {
	const q = "DELETE FROM team_members WHERE team_id = ? AND user_id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(10, "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		assert.NoError(t, repo.RemoveTeamMember(10, "user-1"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.RemoveTeamMember(10, "user-1"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── UpdateTeamCapacity ────────────────────────────────────────────────────

func TestTeamRepository_UpdateTeamCapacity(t *testing.T) {
	const q = "UPDATE teams SET min_capacity = ?, max_capacity = ? WHERE sport_id = ? AND class_id = ?"

	t.Run("success with values", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		min, max := 3, 8
		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(&min, &max, 1, 5).
			WillReturnResult(sqlmock.NewResult(0, 1))

		assert.NoError(t, repo.UpdateTeamCapacity(1, 1, 5, &min, &max))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with nil (clears capacity)", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(nil, nil, 1, 5).
			WillReturnResult(sqlmock.NewResult(0, 1))

		assert.NoError(t, repo.UpdateTeamCapacity(1, 1, 5, nil, nil))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.UpdateTeamCapacity(1, 1, 5, nil, nil))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamCapacity ───────────────────────────────────────────────────────

func TestTeamRepository_GetTeamCapacity(t *testing.T) {
	const q = `
			SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, t.min_capacity, t.max_capacity
			FROM teams t
			JOIN classes c ON t.class_id = c.id
			WHERE c.event_id = ? AND t.sport_id = ? AND t.class_id = ?
		`
	cols := []string{"id", "name", "class_id", "sport_id", "event_id", "min_capacity", "max_capacity"}

	t.Run("success with capacity set", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1, 1, 5).
			WillReturnRows(sqlmock.NewRows(cols).
				AddRow(10, "IS3-A", 5, 1, 1, sql.NullInt64{Int64: 3, Valid: true}, sql.NullInt64{Int64: 8, Valid: true}))

		team, err := repo.GetTeamCapacity(1, 1, 5)
		require.NoError(t, err)
		require.NotNil(t, team)
		assert.Equal(t, 3, *team.MinCapacity)
		assert.Equal(t, 8, *team.MaxCapacity)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1, 1, 999).WillReturnError(sql.ErrNoRows)

		team, err := repo.GetTeamCapacity(1, 1, 999)
		assert.NoError(t, err)
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		team, err := repo.GetTeamCapacity(1, 1, 5)
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── ConfirmTeamMember ─────────────────────────────────────────────────────

func TestTeamRepository_ConfirmTeamMember(t *testing.T) {
	const q = "UPDATE team_members SET is_confirmed = true WHERE team_id = ? AND user_id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(10, "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		assert.NoError(t, repo.ConfirmTeamMember(10, "user-1"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("member not found - 0 rows affected returns ErrNoRows", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(10, "ghost-user").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.ConfirmTeamMember(10, "ghost-user")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.ConfirmTeamMember(10, "user-1"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetConfirmedTeamMembers ───────────────────────────────────────────────

func TestTeamRepository_GetConfirmedTeamMembers(t *testing.T) {
	const q = `
			SELECT u.id, u.email, u.display_name, u.class_id, u.is_profile_complete, u.created_at, u.updated_at
			FROM users u
			INNER JOIN team_members tm ON u.id = tm.user_id
			WHERE tm.team_id = ? AND tm.is_confirmed = true
			ORDER BY u.display_name, u.email
		`
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(10).
			WillReturnRows(sqlmock.NewRows(userCols).
				AddRow("u1", "a@example.com", "Alice", sql.NullInt32{Int32: 5, Valid: true}, true, now, now))

		users, err := repo.GetConfirmedTeamMembers(10)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "Alice", *users[0].DisplayName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999).
			WillReturnRows(sqlmock.NewRows(userCols))

		users, err := repo.GetConfirmedTeamMembers(999)
		assert.NoError(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		users, err := repo.GetConfirmedTeamMembers(10)
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetConfirmedTeamMembersCount ──────────────────────────────────────────

func TestTeamRepository_GetConfirmedTeamMembersCount(t *testing.T) {
	const q = "SELECT COUNT(*) FROM team_members WHERE team_id = ? AND is_confirmed = true"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		count, err := repo.GetConfirmedTeamMembersCount(10)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupTeam(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		count, err := repo.GetConfirmedTeamMembersCount(10)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
