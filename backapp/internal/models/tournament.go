package models

import (
	"database/sql"
	"encoding/json"
)

// Tournament represents a tournament entity in the database

type Tournament struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	EventID int             `json:"event_id"`
	SportID int             `json:"sport_id"`
	Data    json.RawMessage `json:"data"`
}

// MatchDB represents the structure of a match in the database

type MatchDB struct {
	ID                 int
	Round              int
	MatchNumberInRound int
	Team1ID            sql.NullInt64
	Team2ID            sql.NullInt64
	WinnerID           sql.NullInt64
	Status             string
	NextMatchID        sql.NullInt64
	StartTime          sql.NullString
}

// Player represents a player in a contestant

type Player struct {
	Title       string `json:"title"`
	Nationality string `json:"nationality,omitempty"`
}

// Contestant represents a contestant in the tournament

type Contestant struct {
	EntryStatus string   `json:"entryStatus,omitempty"`
	Players     []Player `json:"players"`
}

// Score represents a score for a side

type Score struct {
	MainScore interface{} `json:"mainScore"`
	Subscore  interface{} `json:"subscore,omitempty"`
	IsWinner  bool        `json:"isWinner,omitempty"`
}

// Side represents a side in a match

type Side struct {
	Title        string      `json:"title,omitempty"`
	ContestantID string      `json:"contestantId,omitempty"`
	Scores       []Score     `json:"scores,omitempty"`
	CurrentScore interface{} `json:"currentScore,omitempty"`
	IsServing    bool        `json:"isServing,omitempty"`
	IsWinner     bool        `json:"isWinner,omitempty"`
}

// Match represents a match in the tournament

type Match struct {
	ID            int    `json:"id,omitempty"`
	RoundIndex    int    `json:"roundIndex"`
	Order         int    `json:"order"`
	Sides         []Side `json:"sides,omitempty"`
	MatchStatus   string `json:"matchStatus,omitempty"`
	StartTime     string `json:"startTime,omitempty"`
	IsLive        bool   `json:"isLive,omitempty"`
	IsBronzeMatch bool   `json:"isBronzeMatch,omitempty"`
}

// Round represents a round in the tournament

type Round struct {
	Name string `json:"name,omitempty"`
}

// TournamentData represents the entire tournament structure

type TournamentData struct {
	Rounds      []Round               `json:"rounds"`
	Matches     []Match               `json:"matches,omitempty"`
	Contestants map[string]Contestant `json:"contestants,omitempty"`
}

// Team represents a team for tournament generation

type TeamName struct {
	Name string `json:"name"`
}

type GeneratedTournament struct {
	EventID        int            `json:"event_id"`
	SportID        int            `json:"sport_id"`
	SportName      string         `json:"sport_name"`
	TournamentData TournamentData `json:"tournament_data"`
	ShuffledTeams  []Team         `json:"shuffled_teams"`
}
