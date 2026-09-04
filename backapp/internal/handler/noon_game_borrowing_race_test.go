package handler

import "testing"

func TestRankBorrowingRaceScoresUsesCompetitionRanking(t *testing.T) {
	scores := []borrowingRaceScoreInput{
		{EntryID: 1, CompetitionScore: 12},
		{EntryID: 2, CompetitionScore: 20},
		{EntryID: 3, CompetitionScore: 20},
		{EntryID: 4, CompetitionScore: 8},
	}

	ranked := rankBorrowingRaceScores(scores)
	wantRanks := []int{1, 1, 3, 4}
	wantEntries := []int{2, 3, 1, 4}
	for index := range ranked {
		if ranked[index].Rank != wantRanks[index] || ranked[index].EntryID != wantEntries[index] {
			t.Fatalf("ranked[%d] = entry %d rank %d, want entry %d rank %d", index, ranked[index].EntryID, ranked[index].Rank, wantEntries[index], wantRanks[index])
		}
	}
}

func TestBorrowingRaceRankPointsReadsStoredConfig(t *testing.T) {
	config := map[string]interface{}{
		"rank_points": map[string]interface{}{"1": float64(100), "3": float64(80)},
	}

	points := borrowingRaceRankPoints(config)
	if points[1] != 100 || points[3] != 80 || points[2] != 0 {
		t.Fatalf("unexpected rank points: %#v", points)
	}
}
