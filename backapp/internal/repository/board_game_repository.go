package repository

import (
	"backapp/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrBoardGameRunHasResults = errors.New("board game run has results")

type BoardGameRepository interface {
	CreateRun(input *models.BoardGameRunCreate) (*models.BoardGameRun, error)
	GetRunByID(runID int) (*models.BoardGameRun, error)
	ListRuns(eventID int, publishedOnly bool) ([]*models.BoardGameRun, error)
	SaveRankings(runID, tournamentID int, rankings []models.BoardGameRankingInput, recordedBy string) (*models.BoardGameRun, error)
}

type boardGameRepository struct{ db *sql.DB }

func NewBoardGameRepository(db *sql.DB) BoardGameRepository { return &boardGameRepository{db: db} }

func (r *boardGameRepository) CreateRun(input *models.BoardGameRunCreate) (*models.BoardGameRun, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var existingRunID, sportID int
	err = tx.QueryRow("SELECT id, sport_id FROM board_game_runs WHERE event_id = ? AND game_type = ?", input.EventID, input.GameType).Scan(&existingRunID, &sportID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		var finished int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM matches m JOIN tournaments t ON t.id=m.tournament_id WHERE t.event_id=? AND t.sport_id=? AND m.status='finished'`, input.EventID, sportID).Scan(&finished); err != nil {
			return nil, err
		}
		if finished > 0 {
			return nil, ErrBoardGameRunHasResults
		}
		if err := deleteBoardGameRunData(tx, existingRunID, input.EventID, sportID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec("UPDATE sports SET name=? WHERE id=?", input.Name, sportID); err != nil {
			return nil, err
		}
	} else {
		err = tx.QueryRow(`SELECT s.id FROM sports s
			LEFT JOIN event_sports es ON es.event_id=? AND es.sport_id=s.id
			WHERE s.name=? AND es.sport_id IS NULL
			ORDER BY s.id LIMIT 1`, input.EventID, input.Name).Scan(&sportID)
		if err == sql.ErrNoRows {
			result, insertErr := tx.Exec("INSERT INTO sports (name) VALUES (?)", input.Name)
			if insertErr != nil {
				return nil, insertErr
			}
			id, idErr := result.LastInsertId()
			if idErr != nil {
				return nil, idErr
			}
			sportID = int(id)
		} else if err != nil {
			return nil, err
		}
	}

	templateKey := "board_game_tournament"
	if _, err := tx.Exec(`INSERT INTO event_sports (event_id,sport_id,description,rules_pdf_url,location,template_key,min_capacity,max_capacity)
		VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE description=VALUES(description),rules_pdf_url=VALUES(rules_pdf_url),location=VALUES(location),template_key=VALUES(template_key),min_capacity=VALUES(min_capacity),max_capacity=VALUES(max_capacity)`,
		input.EventID, sportID, input.Description, input.RulesPDFURL, "other", templateKey, input.PlayersPerClass, input.PlayersPerClass+input.SubstitutesPerClass); err != nil {
		return nil, err
	}
	rankJSON, err := json.Marshal(input.RankPoints)
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(`INSERT INTO board_game_runs (event_id,sport_id,game_type,name,description,location,rules_pdf_url,scheduled_date,win_points,rank_points,regular_minutes,final_minutes,players_per_class,substitutes_per_class,status,created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, input.EventID, sportID, input.GameType, input.Name, input.Description, input.Location, input.RulesPDFURL, input.ScheduledDate, input.WinPoints, rankJSON, input.RegularMinutes, input.FinalMinutes, input.PlayersPerClass, input.SubstitutesPerClass, input.Status, input.CreatedBy)
	if err != nil {
		return nil, err
	}
	runID64, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	runID := int(runID64)

	for _, tournament := range input.Tournaments {
		result, err := tx.Exec("INSERT INTO tournaments (name,event_id,sport_id) VALUES (?,?,?)", tournament.Name, input.EventID, sportID)
		if err != nil {
			return nil, err
		}
		tournamentID64, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		tournamentID := int(tournamentID64)
		teamIDs := make([]int, 0, len(tournament.Entries))
		for _, entry := range tournament.Entries {
			result, err := tx.Exec("INSERT INTO teams (name,class_id,sport_id,entry_key,min_capacity,max_capacity) VALUES (?,?,?,?,?,?)", entry.TeamName, entry.ClassID, sportID, entry.EntryKey, entry.MinCapacity, entry.MaxCapacity)
			if err != nil {
				return nil, err
			}
			teamID64, err := result.LastInsertId()
			if err != nil {
				return nil, err
			}
			teamID := int(teamID64)
			teamIDs = append(teamIDs, teamID)
			result, err = tx.Exec("INSERT INTO board_game_entries (run_id,tournament_id,team_id,class_id,slot_key,seed_number) VALUES (?,?,?,?,?,?)", runID, tournamentID, teamID, entry.ClassID, tournament.SlotKey, entry.SeedNumber)
			if err != nil {
				return nil, err
			}
			entryID64, err := result.LastInsertId()
			if err != nil {
				return nil, err
			}
			entryID := int(entryID64)
			memberOrder := 0
			for _, memberID := range entry.MemberIDs {
				if err := insertBoardGameMember(tx, entryID, teamID, entry.ClassID, memberID, memberOrder, false); err != nil {
					return nil, err
				}
				memberOrder++
			}
			for _, memberID := range entry.SubstituteIDs {
				if err := insertBoardGameMember(tx, entryID, teamID, entry.ClassID, memberID, memberOrder, true); err != nil {
					return nil, err
				}
				memberOrder++
			}
		}
		if err := insertBoardGameMatches(tx, tournamentID, &tournament.Data, teamIDs); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetRunByID(runID)
}

func deleteBoardGameRunData(tx *sql.Tx, runID, eventID, sportID int) error {
	rows, err := tx.Query("SELECT team_id FROM board_game_entries WHERE run_id=?", runID)
	if err != nil {
		return err
	}
	var teamIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		teamIDs = append(teamIDs, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM score_logs WHERE board_game_run_id=?", runID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE m FROM matches m JOIN tournaments t ON t.id=m.tournament_id WHERE t.event_id=? AND t.sport_id=?", eventID, sportID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM tournaments WHERE event_id=? AND sport_id=?", eventID, sportID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM board_game_runs WHERE id=?", runID); err != nil {
		return err
	}
	for _, teamID := range teamIDs {
		if _, err := tx.Exec("DELETE FROM teams WHERE id=?", teamID); err != nil {
			return err
		}
	}
	return nil
}

func insertBoardGameMember(tx *sql.Tx, entryID, teamID, classID int, userID string, order int, substitute bool) error {
	var actualClassID sql.NullInt64
	if err := tx.QueryRow("SELECT class_id FROM users WHERE id=?", userID).Scan(&actualClassID); err != nil {
		return err
	}
	if !actualClassID.Valid || int(actualClassID.Int64) != classID {
		return fmt.Errorf("player does not belong to entry class")
	}
	if _, err := tx.Exec("INSERT INTO board_game_entry_members (entry_id,user_id,member_order,is_substitute) VALUES (?,?,?,?)", entryID, userID, order, substitute); err != nil {
		return err
	}
	_, err := tx.Exec("INSERT IGNORE INTO team_members (team_id,user_id,is_confirmed) VALUES (?,?,TRUE)", teamID, userID)
	return err
}

func insertBoardGameMatches(tx *sql.Tx, tournamentID int, data *models.TournamentData, teamIDs []int) error {
	type matchRef struct {
		id    int64
		match models.Match
	}
	refs := make([]matchRef, 0, len(data.Matches))
	for _, match := range data.Matches {
		var team1, team2 interface{}
		if len(match.Sides) > 0 && match.Sides[0].ContestantID != "" {
			index, err := strconv.Atoi(strings.TrimPrefix(match.Sides[0].ContestantID, "c"))
			if err != nil || index < 0 || index >= len(teamIDs) {
				return fmt.Errorf("invalid contestant")
			}
			team1 = teamIDs[index]
		}
		if len(match.Sides) > 1 && match.Sides[1].ContestantID != "" {
			index, err := strconv.Atoi(strings.TrimPrefix(match.Sides[1].ContestantID, "c"))
			if err != nil || index < 0 || index >= len(teamIDs) {
				return fmt.Errorf("invalid contestant")
			}
			team2 = teamIDs[index]
		}
		status := "pending"
		if match.StartTime != "" {
			status = "scheduled"
		}
		result, err := tx.Exec(`INSERT INTO matches (tournament_id,round,match_number_in_round,team1_id,team2_id,status,match_start_time,is_bronze_match) VALUES (?,?,?,?,?,?,?,?)`, tournamentID, match.RoundIndex, match.Order, team1, team2, status, boardGameNullableString(match.StartTime), match.IsBronzeMatch)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		refs = append(refs, matchRef{id, match})
	}
	byRound := map[int][]matchRef{}
	for _, ref := range refs {
		if !ref.match.IsBronzeMatch {
			byRound[ref.match.RoundIndex] = append(byRound[ref.match.RoundIndex], ref)
		}
	}
	for round := range byRound {
		sort.Slice(byRound[round], func(i, j int) bool { return byRound[round][i].match.Order < byRound[round][j].match.Order })
	}
	for round := 0; round < len(data.Rounds)-1; round++ {
		for index, ref := range byRound[round] {
			next := byRound[round+1]
			if index/2 < len(next) {
				if _, err := tx.Exec("UPDATE matches SET next_match_id=? WHERE id=?", next[index/2].id, ref.id); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func boardGameNullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func (r *boardGameRepository) GetRunByID(runID int) (*models.BoardGameRun, error) {
	rows, err := r.db.Query(`SELECT r.id,r.event_id,r.sport_id,r.game_type,r.name,r.description,r.location,r.rules_pdf_url,r.scheduled_date,r.win_points,r.rank_points,r.regular_minutes,r.final_minutes,r.players_per_class,r.substitutes_per_class,r.status,r.created_by,r.created_at,r.updated_at
		FROM board_game_runs r WHERE r.id=?`, runID)
	if err != nil {
		return nil, err
	}
	runs, err := r.readRuns(rows)
	if err != nil || len(runs) == 0 {
		return nil, err
	}
	return runs[0], nil
}

func (r *boardGameRepository) ListRuns(eventID int, publishedOnly bool) ([]*models.BoardGameRun, error) {
	var rows *sql.Rows
	var err error
	if publishedOnly {
		rows, err = r.db.Query(`SELECT r.id,r.event_id,r.sport_id,r.game_type,r.name,r.description,r.location,r.rules_pdf_url,r.scheduled_date,r.win_points,r.rank_points,r.regular_minutes,r.final_minutes,r.players_per_class,r.substitutes_per_class,r.status,r.created_by,r.created_at,r.updated_at
			FROM board_game_runs r WHERE r.event_id=? AND r.status IN ('published','completed') ORDER BY r.id`, eventID)
	} else {
		rows, err = r.db.Query(`SELECT r.id,r.event_id,r.sport_id,r.game_type,r.name,r.description,r.location,r.rules_pdf_url,r.scheduled_date,r.win_points,r.rank_points,r.regular_minutes,r.final_minutes,r.players_per_class,r.substitutes_per_class,r.status,r.created_by,r.created_at,r.updated_at
			FROM board_game_runs r WHERE r.event_id=? ORDER BY r.id`, eventID)
	}
	if err != nil {
		return nil, err
	}
	return r.readRuns(rows)
}

func (r *boardGameRepository) readRuns(rows *sql.Rows) ([]*models.BoardGameRun, error) {
	var runs []*models.BoardGameRun
	for rows.Next() {
		run := &models.BoardGameRun{RankPoints: map[string]int{}, Tournaments: []*models.BoardGameTournament{}}
		var desc, pdf sql.NullString
		var date sql.NullTime
		var rankJSON []byte
		if err := rows.Scan(&run.ID, &run.EventID, &run.SportID, &run.GameType, &run.Name, &desc, &run.Location, &pdf, &date, &run.WinPoints, &rankJSON, &run.RegularMinutes, &run.FinalMinutes, &run.PlayersPerClass, &run.SubstitutesPerClass, &run.Status, &run.CreatedBy, &run.CreatedAt, &run.UpdatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			run.Description = &desc.String
		}
		if pdf.Valid {
			run.RulesPDFURL = &pdf.String
		}
		if date.Valid {
			value := date.Time.Format("2006-01-02")
			run.ScheduledDate = &value
		}
		if err := json.Unmarshal(rankJSON, &run.RankPoints); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for _, run := range runs {
		if err := r.loadRunChildren(run); err != nil {
			return nil, err
		}
	}
	return runs, nil
}

func (r *boardGameRepository) loadRunChildren(run *models.BoardGameRun) error {
	rows, err := r.db.Query(`SELECT DISTINCT t.id,t.name,e.slot_key FROM tournaments t JOIN board_game_entries e ON e.tournament_id=t.id WHERE e.run_id=? ORDER BY t.id`, run.ID)
	if err != nil {
		return err
	}
	tournaments := make([]*models.BoardGameTournament, 0)
	for rows.Next() {
		tournament := &models.BoardGameTournament{Entries: []*models.BoardGameEntry{}, Rankings: []*models.BoardGameRanking{}}
		if err := rows.Scan(&tournament.ID, &tournament.Name, &tournament.SlotKey); err != nil {
			return err
		}
		tournaments = append(tournaments, tournament)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, tournament := range tournaments {
		if err := r.loadTournamentChildren(tournament, run.ID); err != nil {
			return err
		}
		run.Tournaments = append(run.Tournaments, tournament)
	}
	return nil
}

func (r *boardGameRepository) loadTournamentChildren(tournament *models.BoardGameTournament, runID int) error {
	rows, err := r.db.Query(`SELECT e.id,e.run_id,e.tournament_id,e.team_id,e.class_id,c.name,t.name,e.slot_key,e.seed_number FROM board_game_entries e JOIN classes c ON c.id=e.class_id JOIN teams t ON t.id=e.team_id WHERE e.run_id=? AND e.tournament_id=? ORDER BY e.seed_number`, runID, tournament.ID)
	if err != nil {
		return err
	}
	entries := make([]*models.BoardGameEntry, 0)
	for rows.Next() {
		e := &models.BoardGameEntry{Members: []*models.BoardGameEntryMember{}}
		if err := rows.Scan(&e.ID, &e.RunID, &e.TournamentID, &e.TeamID, &e.ClassID, &e.ClassName, &e.TeamName, &e.SlotKey, &e.SeedNumber); err != nil {
			rows.Close()
			return err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, e := range entries {
		memberRows, err := r.db.Query(`SELECT m.user_id,u.display_name,u.email,m.member_order,m.is_substitute FROM board_game_entry_members m JOIN users u ON u.id=m.user_id WHERE m.entry_id=? ORDER BY m.member_order`, e.ID)
		if err != nil {
			return err
		}
		for memberRows.Next() {
			m := &models.BoardGameEntryMember{}
			if err := memberRows.Scan(&m.UserID, &m.DisplayName, &m.Email, &m.MemberOrder, &m.IsSubstitute); err != nil {
				memberRows.Close()
				return err
			}
			e.Members = append(e.Members, m)
		}
		if err := memberRows.Err(); err != nil {
			memberRows.Close()
			return err
		}
		if err := memberRows.Close(); err != nil {
			return err
		}
		tournament.Entries = append(tournament.Entries, e)
	}
	rankRows, err := r.db.Query(`SELECT r.id,r.run_id,r.tournament_id,r.entry_id,r.rank_number,r.win_count,r.win_points,r.rank_points,r.total_points,e.class_id,c.name,t.name FROM board_game_rankings r JOIN board_game_entries e ON e.id=r.entry_id JOIN classes c ON c.id=e.class_id JOIN teams t ON t.id=e.team_id WHERE r.run_id=? AND r.tournament_id=? ORDER BY r.rank_number`, runID, tournament.ID)
	if err != nil {
		return err
	}
	defer rankRows.Close()
	for rankRows.Next() {
		ranking := &models.BoardGameRanking{}
		if err := rankRows.Scan(&ranking.ID, &ranking.RunID, &ranking.TournamentID, &ranking.EntryID, &ranking.Rank, &ranking.WinCount, &ranking.WinPoints, &ranking.RankPoints, &ranking.TotalPoints, &ranking.ClassID, &ranking.ClassName, &ranking.TeamName); err != nil {
			return err
		}
		tournament.Rankings = append(tournament.Rankings, ranking)
	}
	return rankRows.Err()
}

func (r *boardGameRepository) SaveRankings(runID, tournamentID int, inputs []models.BoardGameRankingInput, recordedBy string) (*models.BoardGameRun, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var winPoints int
	var rankJSON []byte
	var eventID int
	if err := tx.QueryRow(`SELECT r.win_points,r.rank_points,r.event_id FROM board_game_runs r JOIN tournaments t ON t.event_id=r.event_id AND t.sport_id=r.sport_id WHERE r.id=? AND t.id=?`, runID, tournamentID).Scan(&winPoints, &rankJSON, &eventID); err != nil {
		return nil, err
	}
	rankPoints := map[string]int{}
	if err := json.Unmarshal(rankJSON, &rankPoints); err != nil {
		return nil, err
	}
	var entryCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM board_game_entries WHERE run_id=? AND tournament_id=?", runID, tournamentID).Scan(&entryCount); err != nil {
		return nil, err
	}
	required := 4
	if entryCount < required {
		required = entryCount
	}
	if len(inputs) != required {
		return nil, fmt.Errorf("top %d rankings are required", required)
	}
	var unfinished int
	if err := tx.QueryRow("SELECT COUNT(*) FROM matches WHERE tournament_id=? AND status<>'finished'", tournamentID).Scan(&unfinished); err != nil {
		return nil, err
	}
	if unfinished > 0 {
		return nil, fmt.Errorf("all matches must be finished before saving rankings")
	}
	seenEntries, seenRanks := map[int]bool{}, map[int]bool{}
	var finalMatchID sql.NullInt64
	if err := tx.QueryRow("SELECT id FROM matches WHERE tournament_id=? AND is_bronze_match=FALSE ORDER BY round DESC,match_number_in_round LIMIT 1", tournamentID).Scan(&finalMatchID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE sl FROM score_logs sl JOIN matches m ON m.id=sl.source_match_id WHERE sl.board_game_run_id=? AND sl.reason='board_game_rank_points' AND m.tournament_id=?`, runID, tournamentID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec("DELETE FROM board_game_rankings WHERE run_id=? AND tournament_id=?", runID, tournamentID); err != nil {
		return nil, err
	}
	for _, input := range inputs {
		if input.Rank < 1 || input.Rank > required || seenEntries[input.EntryID] || seenRanks[input.Rank] {
			return nil, fmt.Errorf("invalid ranking")
		}
		seenEntries[input.EntryID] = true
		seenRanks[input.Rank] = true
		var classID int
		if err := tx.QueryRow("SELECT class_id FROM board_game_entries WHERE id=? AND run_id=? AND tournament_id=?", input.EntryID, runID, tournamentID).Scan(&classID); err != nil {
			return nil, err
		}
		var wins int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM score_logs sl JOIN matches m ON m.id=sl.source_match_id WHERE sl.board_game_run_id=? AND sl.reason='board_game_win_points' AND sl.points>0 AND sl.class_id=? AND m.tournament_id=?`, runID, classID, tournamentID).Scan(&wins); err != nil {
			return nil, err
		}
		winScore, rankScore := wins*winPoints, rankPoints[strconv.Itoa(input.Rank)]
		total := winScore + rankScore
		if _, err := tx.Exec(`INSERT INTO board_game_rankings (run_id,tournament_id,entry_id,rank_number,win_count,win_points,rank_points,total_points,recorded_by) VALUES (?,?,?,?,?,?,?,?,?)`, runID, tournamentID, input.EntryID, input.Rank, wins, winScore, rankScore, total, recordedBy); err != nil {
			return nil, err
		}
		if rankScore != 0 {
			if _, err := tx.Exec(`INSERT INTO score_logs (event_id,class_id,points,reason,source_match_id,board_game_run_id) VALUES (?,?,?,?,?,?)`, eventID, classID, rankScore, "board_game_rank_points", finalMatchID.Int64, runID); err != nil {
				return nil, err
			}
		}
	}
	var tournaments, ranked int
	if err := tx.QueryRow("SELECT COUNT(DISTINCT tournament_id) FROM board_game_entries WHERE run_id=?", runID).Scan(&tournaments); err != nil {
		return nil, err
	}
	if err := tx.QueryRow("SELECT COUNT(DISTINCT tournament_id) FROM board_game_rankings WHERE run_id=?", runID).Scan(&ranked); err != nil {
		return nil, err
	}
	status := "published"
	if tournaments > 0 && ranked == tournaments {
		status = "completed"
	}
	if _, err := tx.Exec("UPDATE board_game_runs SET status=? WHERE id=?", status, runID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetRunByID(runID)
}
