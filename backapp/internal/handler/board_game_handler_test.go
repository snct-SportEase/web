package handler

import (
	"backapp/internal/models"
	"testing"
)

func TestBoardGamePresetFor(t *testing.T) {
	shogi, ok := boardGamePresetFor("shogi")
	if !ok {
		t.Fatal("shogi preset was not found")
	}
	if shogi.WinPoints != 5 || shogi.RankPoints["1"] != 40 || shogi.RegularMinutes != 15 || shogi.FinalMinutes != 30 {
		t.Fatalf("unexpected shogi preset: %+v", shogi)
	}
	if len(shogi.Slots) != 2 || shogi.Slots[0] != "A" || shogi.Slots[1] != "B" {
		t.Fatalf("unexpected shogi slots: %v", shogi.Slots)
	}

	othello, ok := boardGamePresetFor("othello")
	if !ok {
		t.Fatal("othello preset was not found")
	}
	if othello.WinPoints != 10 || othello.RankPoints["4"] != 20 || othello.RegularMinutes != 10 || othello.FinalMinutes != 20 {
		t.Fatalf("unexpected othello preset: %+v", othello)
	}
}

func TestBuildBoardGameBracketCreatesProgressionAndBronzeMatch(t *testing.T) {
	date := "2026-09-10"
	entries := make([]models.BoardGameEntryCreate, 16)
	for i := range entries {
		entries[i] = models.BoardGameEntryCreate{TeamName: "team"}
	}
	data := buildBoardGameBracket(entries, &date)
	if len(data.Rounds) != 4 {
		t.Fatalf("round count = %d, want 4", len(data.Rounds))
	}
	if len(data.Matches) != 16 {
		t.Fatalf("match count = %d, want 16 including bronze", len(data.Matches))
	}
	if data.Rounds[2].Name != "準決勝" || data.Rounds[3].Name != "決勝" {
		t.Fatalf("unexpected final round names: %+v", data.Rounds)
	}

	times := map[string]bool{}
	bronze := 0
	for _, match := range data.Matches {
		times[match.StartTime] = true
		if match.IsBronzeMatch {
			bronze++
			if match.RoundIndex != 3 || match.Order != 1 {
				t.Fatalf("unexpected bronze match: %+v", match)
			}
		}
	}
	for _, want := range []string{"2026-09-10T09:45:00+09:00", "2026-09-10T10:45:00+09:00", "2026-09-10T13:00:00+09:00", "2026-09-10T14:00:00+09:00", "2026-09-10T15:00:00+09:00"} {
		if !times[want] {
			t.Errorf("default start time %s was not generated", want)
		}
	}
	if bronze != 1 {
		t.Fatalf("bronze matches = %d, want 1", bronze)
	}
}

func TestValidateBoardGameRoster(t *testing.T) {
	if err := validateBoardGameRoster("shogi", boardGameParticipantRequest{PlayerIDs: []string{"a", "b"}, SubstituteIDs: []string{"c"}}); err != nil {
		t.Fatalf("valid shogi roster was rejected: %v", err)
	}
	if err := validateBoardGameRoster("shogi", boardGameParticipantRequest{PlayerIDs: []string{"a"}}); err == nil {
		t.Fatal("invalid shogi roster was accepted")
	}
	if err := validateBoardGameRoster("othello", boardGameParticipantRequest{PlayerIDs: []string{"a", "b", "c"}}); err != nil {
		t.Fatalf("valid othello roster was rejected: %v", err)
	}
	if err := validateBoardGameRoster("othello", boardGameParticipantRequest{PlayerIDs: []string{"a", "a"}}); err == nil {
		t.Fatal("duplicate player was accepted")
	}
}

func TestBoardGameSeedOrderValidatesPermutation(t *testing.T) {
	participants := []boardGameParticipantRequest{{ClassID: 1}, {ClassID: 2}, {ClassID: 3}, {ClassID: 4}}
	order, err := boardGameSeedOrder("A", participants, map[string][]int{"A": {4, 2, 1, 3}})
	if err != nil {
		t.Fatalf("valid order was rejected: %v", err)
	}
	if order[0] != 4 || order[3] != 3 {
		t.Fatalf("unexpected order: %v", order)
	}
	if _, err := boardGameSeedOrder("A", participants, map[string][]int{"A": {1, 1, 3, 4}}); err == nil {
		t.Fatal("duplicate seed was accepted")
	}
}
