package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"backapp/internal/websocket"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAllTournamentsPreview_LoserBracketBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Generate tournament with loser bracket A and B blocks for gym2", func(t *testing.T) {
		mockTournRepo := new(MockTournamentRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		hubManager := websocket.NewHubManager()

		h := handler.NewTournamentHandler(mockTournRepo, mockSportRepo, mockTeamRepo, mockClassRepo, mockEventRepo, hubManager)

		eventID := 1
		sportID := 1

		// Mock sport with gym2 location
		mockSportRepo.On("GetSportsByEventID", eventID).Return([]*models.EventSport{
			{
				EventID:  eventID,
				SportID:  sportID,
				Location: "gym2",
			},
		}, nil).Once()

		sport := &models.Sport{
			ID:   sportID,
			Name: "Basketball",
		}
		mockSportRepo.On("GetSportByID", sportID).Return(sport, nil).Once()

		// Create 16 teams for 8 first round matches
		teams := make([]*models.Team, 16)
		for i := 0; i < 16; i++ {
			teams[i] = &models.Team{
				ID:      i + 1,
				Name:    "Team " + string(rune('A'+i)),
				ClassID: i + 1,
				SportID: sportID,
				EventID: eventID,
			}
		}
		mockSportRepo.On("GetTeamsBySportID", sportID).Return(teams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		h.GenerateAllTournamentsPreviewHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var tournaments []models.GeneratedTournament
		err := json.Unmarshal(w.Body.Bytes(), &tournaments)
		assert.NoError(t, err)

		// Should have 3 tournaments: main tournament, loser bracket A block, loser bracket B block
		assert.Len(t, tournaments, 3, "Should have 3 tournaments: main, loser bracket A, loser bracket B")

		// Find main tournament and loser bracket tournaments
		var mainTournament *models.GeneratedTournament
		var loserBracketA *models.GeneratedTournament
		var loserBracketB *models.GeneratedTournament

		for i := range tournaments {
			switch tournaments[i].SportName {
			case "Basketball":
				mainTournament = &tournaments[i]
			case "Basketball Tournament - 敗者戦Aブロック":
				loserBracketA = &tournaments[i]
			case "Basketball Tournament - 敗者戦Bブロック":
				loserBracketB = &tournaments[i]
			}
		}

		// Verify main tournament exists and has no loser bracket matches
		assert.NotNil(t, mainTournament, "Main tournament should exist")
		assert.Equal(t, eventID, mainTournament.EventID)
		assert.Equal(t, sportID, mainTournament.SportID)
		assert.Equal(t, "Basketball", mainTournament.SportName)

		// Check that main tournament has no loser bracket matches
		mainLoserMatches := []models.Match{}
		for _, match := range mainTournament.TournamentData.Matches {
			if match.IsLoserBracketMatch {
				mainLoserMatches = append(mainLoserMatches, match)
			}
		}
		assert.Equal(t, 0, len(mainLoserMatches), "Main tournament should not have loser bracket matches")

		// Verify loser bracket A tournament
		assert.NotNil(t, loserBracketA, "Loser bracket A tournament should exist")
		assert.Equal(t, eventID, loserBracketA.EventID)
		assert.Equal(t, sportID, loserBracketA.SportID)
		assert.Equal(t, "Basketball Tournament - 敗者戦Aブロック", loserBracketA.SportName)

		// Check loser bracket A rounds
		loserARounds := loserBracketA.TournamentData.Rounds
		assert.Equal(t, 2, len(loserARounds), "Loser bracket A should have 2 rounds")
		assert.Equal(t, "一回戦", loserARounds[0].Name)
		assert.Equal(t, "二回戦", loserARounds[1].Name)

		// Check loser bracket A matches (3 matches: 2 in round 1, 1 in round 2)
		loserAMatches := loserBracketA.TournamentData.Matches
		assert.Equal(t, 3, len(loserAMatches), "Loser bracket A should have 3 matches")

		loserARound1Matches := []models.Match{}
		loserARound2Matches := []models.Match{}
		for _, match := range loserAMatches {
			assert.True(t, match.IsLoserBracketMatch, "All matches should be loser bracket matches")
			assert.Equal(t, "A", match.LoserBracketBlock, "All matches should be A block")
			if match.RoundIndex == 0 {
				loserARound1Matches = append(loserARound1Matches, match)
			} else if match.RoundIndex == 1 {
				loserARound2Matches = append(loserARound2Matches, match)
			}
		}
		assert.Equal(t, 2, len(loserARound1Matches), "Loser bracket A round 1 should have 2 matches")
		assert.Equal(t, 1, len(loserARound2Matches), "Loser bracket A round 2 should have 1 match")
		assert.Equal(t, 0, loserARound2Matches[0].Order, "Loser bracket A round 2 match should have order 0")

		// Verify loser bracket B tournament
		assert.NotNil(t, loserBracketB, "Loser bracket B tournament should exist")
		assert.Equal(t, eventID, loserBracketB.EventID)
		assert.Equal(t, sportID, loserBracketB.SportID)
		assert.Equal(t, "Basketball Tournament - 敗者戦Bブロック", loserBracketB.SportName)

		// Check loser bracket B rounds
		loserBRounds := loserBracketB.TournamentData.Rounds
		assert.Equal(t, 2, len(loserBRounds), "Loser bracket B should have 2 rounds")
		assert.Equal(t, "一回戦", loserBRounds[0].Name)
		assert.Equal(t, "二回戦", loserBRounds[1].Name)

		// Check loser bracket B matches (3 matches: 2 in round 1, 1 in round 2)
		loserBMatches := loserBracketB.TournamentData.Matches
		assert.Equal(t, 3, len(loserBMatches), "Loser bracket B should have 3 matches")

		loserBRound1Matches := []models.Match{}
		loserBRound2Matches := []models.Match{}
		for _, match := range loserBMatches {
			assert.True(t, match.IsLoserBracketMatch, "All matches should be loser bracket matches")
			assert.Equal(t, "B", match.LoserBracketBlock, "All matches should be B block")
			if match.RoundIndex == 0 {
				loserBRound1Matches = append(loserBRound1Matches, match)
			} else if match.RoundIndex == 1 {
				loserBRound2Matches = append(loserBRound2Matches, match)
			}
		}
		assert.Equal(t, 2, len(loserBRound1Matches), "Loser bracket B round 1 should have 2 matches")
		assert.Equal(t, 1, len(loserBRound2Matches), "Loser bracket B round 2 should have 1 match")
		assert.Equal(t, 0, loserBRound2Matches[0].Order, "Loser bracket B round 2 match should have order 0")

		mockSportRepo.AssertExpectations(t)
	})

	t.Run("No loser bracket for non-gym2 sports", func(t *testing.T) {
		mockTournRepo := new(MockTournamentRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		hubManager := websocket.NewHubManager()

		h := handler.NewTournamentHandler(mockTournRepo, mockSportRepo, mockTeamRepo, mockClassRepo, mockEventRepo, hubManager)

		eventID := 1
		sportID := 1

		// Mock sport with gym1 location (not gym2)
		mockSportRepo.On("GetSportsByEventID", eventID).Return([]*models.EventSport{
			{
				EventID:  eventID,
				SportID:  sportID,
				Location: "gym1",
			},
		}, nil).Once()

		sport := &models.Sport{
			ID:   sportID,
			Name: "Volleyball",
		}
		mockSportRepo.On("GetSportByID", sportID).Return(sport, nil).Once()

		// Create 16 teams
		teams := make([]*models.Team, 16)
		for i := 0; i < 16; i++ {
			teams[i] = &models.Team{
				ID:      i + 1,
				Name:    "Team " + string(rune('A'+i)),
				ClassID: i + 1,
				SportID: sportID,
				EventID: eventID,
			}
		}
		mockSportRepo.On("GetTeamsBySportID", sportID).Return(teams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		h.GenerateAllTournamentsPreviewHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var tournaments []models.GeneratedTournament
		err := json.Unmarshal(w.Body.Bytes(), &tournaments)
		assert.NoError(t, err)
		assert.Len(t, tournaments, 1)

		tournament := tournaments[0]
		matches := tournament.TournamentData.Matches

		// Check that there are no loser bracket matches for gym1
		loserMatches := []models.Match{}
		for _, match := range matches {
			if match.IsLoserBracketMatch {
				loserMatches = append(loserMatches, match)
			}
		}

		assert.Equal(t, 0, len(loserMatches), "Non-gym2 sports should not have loser bracket matches")

		mockSportRepo.AssertExpectations(t)
	})
}

func TestGenerateAllTournamentsPreview_LoserBracketFirstRoundMatchMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Verify that first round match losers map to correct loser bracket", func(t *testing.T) {
		mockTournRepo := new(MockTournamentRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		hubManager := websocket.NewHubManager()

		h := handler.NewTournamentHandler(mockTournRepo, mockSportRepo, mockTeamRepo, mockClassRepo, mockEventRepo, hubManager)

		eventID := 1
		sportID := 1

		// Mock sport with gym2 location
		mockSportRepo.On("GetSportsByEventID", eventID).Return([]*models.EventSport{
			{
				EventID:  eventID,
				SportID:  sportID,
				Location: "gym2",
			},
		}, nil).Once()

		sport := &models.Sport{
			ID:   sportID,
			Name: "Basketball",
		}
		mockSportRepo.On("GetSportByID", sportID).Return(sport, nil).Once()

		// Create 16 teams for 8 first round matches
		teams := make([]*models.Team, 16)
		for i := 0; i < 16; i++ {
			teams[i] = &models.Team{
				ID:      i + 1,
				Name:    "Team " + string(rune('A'+i)),
				ClassID: i + 1,
				SportID: sportID,
				EventID: eventID,
			}
		}
		mockSportRepo.On("GetTeamsBySportID", sportID).Return(teams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		h.GenerateAllTournamentsPreviewHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var tournaments []models.GeneratedTournament
		err := json.Unmarshal(w.Body.Bytes(), &tournaments)
		assert.NoError(t, err)
		assert.Len(t, tournaments, 3, "Should have 3 tournaments")

		// Find main tournament
		var mainTournament *models.GeneratedTournament
		for i := range tournaments {
			if tournaments[i].SportName == "Basketball" {
				mainTournament = &tournaments[i]
				break
			}
		}

		assert.NotNil(t, mainTournament, "Main tournament should exist")

		// Verify main tournament has 8 first round matches (matches with RoundIndex 0 and not bronze match)
		firstRoundMatches := []models.Match{}
		for _, match := range mainTournament.TournamentData.Matches {
			if match.RoundIndex == 0 && !match.IsBronzeMatch {
				firstRoundMatches = append(firstRoundMatches, match)
			}
		}

		assert.Equal(t, 8, len(firstRoundMatches), "Main tournament should have 8 first round matches")

		// Verify match order (should be 0-7)
		matchOrders := []int{}
		for _, match := range firstRoundMatches {
			matchOrders = append(matchOrders, match.Order)
		}
		for i := 0; i < 8; i++ {
			assert.Contains(t, matchOrders, i, "Should have match with order %d", i)
		}

		// Note: The actual mapping of losers to loser bracket matches will be handled
		// when match results are updated. At tournament generation time, we just create
		// the structure. The test verifies that:
		// - 3 separate tournaments are created
		// - Main tournament has 8 first round matches
		// - Loser bracket A and B tournaments are created separately
		// The actual loser assignment happens when match results are recorded.
	})
}
