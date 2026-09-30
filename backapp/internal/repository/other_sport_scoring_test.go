package repository

import (
	"database/sql"
	"regexp"
	"testing"

	"backapp/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestOtherSportScoringAndCorrection(t *testing.T) {
	for _, location := range []string{"other", "other:サッカー場"} {
		for _, tc := range []struct {
			name    string
			round   int
			bronze  bool
			reasons []string
			points  []int
			classes []int
		}{
			{"1勝点", 0, false, []string{"tournament_win1_points"}, []int{10}, []int{11}},
			{"2勝点", 1, false, []string{"tournament_win2_points"}, []int{10}, []int{11}},
			{"決勝", 2, false, []string{"tournament_win3_points", "tournament_champion_points", "tournament_champion_points"}, []int{10, 80, 60}, []int{11, 11, 12}},
			{"3位決定戦", 2, true, []string{"tournament_win3_points", "tournament_champion_points", "tournament_champion_points"}, []int{10, 50, 40}, []int{11, 11, 12}},
		} {
			t.Run(location+"/"+tc.name, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()
				repo := &tournamentRepository{db: db}
				match := &models.MatchDB{ID: 100, TournamentID: 1, Round: tc.round, IsBronzeMatch: tc.bronze}
				mock.ExpectBegin()
				tx, err := db.Begin()
				require.NoError(t, err)
				mock.ExpectQuery("SELECT t.event_id, t.sport_id, es.location").WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 10, location))
				expectConfig := func() {
					mock.ExpectQuery("SELECT DISTINCT r.id,r.win_points").WithArgs(1).WillReturnError(sql.ErrNoRows)
				}
				expectTeams := func() {
					for id := 1; id <= 2; id++ {
						mock.ExpectQuery("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id").WithArgs(id).
							WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(id, "team", id+10, 10, 1))
					}
				}
				expectLogs := func(sign int) {
					for i, reason := range tc.reasons {
						mock.ExpectExec(regexp.QuoteMeta("INSERT INTO score_logs (event_id, class_id, points, reason, source_match_id, sport_id) VALUES (?, ?, ?, ?, ?, (SELECT t.sport_id FROM matches m JOIN tournaments t ON t.id = m.tournament_id WHERE m.id = ?))")).
							WithArgs(1, tc.classes[i], sign*tc.points[i], reason, 100, 100).WillReturnResult(sqlmock.NewResult(1, 1))
					}
				}
				expectConfig()
				expectTeams()
				expectLogs(1)
				require.NoError(t, repo.applyScoring(tx, match, 1, 2, 2))
				expectConfig()
				expectTeams()
				mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(round) FROM matches WHERE tournament_id = ?")).WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))
				expectLogs(-1)
				require.NoError(t, repo.revertScoring(tx, match, 1, 2, 1, location))
				mock.ExpectRollback()
				require.NoError(t, tx.Rollback())
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}
