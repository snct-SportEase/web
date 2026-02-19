package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"backapp/internal/models"
)

type TournamentRepository interface {
	SaveTournament(eventID int, sportID int, sportName string, tournamentData *models.TournamentData, teams []*models.Team) error
	DeleteTournamentsByEventID(eventID int) error
	DeleteTournamentsByEventAndSportID(eventID int, sportID int) error
	GetTournamentsByEventID(eventID int) ([]*models.Tournament, error)
	GetMatchesForTeam(eventID int, teamID int) ([]*models.MatchDetail, error)
	UpdateMatchStartTime(matchID int, startTime string) error
	UpdateMatchRainyModeStartTime(matchID int, rainyModeStartTime string) error
	UpdateMatchStatus(matchID int, status string) error
	UpdateMatchResult(matchID, team1Score, team2Score, winnerID int) error
	UpdateMatchResultForCorrection(matchID, team1Score, team2Score, winnerID int) error
	GetTournamentIDByMatchID(matchID int) (int, error)
	ApplyRainyModeStartTimes(eventID int) error
	IsMatchResultAlreadyEntered(matchID int) (bool, error)
}

type tournamentRepository struct {
	db *sql.DB
}

func NewTournamentRepository(db *sql.DB) TournamentRepository {
	return &tournamentRepository{db: db}
}

func (r *tournamentRepository) GetTournamentsByEventID(eventID int) ([]*models.Tournament, error) {
	// Check if rainy mode is enabled for this event
	var isRainyMode bool
	err := r.db.QueryRow("SELECT is_rainy_mode FROM events WHERE id = ?", eventID).Scan(&isRainyMode)
	if err != nil {
		// If event not found or error, default to false
		isRainyMode = false
	}

	rows, err := r.db.Query("SELECT id, name, sport_id FROM tournaments WHERE event_id = ?", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tournaments := make([]*models.Tournament, 0)
	for rows.Next() {
		var t models.Tournament
		if err := rows.Scan(&t.ID, &t.Name, &t.SportID); err != nil {
			return nil, err
		}

		matches, err := r.getMatchesByTournamentID(int64(t.ID))
		if err != nil {
			return nil, err
		}

		var bracketryMatches []models.Match
		contestants := make(map[string]models.Contestant)
		teamMap := make(map[int]*models.Team)
		contestantCounter := 0

		for _, m := range matches {
			var sides []models.Side

			if m.Team1ID.Valid {
				side, team, err := r.getSide(m.Team1ID.Int64, &contestantCounter, teamMap, contestants)
				if err != nil {
					return nil, err
				}
				if m.Team1Score.Valid {
					side.Scores = []models.Score{{MainScore: m.Team1Score.Int32}}
				}
				if m.WinnerID.Valid && m.WinnerID.Int64 == m.Team1ID.Int64 {
					side.IsWinner = true
				}
				sides = append(sides, side)
				if team != nil {
					teamMap[int(m.Team1ID.Int64)] = team
				}
			}

			if m.Team2ID.Valid {
				side, team, err := r.getSide(m.Team2ID.Int64, &contestantCounter, teamMap, contestants)
				if err != nil {
					return nil, err
				}
				if m.Team2Score.Valid {
					side.Scores = []models.Score{{MainScore: m.Team2Score.Int32}}
				}
				if m.WinnerID.Valid && m.WinnerID.Int64 == m.Team2ID.Int64 {
					side.IsWinner = true
				}
				sides = append(sides, side)
				if team != nil {
					teamMap[int(m.Team2ID.Int64)] = team
				}
			}

			var loserBracketRound *int
			if m.LoserBracketRound.Valid {
				round := int(m.LoserBracketRound.Int64)
				loserBracketRound = &round
			}
			var loserBracketBlock string
			if m.LoserBracketBlock.Valid {
				loserBracketBlock = m.LoserBracketBlock.String
			}

			// Determine which start time to use
			var effectiveStartTime string
			var effectiveStartTimeForStatus string
			if isRainyMode && m.RainyModeStartTime.Valid && m.RainyModeStartTime.String != "" {
				// Use rainy mode start time when rainy mode is enabled
				effectiveStartTime = m.RainyModeStartTime.String
				effectiveStartTimeForStatus = m.RainyModeStartTime.String
			} else if m.StartTime.Valid {
				// Use normal start time
				effectiveStartTime = m.StartTime.String
				effectiveStartTimeForStatus = m.StartTime.String
			}

			bracketryMatches = append(bracketryMatches, models.Match{
				ID:                  m.ID,
				RoundIndex:          m.Round,
				Order:               m.MatchNumberInRound,
				Sides:               sides,
				MatchStatus:         effectiveStartTimeForStatus,
				IsBronzeMatch:       m.IsBronzeMatch,
				IsLoserBracketMatch: m.IsLoserBracketMatch,
				LoserBracketRound:   loserBracketRound,
				LoserBracketBlock:   loserBracketBlock,
				StartTime:           effectiveStartTime,
				RainyModeStartTime: func() string {
					if m.RainyModeStartTime.Valid {
						return m.RainyModeStartTime.String
					}
					return ""
				}(),
			})
		}

		numRounds := 0
		for _, m := range bracketryMatches {
			if m.RoundIndex+1 > numRounds {
				numRounds = m.RoundIndex + 1
			}
		}
		rounds := make([]models.Round, numRounds)
		for i := 0; i < numRounds; i++ {
			rounds[i] = models.Round{Name: fmt.Sprintf("Round %d", i+1)}
		}

		t.Data, err = r.marshal(models.TournamentData{
			Rounds:      rounds,
			Matches:     bracketryMatches,
			Contestants: contestants,
		})
		if err != nil {
			return nil, err
		}
		tournaments = append(tournaments, &t)
	}

	return tournaments, nil
}

func (r *tournamentRepository) GetMatchesForTeam(eventID int, teamID int) ([]*models.MatchDetail, error) {
	query := `
		SELECT
			m.id,
			t.id,
			t.name,
			s.name,
			(
				SELECT MAX(round)
				FROM matches
				WHERE tournament_id = t.id
			) AS max_round,
			m.round,
			m.match_number_in_round,
			m.team1_id,
			m.team2_id,
			m.team1_score,
			m.team2_score,
			m.winner_team_id,
			m.status,
			m.next_match_id,
			m.start_time,
			m.is_bronze_match,
			team1.name,
			team2.name
		FROM matches m
		JOIN tournaments t ON m.tournament_id = t.id
		JOIN sports s ON t.sport_id = s.id
		LEFT JOIN teams team1 ON m.team1_id = team1.id
		LEFT JOIN teams team2 ON m.team2_id = team2.id
		WHERE t.event_id = ? AND (m.team1_id = ? OR m.team2_id = ?)
		ORDER BY t.id, m.round, m.match_number_in_round
	`

	rows, err := r.db.Query(query, eventID, teamID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*models.MatchDetail
	for rows.Next() {
		var detail models.MatchDetail
		if err := rows.Scan(
			&detail.MatchID,
			&detail.TournamentID,
			&detail.TournamentName,
			&detail.SportName,
			&detail.MaxRound,
			&detail.Round,
			&detail.MatchNumber,
			&detail.Team1ID,
			&detail.Team2ID,
			&detail.Team1Score,
			&detail.Team2Score,
			&detail.WinnerTeamID,
			&detail.Status,
			&detail.NextMatchID,
			&detail.StartTime,
			&detail.IsBronzeMatch,
			&detail.Team1Name,
			&detail.Team2Name,
		); err != nil {
			return nil, err
		}
		matches = append(matches, &detail)
	}

	return matches, nil
}

func (r *tournamentRepository) getSide(teamID int64, contestantCounter *int, teamMap map[int]*models.Team, contestants map[string]models.Contestant) (models.Side, *models.Team, error) {
	var team *models.Team
	var ok bool

	if team, ok = teamMap[int(teamID)]; !ok {
		var err error
		team, err = r.getTeamByID(teamID)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.Side{}, nil, nil // Team not found is not a fatal error here
			}
			return models.Side{}, nil, err
		}
	}

	if team == nil {
		return models.Side{}, nil, nil
	}

	contestantID := ""
	for id, c := range contestants {
		if len(c.Players) > 0 && c.Players[0].Title == team.Name {
			contestantID = id
			break
		}
	}

	if contestantID == "" {
		contestantID = "c" + strconv.Itoa(*contestantCounter)
		contestants[contestantID] = models.Contestant{
			Players: []models.Player{{Title: team.Name}},
		}
		(*contestantCounter)++
	}

	return models.Side{ContestantID: contestantID, TeamID: teamID}, team, nil
}

func (r *tournamentRepository) getMatchesByTournamentID(tournamentID int64) ([]*models.MatchDB, error) {
	rows, err := r.db.Query("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, team1_score, team2_score, winner_team_id, status, next_match_id, start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE tournament_id = ? ORDER BY round, match_number_in_round", tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*models.MatchDB
	for rows.Next() {
		var m models.MatchDB
		var loserBracketRound sql.NullInt64
		var loserBracketBlock sql.NullString
		if err := rows.Scan(&m.ID, &m.TournamentID, &m.Round, &m.MatchNumberInRound, &m.Team1ID, &m.Team2ID, &m.Team1Score, &m.Team2Score, &m.WinnerID, &m.Status, &m.NextMatchID, &m.StartTime, &m.IsBronzeMatch, &m.IsLoserBracketMatch, &loserBracketRound, &loserBracketBlock, &m.RainyModeStartTime); err != nil {
			return nil, err
		}
		if loserBracketRound.Valid {
			m.LoserBracketRound = loserBracketRound
		}
		if loserBracketBlock.Valid {
			m.LoserBracketBlock = loserBracketBlock
		}
		matches = append(matches, &m)
	}
	return matches, nil
}

func (r *tournamentRepository) getMatchByID(tx *sql.Tx, matchID int) (*models.MatchDB, error) {
	var m models.MatchDB
	var loserBracketRound sql.NullInt64
	var loserBracketBlock sql.NullString
	row := tx.QueryRow("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?", matchID)
	if err := row.Scan(&m.ID, &m.TournamentID, &m.Round, &m.MatchNumberInRound, &m.Team1ID, &m.Team2ID, &m.WinnerID, &m.Status, &m.NextMatchID, &m.StartTime, &m.IsBronzeMatch, &m.IsLoserBracketMatch, &loserBracketRound, &loserBracketBlock, &m.RainyModeStartTime); err != nil {
		return nil, err
	}
	if loserBracketRound.Valid {
		m.LoserBracketRound = loserBracketRound
	}
	if loserBracketBlock.Valid {
		m.LoserBracketBlock = loserBracketBlock
	}
	return &m, nil
}

func (r *tournamentRepository) IsMatchResultAlreadyEntered(matchID int) (bool, error) {
	var winnerID sql.NullInt64
	var status string
	err := r.db.QueryRow("SELECT winner_team_id, status FROM matches WHERE id = ?", matchID).Scan(&winnerID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("match not found")
		}
		return false, err
	}
	// 試合結果が既に入力済みかどうか: winner_team_idが設定されていて、statusが"finished"
	return winnerID.Valid && status == "finished", nil
}

func (r *tournamentRepository) getTeamByID(teamID int64) (*models.Team, error) {
	var team models.Team
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id 
		FROM teams t 
		JOIN classes c ON t.class_id = c.id 
		WHERE t.id = ?
	`
	row := r.db.QueryRow(query, teamID)
	if err := row.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID); err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *tournamentRepository) getTeamByIDTx(tx *sql.Tx, teamID int64) (*models.Team, error) {
	var team models.Team
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id 
		FROM teams t 
		JOIN classes c ON t.class_id = c.id 
		WHERE t.id = ?
	`
	row := tx.QueryRow(query, teamID)
	if err := row.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID); err != nil {
		return nil, err
	}
	return &team, nil
}

type locationScoreColumns struct {
	win      [3]string
	champion string
}

var locationColumns = map[string]locationScoreColumns{
	"gym1": {
		win:      [3]string{"gym1_win1_points", "gym1_win2_points", "gym1_win3_points"},
		champion: "gym1_champion_points",
	},
	"gym2": {
		win:      [3]string{"gym2_win1_points", "gym2_win2_points", "gym2_win3_points"},
		champion: "gym2_champion_points",
	},
	"ground": {
		win:      [3]string{"ground_win1_points", "ground_win2_points", "ground_win3_points"},
		champion: "ground_champion_points",
	},
}

func (r *tournamentRepository) getTournamentMetadata(tx *sql.Tx, tournamentID int) (int, int, string, error) {
	var eventID, sportID int
	var location sql.NullString

	err := tx.QueryRow(
		"SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?",
		tournamentID,
	).Scan(&eventID, &sportID, &location)
	if err != nil {
		return 0, 0, "", err
	}

	loc := ""
	if location.Valid {
		loc = location.String
	}

	return eventID, sportID, loc, nil
}

func (r *tournamentRepository) addPoints(tx *sql.Tx, eventID int, classID int, column string, points int) error {
	if column == "" || points == 0 {
		return nil
	}

	query := `
		INSERT INTO score_logs (event_id, class_id, points, reason)
		VALUES (?, ?, ?, ?)
	`

	_, err := tx.Exec(query, eventID, classID, points, column)
	return err
}

func (r *tournamentRepository) subtractPoints(tx *sql.Tx, eventID int, classID int, column string, points int) error {
	if column == "" || points == 0 {
		return nil
	}

	query := `
		INSERT INTO score_logs (event_id, class_id, points, reason)
		VALUES (?, ?, ?, ?)
	`

	_, err := tx.Exec(query, eventID, classID, -points, column)
	return err
}

func (r *tournamentRepository) applyScoring(tx *sql.Tx, match *models.MatchDB, winnerID, loserID int64, totalRounds int) error {
	if winnerID == 0 || loserID == 0 {
		return nil
	}

	eventID, _, location, err := r.getTournamentMetadata(tx, match.TournamentID)
	if err != nil {
		return err
	}

	if location == "noon_game" {
		return nil
	}

	columns, ok := locationColumns[location]
	if !ok {
		return nil
	}

	winnerTeam, err := r.getTeamByIDTx(tx, winnerID)
	if err != nil {
		return err
	}
	loserTeam, err := r.getTeamByIDTx(tx, loserID)
	if err != nil {
		return err
	}

	if winnerTeam == nil || loserTeam == nil {
		return nil
	}

	if winnerTeam.EventID != eventID || loserTeam.EventID != eventID {
		return nil
	}

	// 敗者戦二回戦の場合、勝者に10点付与（敗者戦ブロック優勝）
	if match.IsLoserBracketMatch && match.LoserBracketRound.Valid && match.LoserBracketRound.Int64 == 2 {
		if location == "gym2" {
			if err := r.addPoints(tx, eventID, winnerTeam.ClassID, "gym2_loser_bracket_champion_points", 10); err != nil {
				return err
			}
		}
		return nil
	}

	if match.Round >= 0 && match.Round < len(columns.win) {
		column := columns.win[match.Round]
		if err := r.addPoints(tx, eventID, winnerTeam.ClassID, column, 10); err != nil {
			return err
		}
	}

	if totalRounds <= 0 {
		return nil
	}

	if match.IsBronzeMatch {
		if err := r.addPoints(tx, eventID, winnerTeam.ClassID, columns.champion, 50); err != nil {
			return err
		}
		if err := r.addPoints(tx, eventID, loserTeam.ClassID, columns.champion, 40); err != nil {
			return err
		}
		return nil
	}

	if match.Round == totalRounds {
		if err := r.addPoints(tx, eventID, winnerTeam.ClassID, columns.champion, 80); err != nil {
			return err
		}
		if err := r.addPoints(tx, eventID, loserTeam.ClassID, columns.champion, 60); err != nil {
			return err
		}
	}

	return nil
}

func (r *tournamentRepository) updateRanks(tx *sql.Tx, eventID int) error {
	// class_scores is a VIEW, ranking is dynamic
	return nil
}

func (r *tournamentRepository) marshal(data interface{}) (json.RawMessage, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (r *tournamentRepository) SaveTournament(eventID int, sportID int, sportName string, tournamentData *models.TournamentData, teams []*models.Team) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// トーナメント名を生成
	// sportNameが既に完全なトーナメント名の場合はそのまま使用（" Tournament"が含まれている場合）
	// そうでない場合は "{sportName} Tournament" を生成
	tournamentName := sportName
	if !strings.Contains(sportName, " Tournament") {
		tournamentName = fmt.Sprintf("%s Tournament", sportName)
	}
	res, err := tx.Exec("INSERT INTO tournaments (name, event_id, sport_id) VALUES (?, ?, ?)", tournamentName, eventID, sportID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert tournament: %w (tournamentName: %s, eventID: %d, sportID: %d)", err, tournamentName, eventID, sportID)
	}
	tournamentID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	roundMatchIDs := make([][]int64, len(tournamentData.Rounds))
	matchMetas := map[int64]models.Match{}

	for _, match := range tournamentData.Matches {
		var team1ID, team2ID sql.NullInt64

		if len(match.Sides) > 0 && match.Sides[0].ContestantID != "" {
			contestantIndex, _ := strconv.Atoi(strings.TrimPrefix(match.Sides[0].ContestantID, "c"))
			if contestantIndex < len(teams) {
				team1ID = sql.NullInt64{Int64: int64(teams[contestantIndex].ID), Valid: true}
			}
		}
		if len(match.Sides) > 1 && match.Sides[1].ContestantID != "" {
			contestantIndex, _ := strconv.Atoi(strings.TrimPrefix(match.Sides[1].ContestantID, "c"))
			if contestantIndex < len(teams) {
				team2ID = sql.NullInt64{Int64: int64(teams[contestantIndex].ID), Valid: true}
			}
		}

		var loserBracketRound interface{}
		if match.LoserBracketRound != nil {
			loserBracketRound = *match.LoserBracketRound
		} else {
			loserBracketRound = nil
		}
		var loserBracketBlock interface{}
		if match.LoserBracketBlock != "" {
			loserBracketBlock = match.LoserBracketBlock
		} else {
			loserBracketBlock = nil
		}
		res, err := tx.Exec(
			"INSERT INTO matches (tournament_id, round, match_number_in_round, team1_id, team2_id, status, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			tournamentID,
			match.RoundIndex,
			match.Order,
			team1ID,
			team2ID,
			"pending",
			match.IsBronzeMatch,
			match.IsLoserBracketMatch,
			loserBracketRound,
			loserBracketBlock,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
		matchID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
		if match.RoundIndex < len(roundMatchIDs) {
			roundMatchIDs[match.RoundIndex] = append(roundMatchIDs[match.RoundIndex], matchID)
			matchMetas[matchID] = match
		}
	}

	// roundMatchIDsが空でない場合のみ処理
	if len(roundMatchIDs) == 0 {
		return tx.Commit()
	}

	for i := 0; i < len(roundMatchIDs)-1; i++ {
		for j, matchID := range roundMatchIDs[i] {
			if match, ok := matchMetas[matchID]; ok && match.IsBronzeMatch {
				continue
			}
			match := matchMetas[matchID]

			// 敗者戦の試合の場合は、ブロックとラウンドを考慮して次の試合を決定
			if match.IsLoserBracketMatch && match.LoserBracketRound != nil && *match.LoserBracketRound == 1 {
				// 敗者戦一回戦（ラウンド1）の勝者は、同じブロックの敗者戦二回戦（ラウンド2）に進む
				var nextMatchID int64 = 0
				if i+1 < len(roundMatchIDs) {
					for _, candidateID := range roundMatchIDs[i+1] {
						if candidateMatch, ok := matchMetas[candidateID]; ok {
							if candidateMatch.IsLoserBracketMatch && candidateMatch.LoserBracketBlock == match.LoserBracketBlock {
								// 同じブロックで、敗者戦二回戦（ラウンド2）の試合
								if candidateMatch.LoserBracketRound != nil && *candidateMatch.LoserBracketRound == 2 {
									// 同じブロックの敗者戦二回戦の試合（Orderは常に0）
									nextMatchID = candidateID
									break
								}
							}
						}
					}
				}
				if nextMatchID != 0 {
					_, err := tx.Exec("UPDATE matches SET next_match_id = ? WHERE id = ?", nextMatchID, matchID)
					if err != nil {
						tx.Rollback()
						return err
					}
				}
			} else if match.IsLoserBracketMatch {
				// 敗者戦二回戦の試合には次の試合なし（10点獲得で終了）
				// 何もしない
			} else {
				// 本戦の試合の場合は従来通り
				if i+1 < len(roundMatchIDs) && j/2 < len(roundMatchIDs[i+1]) {
					nextMatchID := roundMatchIDs[i+1][j/2]
					_, err := tx.Exec("UPDATE matches SET next_match_id = ? WHERE id = ?", nextMatchID, matchID)
					if err != nil {
						tx.Rollback()
						return err
					}
				}
			}
		}
	}

	return tx.Commit()
}

func (r *tournamentRepository) DeleteTournamentsByEventID(eventID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM tournaments WHERE event_id = ?", eventID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var tournamentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			tx.Rollback()
			return err
		}
		tournamentIDs = append(tournamentIDs, id)
	}

	if len(tournamentIDs) > 0 {
		args := make([]interface{}, len(tournamentIDs))
		for i, v := range tournamentIDs {
			args[i] = v
		}

		qMarks := strings.Repeat("?,", len(args)-1) + "?"

		query := fmt.Sprintf("DELETE FROM matches WHERE tournament_id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}

		query = fmt.Sprintf("DELETE FROM tournaments WHERE id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *tournamentRepository) DeleteTournamentsByEventAndSportID(eventID int, sportID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM tournaments WHERE event_id = ? AND sport_id = ?", eventID, sportID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var tournamentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			tx.Rollback()
			return err
		}
		tournamentIDs = append(tournamentIDs, id)
	}

	if len(tournamentIDs) > 0 {
		args := make([]interface{}, len(tournamentIDs))
		for i, v := range tournamentIDs {
			args[i] = v
		}

		qMarks := strings.Repeat("?,", len(args)-1) + "?"

		query := fmt.Sprintf("DELETE FROM matches WHERE tournament_id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}

		query = fmt.Sprintf("DELETE FROM tournaments WHERE id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *tournamentRepository) UpdateMatchStartTime(matchID int, startTime string) error {
	if startTime != "" {
		// Only change status to 'scheduled' if it is currently 'pending'.
		_, err := r.db.Exec("UPDATE matches SET start_time = ?, status = CASE WHEN status = 'pending' THEN 'scheduled' ELSE status END WHERE id = ?", startTime, matchID)
		return err
	}
	// If startTime is empty, just update the time.
	_, err := r.db.Exec("UPDATE matches SET start_time = ? WHERE id = ?", startTime, matchID)
	return err
}

func (r *tournamentRepository) UpdateMatchRainyModeStartTime(matchID int, rainyModeStartTime string) error {
	_, err := r.db.Exec("UPDATE matches SET rainy_mode_start_time = ? WHERE id = ?", rainyModeStartTime, matchID)
	return err
}

func (r *tournamentRepository) UpdateMatchStatus(matchID int, status string) error {
	_, err := r.db.Exec("UPDATE matches SET status = ? WHERE id = ?", status, matchID)
	return err
}

func (r *tournamentRepository) UpdateMatchResult(matchID, team1Score, team2Score, winnerIDInput int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	match, err := r.getMatchByID(tx, matchID)
	if err != nil {
		return err
	}

	// 雨天時モードのチェック: 昼競技とグラウンド競技をブロック
	eventID, sportID, location, err := r.getTournamentMetadata(tx, match.TournamentID)
	if err == nil {
		var isRainyMode bool
		err := tx.QueryRow("SELECT is_rainy_mode FROM events WHERE id = ?", eventID).Scan(&isRainyMode)
		if err == nil && isRainyMode {
			if location == "noon_game" || location == "ground" {
				return fmt.Errorf("雨天時モードでは、昼競技とグラウンド競技の試合結果を更新できません")
			}
		}
	} else {
		return err
	}

	alreadyFinished := match.WinnerID.Valid && match.Status == "finished"

	var winnerID, loserID int64
	if team1Score > team2Score {
		winnerID = match.Team1ID.Int64
		loserID = match.Team2ID.Int64
	} else if team2Score > team1Score {
		winnerID = match.Team2ID.Int64
		loserID = match.Team1ID.Int64
	} else {
		winnerID = int64(winnerIDInput)
		if winnerID == match.Team1ID.Int64 {
			loserID = match.Team2ID.Int64
		} else {
			loserID = match.Team1ID.Int64
		}
	}

	_, err = tx.Exec("UPDATE matches SET team1_score = ?, team2_score = ?, winner_team_id = ?, status = 'finished' WHERE id = ?", team1Score, team2Score, winnerID, matchID)
	if err != nil {
		return err
	}

	// Advance winner to the next match
	if match.NextMatchID.Valid {
		nextMatch, err := r.getMatchByID(tx, int(match.NextMatchID.Int64))
		if err != nil {
			return err
		}

		if nextMatch.Team1ID.Valid {
			_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", winnerID, nextMatch.ID)
		} else {
			_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", winnerID, nextMatch.ID)
		}
		if err != nil {
			return err
		}
	}

	var maxRound sql.NullInt64
	if err := tx.QueryRow("SELECT MAX(round) FROM matches WHERE tournament_id = ?", match.TournamentID).Scan(&maxRound); err != nil {
		return err
	}
	totalRounds := 0
	if maxRound.Valid {
		totalRounds = int(maxRound.Int64)
	}

	// Handle bronze match for semi-final losers
	if !match.IsBronzeMatch && totalRounds > 0 {
		// Semi-finals are in the second to last round
		if match.Round == totalRounds-1 {
			var bronzeMatchID int64
			err := tx.QueryRow("SELECT id FROM matches WHERE tournament_id = ? AND is_bronze_match = TRUE", match.TournamentID).Scan(&bronzeMatchID)
			if err != nil {
				if err == sql.ErrNoRows {
					// No bronze match, so we do nothing
				} else {
					return err
				}
			} else {
				bronzeMatch, err := r.getMatchByID(tx, int(bronzeMatchID))
				if err != nil {
					return err
				}
				if bronzeMatch.Team1ID.Valid {
					_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", loserID, bronzeMatch.ID)
				} else {
					_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", loserID, bronzeMatch.ID)
				}
				if err != nil {
					return err
				}
			}
		}
	}

	// Handle loser bracket tournament assignment for first round losers (gym2 only)
	if !match.IsBronzeMatch && !match.IsLoserBracketMatch && match.Round == 0 && location == "gym2" {
		// Get loser bracket tournaments (A and B blocks)
		var loserBracketAID, loserBracketBID int64
		errA := tx.QueryRow("SELECT id FROM tournaments WHERE event_id = ? AND sport_id = ? AND name LIKE ?", eventID, sportID, "%敗者戦Aブロック%").Scan(&loserBracketAID)
		errB := tx.QueryRow("SELECT id FROM tournaments WHERE event_id = ? AND sport_id = ? AND name LIKE ?", eventID, sportID, "%敗者戦Bブロック%").Scan(&loserBracketBID)

		if errA == nil && errB == nil {
			// Determine which loser bracket match to assign based on match order
			// Aブロック: 本戦1-4試合の敗者
			// Bブロック: 本戦5-8試合の敗者
			matchOrder := match.MatchNumberInRound

			var targetTournamentID int64
			var targetMatchOrder int
			var isTeam1 bool

			if matchOrder < 4 {
				// Aブロック
				targetTournamentID = loserBracketAID
				targetMatchOrder = matchOrder / 2 // 0 or 1
				isTeam1 = (matchOrder % 2) == 0   // true for orders 0,2; false for 1,3
			} else if matchOrder < 8 {
				// Bブロック
				targetTournamentID = loserBracketBID
				targetMatchOrder = (matchOrder - 4) / 2 // 0 or 1
				isTeam1 = ((matchOrder - 4) % 2) == 0   // true for orders 4,6; false for 5,7
			}

			if targetTournamentID > 0 {
				// Find the target match in the loser bracket tournament
				// Loser bracket round 1 matches have loser_bracket_round = 1
				var targetMatchID int64
				loserRound1 := 1
				err := tx.QueryRow(
					"SELECT id FROM matches WHERE tournament_id = ? AND round = ? AND match_number_in_round = ? AND is_loser_bracket_match = TRUE AND loser_bracket_round = ?",
					targetTournamentID, 0, targetMatchOrder, loserRound1,
				).Scan(&targetMatchID)

				if err == nil {
					if isTeam1 {
						// Set team1_id
						_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", loserID, targetMatchID)
					} else {
						// Set team2_id
						_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", loserID, targetMatchID)
					}
					if err != nil {
						return err
					}
				} else if err != sql.ErrNoRows {
					return err
				}
			}
		}
	}

	if !alreadyFinished {
		if err := r.applyScoring(tx, match, winnerID, loserID, totalRounds); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateMatchResultForCorrection updates an already entered match result and corrects the next match teams
func (r *tournamentRepository) UpdateMatchResultForCorrection(matchID, team1Score, team2Score, winnerIDInput int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	match, err := r.getMatchByID(tx, matchID)
	if err != nil {
		return err
	}

	// 既に入力済みでない場合はエラー
	if !match.WinnerID.Valid || match.Status != "finished" {
		return fmt.Errorf("試合結果がまだ入力されていません。通常の更新メソッドを使用してください")
	}

	// 雨天時モードのチェック: 昼競技とグラウンド競技をブロック
	eventID, sportID, location, err := r.getTournamentMetadata(tx, match.TournamentID)
	if err == nil {
		var isRainyMode bool
		err := tx.QueryRow("SELECT is_rainy_mode FROM events WHERE id = ?", eventID).Scan(&isRainyMode)
		if err == nil && isRainyMode {
			if location == "noon_game" || location == "ground" {
				return fmt.Errorf("雨天時モードでは、昼競技とグラウンド競技の試合結果を更新できません")
			}
		}
	} else {
		return err
	}

	// 前回の勝者を取得
	previousWinnerID := match.WinnerID.Int64

	// 前回の敗者を取得
	previousLoserID := match.Team1ID.Int64
	if previousLoserID == previousWinnerID {
		previousLoserID = match.Team2ID.Int64
	}

	// 新しい勝者を決定
	var newWinnerID, loserID int64
	if team1Score > team2Score {
		newWinnerID = match.Team1ID.Int64
		loserID = match.Team2ID.Int64
	} else if team2Score > team1Score {
		newWinnerID = match.Team2ID.Int64
		loserID = match.Team1ID.Int64
	} else {
		newWinnerID = int64(winnerIDInput)
		if newWinnerID == match.Team1ID.Int64 {
			loserID = match.Team2ID.Int64
		} else {
			loserID = match.Team1ID.Int64
		}
	}

	// 勝者が変わらない場合は通常の更新メソッドと同じ処理
	if previousWinnerID == newWinnerID {
		_, err = tx.Exec("UPDATE matches SET team1_score = ?, team2_score = ?, winner_team_id = ?, status = 'finished' WHERE id = ?", team1Score, team2Score, newWinnerID, matchID)
		if err != nil {
			return err
		}
		return tx.Commit()
	}

	// 前回の勝者と敗者のチーム情報を取得
	previousWinnerTeam, err := r.getTeamByIDTx(tx, previousWinnerID)
	if err != nil {
		return err
	}
	previousLoserTeam, err := r.getTeamByIDTx(tx, previousLoserID)
	if err != nil {
		return err
	}

	// 前回の勝者に付与された点数をリセット
	if previousWinnerTeam != nil && previousLoserTeam != nil {
		// 前回の点数を計算して減算
		if err := r.revertScoring(tx, match, previousWinnerID, previousLoserID, eventID, location); err != nil {
			return err
		}
	}

	// 試合結果を更新
	_, err = tx.Exec("UPDATE matches SET team1_score = ?, team2_score = ?, winner_team_id = ?, status = 'finished' WHERE id = ?", team1Score, team2Score, newWinnerID, matchID)
	if err != nil {
		return err
	}

	// 次の試合から前の勝者を削除し、新しい勝者を設定
	if match.NextMatchID.Valid {
		nextMatch, err := r.getMatchByID(tx, int(match.NextMatchID.Int64))
		if err != nil {
			return err
		}

		// 次の試合が既に終了している場合、その試合の結果を無効化
		if nextMatch.WinnerID.Valid && nextMatch.Status == "finished" {
			// 次の試合の勝者と敗者を取得
			nextMatchWinnerID := nextMatch.WinnerID.Int64
			nextMatchLoserID := nextMatch.Team1ID.Int64
			if nextMatchLoserID == nextMatchWinnerID {
				nextMatchLoserID = nextMatch.Team2ID.Int64
			}

			// 次の試合の勝者に付与された点数をリセット
			if nextMatchWinnerID != 0 && nextMatchLoserID != 0 {
				if err := r.revertScoring(tx, nextMatch, nextMatchWinnerID, nextMatchLoserID, eventID, location); err != nil {
					return err
				}
			}

			// 次の試合の結果を無効化
			_, err = tx.Exec("UPDATE matches SET team1_score = NULL, team2_score = NULL, winner_team_id = NULL, status = 'pending' WHERE id = ?", nextMatch.ID)
			if err != nil {
				return err
			}

			// さらにその次の試合も連鎖的に無効化（再帰的に処理）
			if nextMatch.NextMatchID.Valid {
				if err := r.invalidateSubsequentMatches(tx, int(nextMatch.NextMatchID.Int64), eventID, location); err != nil {
					return err
				}
			}
		}

		// 前の勝者を次の試合から削除
		if nextMatch.Team1ID.Valid && nextMatch.Team1ID.Int64 == previousWinnerID {
			_, err = tx.Exec("UPDATE matches SET team1_id = NULL WHERE id = ?", nextMatch.ID)
			if err != nil {
				return err
			}
		} else if nextMatch.Team2ID.Valid && nextMatch.Team2ID.Int64 == previousWinnerID {
			_, err = tx.Exec("UPDATE matches SET team2_id = NULL WHERE id = ?", nextMatch.ID)
			if err != nil {
				return err
			}
		}

		// 新しい勝者を次の試合に設定
		if !nextMatch.Team1ID.Valid {
			_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", newWinnerID, nextMatch.ID)
		} else if !nextMatch.Team2ID.Valid {
			_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", newWinnerID, nextMatch.ID)
		} else {
			// 両方のチームが既に設定されている場合は、前の勝者の位置に新しい勝者を設定
			if nextMatch.Team1ID.Int64 == previousWinnerID {
				_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", newWinnerID, nextMatch.ID)
			} else if nextMatch.Team2ID.Int64 == previousWinnerID {
				_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", newWinnerID, nextMatch.ID)
			} else {
				// 前の勝者が次の試合にいない場合（既に削除されているなど）、空いている位置に設定
				if !nextMatch.Team1ID.Valid {
					_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", newWinnerID, nextMatch.ID)
				} else {
					_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", newWinnerID, nextMatch.ID)
				}
			}
		}
		if err != nil {
			return err
		}
	}

	// 敗者戦への進出も更新する必要がある場合
	var maxRound sql.NullInt64
	if err := tx.QueryRow("SELECT MAX(round) FROM matches WHERE tournament_id = ?", match.TournamentID).Scan(&maxRound); err != nil {
		return err
	}
	totalRounds := 0
	if maxRound.Valid {
		totalRounds = int(maxRound.Int64)
	}

	// Handle bronze match for semi-final losers
	if !match.IsBronzeMatch && totalRounds > 0 {
		// Semi-finals are in the second to last round
		if match.Round == totalRounds-1 {
			var bronzeMatchID int64
			err := tx.QueryRow("SELECT id FROM matches WHERE tournament_id = ? AND is_bronze_match = TRUE", match.TournamentID).Scan(&bronzeMatchID)
			if err == nil {
				// 前の敗者を敗者戦から削除
				bronzeMatch, err := r.getMatchByID(tx, int(bronzeMatchID))
				if err == nil {
					if bronzeMatch.Team1ID.Valid && bronzeMatch.Team1ID.Int64 == previousLoserID {
						_, err = tx.Exec("UPDATE matches SET team1_id = NULL WHERE id = ?", bronzeMatchID)
					} else if bronzeMatch.Team2ID.Valid && bronzeMatch.Team2ID.Int64 == previousLoserID {
						_, err = tx.Exec("UPDATE matches SET team2_id = NULL WHERE id = ?", bronzeMatchID)
					}
					if err != nil {
						return err
					}

					// 新しい敗者を敗者戦に設定
					if !bronzeMatch.Team1ID.Valid {
						_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", loserID, bronzeMatchID)
					} else if !bronzeMatch.Team2ID.Valid {
						_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", loserID, bronzeMatchID)
					} else {
						// 前の敗者の位置に新しい敗者を設定
						if bronzeMatch.Team1ID.Int64 == previousLoserID {
							_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", loserID, bronzeMatchID)
						} else if bronzeMatch.Team2ID.Int64 == previousLoserID {
							_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", loserID, bronzeMatchID)
						}
					}
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Handle loser bracket tournament assignment for first round losers (gym2 only)
	if !match.IsBronzeMatch && !match.IsLoserBracketMatch && match.Round == 0 && location == "gym2" {
		// Get loser bracket tournaments (A and B blocks)
		var loserBracketAID, loserBracketBID int64
		errA := tx.QueryRow("SELECT id FROM tournaments WHERE event_id = ? AND sport_id = ? AND name LIKE ?", eventID, sportID, "%敗者戦Aブロック%").Scan(&loserBracketAID)
		errB := tx.QueryRow("SELECT id FROM tournaments WHERE event_id = ? AND sport_id = ? AND name LIKE ?", eventID, sportID, "%敗者戦Bブロック%").Scan(&loserBracketBID)

		if errA == nil && errB == nil {
			// Determine which loser bracket match to assign based on match order
			matchOrder := match.MatchNumberInRound

			var targetTournamentID int64
			var targetMatchOrder int
			var isTeam1 bool

			if matchOrder < 4 {
				// Aブロック
				targetTournamentID = loserBracketAID
				targetMatchOrder = matchOrder / 2 // 0 or 1
				isTeam1 = (matchOrder % 2) == 0   // true for orders 0,2; false for 1,3
			} else if matchOrder < 8 {
				// Bブロック
				targetTournamentID = loserBracketBID
				targetMatchOrder = (matchOrder - 4) / 2 // 0 or 1
				isTeam1 = ((matchOrder - 4) % 2) == 0   // true for orders 4,6; false for 5,7
			}

			if targetTournamentID > 0 {
				// Find the target match in the loser bracket tournament
				var targetMatchID int64
				loserRound1 := 1
				err := tx.QueryRow(
					"SELECT id FROM matches WHERE tournament_id = ? AND round = ? AND match_number_in_round = ? AND is_loser_bracket_match = TRUE AND loser_bracket_round = ?",
					targetTournamentID, 0, targetMatchOrder, loserRound1,
				).Scan(&targetMatchID)

				if err == nil {
					// 前の敗者を敗者戦から削除
					targetMatch, err := r.getMatchByID(tx, int(targetMatchID))
					if err == nil {
						if isTeam1 {
							if targetMatch.Team1ID.Valid && targetMatch.Team1ID.Int64 == previousLoserID {
								_, err = tx.Exec("UPDATE matches SET team1_id = NULL WHERE id = ?", targetMatchID)
							}
						} else {
							if targetMatch.Team2ID.Valid && targetMatch.Team2ID.Int64 == previousLoserID {
								_, err = tx.Exec("UPDATE matches SET team2_id = NULL WHERE id = ?", targetMatchID)
							}
						}
						if err != nil {
							return err
						}

						// 新しい敗者を敗者戦に設定
						if isTeam1 {
							_, err = tx.Exec("UPDATE matches SET team1_id = ? WHERE id = ?", loserID, targetMatchID)
						} else {
							_, err = tx.Exec("UPDATE matches SET team2_id = ? WHERE id = ?", loserID, targetMatchID)
						}
						if err != nil {
							return err
						}
					}
				} else if err != sql.ErrNoRows {
					return err
				}
			}
		}
	}

	// 新しい勝者に点数を付与（totalRoundsは既に取得済み）
	if err := r.applyScoring(tx, match, newWinnerID, loserID, totalRounds); err != nil {
		return err
	}

	return tx.Commit()
}

// revertScoring reverts the points that were awarded to the previous winner and loser
func (r *tournamentRepository) revertScoring(tx *sql.Tx, match *models.MatchDB, previousWinnerID, previousLoserID int64, eventID int, location string) error {
	if previousWinnerID == 0 || previousLoserID == 0 {
		return nil
	}

	if location == "noon_game" {
		return nil
	}

	columns, ok := locationColumns[location]
	if !ok {
		return nil
	}

	previousWinnerTeam, err := r.getTeamByIDTx(tx, previousWinnerID)
	if err != nil {
		return err
	}
	previousLoserTeam, err := r.getTeamByIDTx(tx, previousLoserID)
	if err != nil {
		return err
	}

	if previousWinnerTeam == nil || previousLoserTeam == nil {
		return nil
	}

	if previousWinnerTeam.EventID != eventID || previousLoserTeam.EventID != eventID {
		return nil
	}

	// 敗者戦二回戦の場合、前回の勝者から10点減算
	if match.IsLoserBracketMatch && match.LoserBracketRound.Valid && match.LoserBracketRound.Int64 == 2 {
		if location == "gym2" {
			if err := r.subtractPoints(tx, eventID, previousWinnerTeam.ClassID, "gym2_loser_bracket_champion_points", 10); err != nil {
				return err
			}
			return r.updateRanks(tx, eventID)
		}
	}

	// ラウンド勝利の点数を減算
	var maxRound sql.NullInt64
	if err := tx.QueryRow("SELECT MAX(round) FROM matches WHERE tournament_id = ?", match.TournamentID).Scan(&maxRound); err != nil {
		return err
	}
	totalRounds := 0
	if maxRound.Valid {
		totalRounds = int(maxRound.Int64)
	}

	if match.Round >= 0 && match.Round < len(columns.win) {
		column := columns.win[match.Round]
		if err := r.subtractPoints(tx, eventID, previousWinnerTeam.ClassID, column, 10); err != nil {
			return err
		}
	}

	if totalRounds <= 0 {
		return r.updateRanks(tx, eventID)
	}

	// 3位決定戦の場合
	if match.IsBronzeMatch {
		if err := r.subtractPoints(tx, eventID, previousWinnerTeam.ClassID, columns.champion, 50); err != nil {
			return err
		}
		if err := r.subtractPoints(tx, eventID, previousLoserTeam.ClassID, columns.champion, 40); err != nil {
			return err
		}
		return r.updateRanks(tx, eventID)
	}

	// 決勝の場合
	if match.Round == totalRounds {
		if err := r.subtractPoints(tx, eventID, previousWinnerTeam.ClassID, columns.champion, 80); err != nil {
			return err
		}
		if err := r.subtractPoints(tx, eventID, previousLoserTeam.ClassID, columns.champion, 60); err != nil {
			return err
		}
	}

	return r.updateRanks(tx, eventID)
}

// invalidateSubsequentMatches invalidates all subsequent matches that depend on the given match
// This is used when a match result is corrected and subsequent matches need to be invalidated
func (r *tournamentRepository) invalidateSubsequentMatches(tx *sql.Tx, matchID int, eventID int, location string) error {
	match, err := r.getMatchByID(tx, matchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Match not found, nothing to invalidate
		}
		return err
	}

	// 既に終了していない場合は何もしない
	if !match.WinnerID.Valid || match.Status != "finished" {
		return nil
	}

	// この試合の勝者と敗者を取得
	winnerID := match.WinnerID.Int64
	loserID := match.Team1ID.Int64
	if loserID == winnerID {
		loserID = match.Team2ID.Int64
	}

	// この試合の勝者に付与された点数をリセット
	if winnerID != 0 && loserID != 0 {
		if err := r.revertScoring(tx, match, winnerID, loserID, eventID, location); err != nil {
			return err
		}
	}

	// この試合の結果を無効化
	_, err = tx.Exec("UPDATE matches SET team1_score = NULL, team2_score = NULL, winner_team_id = NULL, status = 'pending' WHERE id = ?", matchID)
	if err != nil {
		return err
	}

	// さらにその次の試合も連鎖的に無効化（再帰的に処理）
	if match.NextMatchID.Valid {
		if err := r.invalidateSubsequentMatches(tx, int(match.NextMatchID.Int64), eventID, location); err != nil {
			return err
		}
	}

	return nil
}

func (r *tournamentRepository) GetTournamentIDByMatchID(matchID int) (int, error) {
	var tournamentID int
	err := r.db.QueryRow("SELECT tournament_id FROM matches WHERE id = ?", matchID).Scan(&tournamentID)
	return tournamentID, err
}

// ApplyRainyModeStartTimes applies rainy_mode_start_time to start_time for all matches in the event's tournaments
// This is called when rainy mode is enabled
func (r *tournamentRepository) ApplyRainyModeStartTimes(eventID int) error {
	// Update all matches that have rainy_mode_start_time set
	// Set start_time = rainy_mode_start_time where rainy_mode_start_time is not null
	query := `
		UPDATE matches m
		INNER JOIN tournaments t ON m.tournament_id = t.id
		SET m.start_time = m.rainy_mode_start_time,
		    m.status = CASE WHEN m.status = 'pending' THEN 'scheduled' ELSE m.status END
		WHERE t.event_id = ?
		  AND m.rainy_mode_start_time IS NOT NULL
		  AND m.rainy_mode_start_time != ''
	`
	_, err := r.db.Exec(query, eventID)
	return err
}
