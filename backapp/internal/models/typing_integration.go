package models

import "time"

const TypingIntegrationAPIVersion = "v1"

type TypingEvent struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Year      int        `json:"year"`
	Season    string     `json:"season"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

type TypingSport struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TypingPlayer struct {
	ID         string  `json:"id"`
	Name       *string `json:"name,omitempty"`
	EntryOrder int     `json:"entry_order"`
}

type TypingTeam struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	ClassID   int             `json:"class_id"`
	ClassName string          `json:"class_name"`
	Players   []*TypingPlayer `json:"players"`
}

type TypingMatch struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	RoundNumber int        `json:"round_number"`
	MatchNumber int        `json:"match_number"`
	Status      string     `json:"status"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	TeamIDs     []int      `json:"team_ids"`
}

type TypingTournament struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Matches []*TypingMatch `json:"matches"`
}

type TypingActiveEventResponse struct {
	APIVersion  string         `json:"api_version"`
	GeneratedAt time.Time      `json:"generated_at"`
	Event       *TypingEvent   `json:"event"`
	Sports      []*TypingSport `json:"sports"`
}

type TypingCompetitionSnapshot struct {
	APIVersion  string              `json:"api_version"`
	GeneratedAt time.Time           `json:"generated_at"`
	Event       *TypingEvent        `json:"event"`
	Sport       *TypingSport        `json:"sport"`
	Tournaments []*TypingTournament `json:"tournaments"`
	Teams       []*TypingTeam       `json:"teams"`
	Warnings    []string            `json:"warnings"`
}
