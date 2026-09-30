package repository

import (
	"os"
	"testing"

	"backapp/internal/models"
	"github.com/stretchr/testify/require"
)

func TestSportScoresAndMigrationMySQL(t *testing.T) {
	db := seasonalMembershipDB(t)
	_, err := db.Exec(`
        CREATE TABLE sports (id INT PRIMARY KEY, name VARCHAR(100));
        CREATE TABLE event_sports (event_id INT, sport_id INT, location VARCHAR(255), PRIMARY KEY(event_id,sport_id));
        CREATE TABLE tournaments (id INT PRIMARY KEY, event_id INT, sport_id INT, name VARCHAR(255));
        CREATE TABLE teams (id INT PRIMARY KEY, class_id INT);
        CREATE TABLE matches (
            id INT PRIMARY KEY, tournament_id INT, round INT, match_number_in_round INT,
            team1_id INT NULL, team2_id INT NULL, winner_team_id INT NULL,
            team1_score INT, team2_score INT, status VARCHAR(20),
            is_bronze_match BOOLEAN DEFAULT FALSE, is_loser_bracket_match BOOLEAN DEFAULT FALSE
        );
        CREATE TABLE board_game_runs (id INT PRIMARY KEY, event_id INT, sport_id INT);
        CREATE TABLE board_game_entries (id INT PRIMARY KEY, tournament_id INT, run_id INT, slot_key VARCHAR(64));
        CREATE TABLE score_logs (
            id INT AUTO_INCREMENT PRIMARY KEY, event_id INT, class_id INT,
            points INT, reason VARCHAR(100), source_match_id INT NULL, board_game_run_id INT NULL,
            CONSTRAINT chk_score_logs_reason CHECK (reason <> 'unsupported')
        );
        INSERT INTO sports VALUES (10,'サッカー'),(11,'キックベース'),(12,'将棋'),(13,'未開始競技');
        INSERT INTO event_sports VALUES (1,10,'other'),(1,11,'other:校庭'),(1,12,'other'),(1,13,'other'),(2,10,'other');
        INSERT INTO tournaments VALUES
            (1,1,10,'サッカー'),(2,1,11,'キックベース'),
            (3,1,12,'将棋 Aブロック'),(4,1,12,'将棋 Bブロック');
        INSERT INTO teams VALUES (1,11),(2,12);
        INSERT INTO matches (id,tournament_id,round,match_number_in_round,team1_id,team2_id,winner_team_id,team1_score,team2_score,status,is_bronze_match) VALUES
            (1,1,0,0,1,2,1,2,1,'finished',FALSE),
            (2,1,2,0,1,2,2,1,3,'finished',FALSE),
            (3,1,2,1,1,2,1,4,2,'finished',TRUE),
            (4,2,0,0,1,2,1,2,1,'finished',FALSE),
            (5,3,0,0,1,2,1,1,0,'finished',FALSE),
            (6,1,0,1,1,NULL,1,0,0,'finished',FALSE),
            (7,1,0,2,1,2,NULL,0,0,'pending',FALSE),
            (8,1,0,3,1,2,NULL,1,1,'finished',FALSE),
            (9,4,0,0,1,2,1,1,0,'finished',FALSE);
        INSERT INTO board_game_runs VALUES (1,1,12);
        INSERT INTO board_game_entries VALUES (1,3,1,'A'),(2,4,1,'B');
        INSERT INTO score_logs (event_id,class_id,points,reason,source_match_id,board_game_run_id) VALUES
            (1,11,10,'ground_win1_points',4,NULL),
            (1,11,7,'board_game_win_points',5,1),
            (1,11,80,'board_game_rank_points',5,1),
            (1,11,14,'board_game_win_points',9,1),
            (1,11,60,'board_game_rank_points',9,1),
            (1,11,5,'attendance_points',NULL,NULL);
    `)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../db/migrations/0027_add_sport_scoped_score_logs.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM score_logs WHERE source_match_id IN (6,7,8)").Scan(&count))
	require.Zero(t, count, "byes, pending matches and undecided ties must not earn points")
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM score_logs WHERE source_match_id=4").Scan(&count))
	require.Equal(t, 1, count, "existing points must not be duplicated")
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM score_logs WHERE source_match_id=5 AND reason='board_game_win_points'").Scan(&count))
	require.Equal(t, 1, count, "board games must keep their configured scoring")

	scores := map[int][]*models.ClassScore{
		1: {{ClassID: 11}, {ClassID: 12}},
		2: {{ClassID: 21}},
	}
	repo := &classRepository{db: db}
	require.NoError(t, repo.attachSportScores([]int{1, 2}, scores))
	require.Len(t, scores[1][0].SportScores, 5)
	require.Equal(t, models.SportScore{SportID: 10, SportName: "サッカー", Win1Points: 10, Win3Points: 10, ChampionPoints: 110}, scores[1][0].SportScores[0])
	require.Equal(t, models.SportScore{SportID: 11, SportName: "キックベース", Win1Points: 10}, scores[1][0].SportScores[1])
	require.Equal(t, models.SportScore{SportID: 12, SportName: "将棋", TournamentID: 3, TournamentName: "将棋 Aブロック", SlotKey: "A", IsBoardGame: true, WinPoints: 7, RankPoints: 80}, scores[1][0].SportScores[2])
	require.Equal(t, models.SportScore{SportID: 12, SportName: "将棋", TournamentID: 4, TournamentName: "将棋 Bブロック", SlotKey: "B", IsBoardGame: true, WinPoints: 14, RankPoints: 60}, scores[1][0].SportScores[3])
	require.Equal(t, models.SportScore{SportID: 13, SportName: "未開始競技"}, scores[1][0].SportScores[4])
	require.Equal(t, models.SportScore{SportID: 10, SportName: "サッカー"}, scores[2][0].SportScores[0], "events must stay isolated")

	down, err := os.ReadFile("../../db/migrations/0027_add_sport_scoped_score_logs.down.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(down))
	require.NoError(t, err)
	require.NoError(t, db.QueryRow("SELECT SUM(points) FROM score_logs").Scan(&count))
	require.Equal(t, 176, count, "rollback preserves legacy and board-game scores")
}
