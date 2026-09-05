package models

import "time"

type BoardGameRun struct {
	ID                  int                    `json:"id"`
	EventID             int                    `json:"event_id"`
	SportID             int                    `json:"sport_id"`
	GameType            string                 `json:"game_type"`
	Name                string                 `json:"name"`
	Description         *string                `json:"description,omitempty"`
	Location            string                 `json:"location"`
	RulesPDFURL         *string                `json:"rules_pdf_url,omitempty"`
	ScheduledDate       *string                `json:"scheduled_date,omitempty"`
	WinPoints           int                    `json:"win_points"`
	RankPoints          map[string]int         `json:"rank_points"`
	RegularMinutes      int                    `json:"regular_minutes"`
	FinalMinutes        int                    `json:"final_minutes"`
	PlayersPerClass     int                    `json:"players_per_class"`
	SubstitutesPerClass int                    `json:"substitutes_per_class"`
	Status              string                 `json:"status"`
	CreatedBy           string                 `json:"created_by"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	Tournaments         []*BoardGameTournament `json:"tournaments"`
}

type BoardGameTournament struct {
	ID       int                 `json:"id"`
	Name     string              `json:"name"`
	SlotKey  string              `json:"slot_key"`
	Entries  []*BoardGameEntry   `json:"entries"`
	Rankings []*BoardGameRanking `json:"rankings"`
}

type BoardGameEntry struct {
	ID           int                     `json:"id"`
	RunID        int                     `json:"run_id"`
	TournamentID int                     `json:"tournament_id"`
	TeamID       int                     `json:"team_id"`
	ClassID      int                     `json:"class_id"`
	ClassName    string                  `json:"class_name"`
	TeamName     string                  `json:"team_name"`
	SlotKey      string                  `json:"slot_key"`
	SeedNumber   int                     `json:"seed_number"`
	Members      []*BoardGameEntryMember `json:"members"`
}

type BoardGameEntryMember struct {
	UserID       string  `json:"user_id"`
	DisplayName  *string `json:"display_name,omitempty"`
	Email        string  `json:"email"`
	MemberOrder  int     `json:"member_order"`
	IsSubstitute bool    `json:"is_substitute"`
}

type BoardGameRanking struct {
	ID           int    `json:"id"`
	RunID        int    `json:"run_id"`
	TournamentID int    `json:"tournament_id"`
	EntryID      int    `json:"entry_id"`
	Rank         int    `json:"rank"`
	WinCount     int    `json:"win_count"`
	WinPoints    int    `json:"win_points"`
	RankPoints   int    `json:"rank_points"`
	TotalPoints  int    `json:"total_points"`
	ClassID      int    `json:"class_id"`
	ClassName    string `json:"class_name"`
	TeamName     string `json:"team_name"`
}

type BoardGameRunCreate struct {
	EventID             int
	GameType            string
	Name                string
	Description         *string
	Location            string
	RulesPDFURL         *string
	ScheduledDate       *string
	WinPoints           int
	RankPoints          map[string]int
	RegularMinutes      int
	FinalMinutes        int
	PlayersPerClass     int
	SubstitutesPerClass int
	Status              string
	CreatedBy           string
	Tournaments         []BoardGameTournamentCreate
}

type BoardGameTournamentCreate struct {
	Name    string
	SlotKey string
	Entries []BoardGameEntryCreate
	Data    TournamentData
}

type BoardGameEntryCreate struct {
	ClassID       int
	ClassName     string
	TeamName      string
	EntryKey      string
	SeedNumber    int
	MinCapacity   int
	MaxCapacity   int
	MemberIDs     []string
	SubstituteIDs []string
}

type BoardGameRankingInput struct {
	EntryID int `json:"entry_id" binding:"required"`
	Rank    int `json:"rank" binding:"required"`
}
