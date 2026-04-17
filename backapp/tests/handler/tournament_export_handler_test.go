package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"backapp/internal/websocket"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestTournamentHandler_ExportTournamentsExcelHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeTournamentData := func() json.RawMessage {
		data, err := json.Marshal(models.TournamentData{
			Rounds: []models.Round{
				{Name: "決勝"},
			},
			Matches: []models.Match{
				{
					RoundIndex: 0,
					Order:      0,
					Sides: []models.Side{
						{ContestantID: "c0", Scores: []models.Score{{MainScore: 3}}, IsWinner: true},
						{ContestantID: "c1", Scores: []models.Score{{MainScore: 1}}},
					},
				},
			},
			Contestants: map[string]models.Contestant{
				"c0": {Players: []models.Player{{Title: "1A"}}},
				"c1": {Players: []models.Player{{Title: "2B"}}},
			},
		})
		assert.NoError(t, err)
		return data
	}

	t.Run("Success", func(t *testing.T) {
		mockTournRepo := new(MockTournamentRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		hubManager := websocket.NewHubManager()

		mockEventRepo.On("GetEventByID", 1).Return(&models.Event{ID: 1, Name: "2025春季スポーツ大会"}, nil).Once()
		mockTournRepo.On("GetTournamentsByEventID", 1).Return([]*models.Tournament{
			{
				ID:   1,
				Name: "バスケットボール",
				Data: makeTournamentData(),
			},
		}, nil).Once()

		h := handler.NewTournamentHandler(mockTournRepo, mockSportRepo, mockTeamRepo, mockClassRepo, mockEventRepo, hubManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		h.ExportTournamentsExcelHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
		assert.Equal(t, "attachment; filename=\"event_1_tournaments.xlsx\"", w.Header().Get("Content-Disposition"))

		file, err := excelize.OpenReader(bytes.NewReader(w.Body.Bytes()))
		assert.NoError(t, err)

		sheets := file.GetSheetList()
		assert.Equal(t, []string{"バスケットボール"}, sheets)

		title, err := file.GetCellValue("バスケットボール", "A1")
		assert.NoError(t, err)
		assert.Equal(t, "2025春季スポーツ大会 - バスケットボール", title)

		roundHeader, err := file.GetCellValue("バスケットボール", "A3")
		assert.NoError(t, err)
		assert.Equal(t, "決勝", roundHeader)

		team1, err := file.GetCellValue("バスケットボール", "A5")
		assert.NoError(t, err)
		assert.Equal(t, "1A", team1)

		team2, err := file.GetCellValue("バスケットボール", "A7")
		assert.NoError(t, err)
		assert.Equal(t, "2B", team2)
	})

	t.Run("No tournaments", func(t *testing.T) {
		mockTournRepo := new(MockTournamentRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		hubManager := websocket.NewHubManager()

		mockTournRepo.On("GetTournamentsByEventID", 1).Return([]*models.Tournament{}, nil).Once()

		h := handler.NewTournamentHandler(mockTournRepo, mockSportRepo, mockTeamRepo, mockClassRepo, mockEventRepo, hubManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		h.ExportTournamentsExcelHandler(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
