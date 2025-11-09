package models

import (
	"database/sql"
)

type ClassProgress struct {
	SportName      string              `json:"sport_name"`
	TeamName       string              `json:"team_name"`
	TournamentName string              `json:"tournament_name"`
	Status         string              `json:"status"`
	CurrentRound   string              `json:"current_round"`
	NextMatch      *ClassProgressMatch `json:"next_match,omitempty"`
	LastMatch      *ClassProgressMatch `json:"last_match,omitempty"`
}

type ClassProgressMatch struct {
	MatchID      int     `json:"match_id"`
	Round        int     `json:"round"`
	RoundLabel   string  `json:"round_label"`
	OpponentName string  `json:"opponent_name,omitempty"`
	MatchStatus  string  `json:"match_status,omitempty"`
	StartTime    *string `json:"start_time,omitempty"`
	Result       string  `json:"result,omitempty"`
	Score        *string `json:"score,omitempty"`
}

type MatchDetail struct {
	MatchID        int
	TournamentID   int
	TournamentName string
	SportName      string
	MaxRound       int
	Round          int
	MatchNumber    int
	Team1ID        sql.NullInt64
	Team2ID        sql.NullInt64
	Team1Score     sql.NullInt32
	Team2Score     sql.NullInt32
	WinnerTeamID   sql.NullInt64
	Status         string
	NextMatchID    sql.NullInt64
	StartTime      sql.NullString
	IsBronzeMatch  bool
	Team1Name      sql.NullString
	Team2Name      sql.NullString
}
