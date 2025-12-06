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
	GetTournamentIDByMatchID(matchID int) (int, error)
	ApplyRainyModeStartTimes(eventID int) error
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

func (r *tournamentRepository) getTeamByID(teamID int64) (*models.Team, error) {
	var team models.Team
	row := r.db.QueryRow("SELECT id, name, class_id, sport_id, event_id FROM teams WHERE id = ?", teamID)
	if err := row.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID); err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *tournamentRepository) getTeamByIDTx(tx *sql.Tx, teamID int64) (*models.Team, error) {
	var team models.Team
	row := tx.QueryRow("SELECT id, name, class_id, sport_id, event_id FROM teams WHERE id = ?", teamID)
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

func (r *tournamentRepository) getTournamentMetadata(tx *sql.Tx, tournamentID int) (int, string, error) {
	var eventID, sportID int
	var location sql.NullString

	err := tx.QueryRow(
		"SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?",
		tournamentID,
	).Scan(&eventID, &sportID, &location)
	if err != nil {
		return 0, "", err
	}

	loc := ""
	if location.Valid {
		loc = location.String
	}

	return eventID, loc, nil
}

func (r *tournamentRepository) addPoints(tx *sql.Tx, eventID int, classID int, column string, points int) error {
	if column == "" || points == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		INSERT INTO class_scores (event_id, class_id, %[1]s)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE %[1]s = %[1]s + VALUES(%[1]s)
	`, column)

	_, err := tx.Exec(query, eventID, classID, points)
	return err
}

func (r *tournamentRepository) applyScoring(tx *sql.Tx, match *models.MatchDB, winnerID, loserID int64, totalRounds int) error {
	if winnerID == 0 || loserID == 0 {
		return nil
	}

	eventID, location, err := r.getTournamentMetadata(tx, match.TournamentID)
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

	pointsAwarded := false

	// 敗者戦二回戦の場合、勝者に10点付与（敗者戦ブロック優勝）
	if match.IsLoserBracketMatch && match.LoserBracketRound.Valid && match.LoserBracketRound.Int64 == 2 {
		// gym2の場合、gym2_loser_bracket_champion_pointsに10点追加
		if location == "gym2" {
			if err := r.addPoints(tx, eventID, winnerTeam.ClassID, "gym2_loser_bracket_champion_points", 10); err != nil {
				return err
			}
			pointsAwarded = true
		}
		// 敗者戦のスコアリング後は通常の処理をスキップ
		if pointsAwarded {
			return r.updateRanks(tx, eventID)
		}
	}

	if match.Round >= 0 && match.Round < len(columns.win) {
		column := columns.win[match.Round]
		if err := r.addPoints(tx, eventID, winnerTeam.ClassID, column, 10); err != nil {
			return err
		}
		pointsAwarded = true
	}

	if totalRounds <= 0 {
		return nil
	}

	if match.IsBronzeMatch {
		if err := r.addPoints(tx, eventID, winnerTeam.ClassID, columns.champion, 50); err != nil {
			return err
		}
		pointsAwarded = true
		if err := r.addPoints(tx, eventID, loserTeam.ClassID, columns.champion, 40); err != nil {
			return err
		}

		return r.updateRanks(tx, eventID)
	}

	if match.Round == totalRounds {
		if err := r.addPoints(tx, eventID, winnerTeam.ClassID, columns.champion, 80); err != nil {
			return err
		}
		pointsAwarded = true
		if err := r.addPoints(tx, eventID, loserTeam.ClassID, columns.champion, 60); err != nil {
			return err
		}
		pointsAwarded = true
	}

	if pointsAwarded {
		return r.updateRanks(tx, eventID)
	}

	return nil
}

func (r *tournamentRepository) updateRanks(tx *sql.Tx, eventID int) error {
	var season string
	if err := tx.QueryRow("SELECT season FROM events WHERE id = ?", eventID).Scan(&season); err != nil {
		return err
	}

	if err := updateCurrentEventRanksTx(tx, eventID); err != nil {
		return err
	}

	if season == "autumn" {
		if err := updateOverallRanksTx(tx, eventID); err != nil {
			return err
		}
	}

	return nil
}

func updateCurrentEventRanksTx(tx *sql.Tx, eventID int) error {
	// Check if all classes have 0 points (competition not started)
	var maxPoints int
	err := tx.QueryRow(`
		SELECT COALESCE(MAX(total_points_current_event), 0)
		FROM class_scores
		WHERE event_id = ?
	`, eventID).Scan(&maxPoints)
	if err != nil {
		return fmt.Errorf("failed to check max points: %w", err)
	}

	if maxPoints == 0 {
		// All classes have 0 points, set all ranks to 0
		_, err = tx.Exec(`
			UPDATE class_scores
			SET rank_current_event = 0
			WHERE event_id = ?
		`, eventID)
		if err != nil {
			return fmt.Errorf("failed to reset current event ranks: %w", err)
		}
		return nil
	}

	// Normal ranking
	const query = `
		UPDATE class_scores cs
		JOIN (
			SELECT
				class_id,
				RANK() OVER (ORDER BY total_points_current_event DESC) AS new_rank
			FROM class_scores
			WHERE event_id = ?
		) ranked_data ON cs.class_id = ranked_data.class_id
		SET cs.rank_current_event = ranked_data.new_rank
		WHERE cs.event_id = ?
	`

	if _, err := tx.Exec(query, eventID, eventID); err != nil {
		return fmt.Errorf("failed to update current event ranks: %w", err)
	}

	return nil
}

func updateOverallRanksTx(tx *sql.Tx, eventID int) error {
	// Check if all classes have 0 points (competition not started)
	var maxPoints int
	err := tx.QueryRow(`
		SELECT COALESCE(MAX(total_points_overall), 0)
		FROM class_scores
		WHERE event_id = ?
	`, eventID).Scan(&maxPoints)
	if err != nil {
		return fmt.Errorf("failed to check max points: %w", err)
	}

	if maxPoints == 0 {
		// All classes have 0 points, set all ranks to 0
		_, err = tx.Exec(`
			UPDATE class_scores
			SET rank_overall = 0
			WHERE event_id = ?
		`, eventID)
		if err != nil {
			return fmt.Errorf("failed to reset overall ranks: %w", err)
		}
		return nil
	}

	// Normal ranking
	const query = `
		UPDATE class_scores cs
		JOIN (
			SELECT
				class_id,
				RANK() OVER (ORDER BY total_points_overall DESC) AS new_rank
			FROM class_scores
			WHERE event_id = ?
		) ranked_data ON cs.class_id = ranked_data.class_id
		SET cs.rank_overall = ranked_data.new_rank
		WHERE cs.event_id = ?
	`

	if _, err := tx.Exec(query, eventID, eventID); err != nil {
		return fmt.Errorf("failed to update overall ranks: %w", err)
	}

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
	eventID, location, err := r.getTournamentMetadata(tx, match.TournamentID)
	if err == nil {
		var isRainyMode bool
		err := tx.QueryRow("SELECT is_rainy_mode FROM events WHERE id = ?", eventID).Scan(&isRainyMode)
		if err == nil && isRainyMode {
			if location == "noon_game" || location == "ground" {
				return fmt.Errorf("雨天時モードでは、昼競技とグラウンド競技の試合結果を更新できません")
			}
		}
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

	if !alreadyFinished {
		if err := r.applyScoring(tx, match, winnerID, loserID, totalRounds); err != nil {
			return err
		}
	}

	return tx.Commit()
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
