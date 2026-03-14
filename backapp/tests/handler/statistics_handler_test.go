package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatisticsHandler_GetClassScoreTrends(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("hide_scores=true のとき隠蔽メッセージを返す", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournRepo := new(MockTournamentRepository)

		h := handler.NewStatisticsHandler(mockClassRepo, mockEventRepo, mockSportRepo, mockTournRepo)

		activeEvent := &models.Event{
			ID:         1,
			Name:       "2026春季スポーツ大会",
			HideScores: true,
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockEventRepo.On("GetEventByID", 1).Return(activeEvent, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/statistics/scores", nil)

		h.GetClassScoreTrends(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Scores are hidden for this event", resp["message"])

		// スコア取得は呼ばれないこと
		mockEventRepo.AssertNotCalled(t, "GetAllEvents")
		mockClassRepo.AssertNotCalled(t, "GetClassScoresByEvent", mock.Anything)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("hide_scores=false のときスコアデータを返す", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournRepo := new(MockTournamentRepository)

		h := handler.NewStatisticsHandler(mockClassRepo, mockEventRepo, mockSportRepo, mockTournRepo)

		activeEvent := &models.Event{
			ID:         1,
			Name:       "2026春季スポーツ大会",
			HideScores: false,
		}

		events := []*models.Event{
			{ID: 1, Name: "2026春季スポーツ大会"},
		}

		scores := []*models.ClassScore{
			{ClassName: "1-A", TotalPointsCurrentEvent: 100},
			{ClassName: "1-B", TotalPointsCurrentEvent: 80},
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockEventRepo.On("GetEventByID", 1).Return(activeEvent, nil).Once()
		mockEventRepo.On("GetAllEvents").Return(events, nil).Once()
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(scores, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/statistics/scores", nil)

		h.GetClassScoreTrends(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		// "message" キーが存在しないこと
		assert.NotContains(t, resp, "message")
		// イベント名のキーでスコアが返ること
		assert.Contains(t, resp, "2026春季スポーツ大会")
		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
	})

	t.Run("アクティブイベント取得エラー時に500を返す", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournRepo := new(MockTournamentRepository)

		h := handler.NewStatisticsHandler(mockClassRepo, mockEventRepo, mockSportRepo, mockTournRepo)

		mockEventRepo.On("GetActiveEvent").Return(0, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/statistics/scores", nil)

		h.GetClassScoreTrends(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestStatisticsHandler_GetOverallAttendanceRate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("出席率を正しく計算して返す", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournRepo := new(MockTournamentRepository)

		h := handler.NewStatisticsHandler(mockClassRepo, mockEventRepo, mockSportRepo, mockTournRepo)

		classes := []*models.Class{
			{ID: 1, Name: "1-A", AttendCount: 30, StudentCount: 40},
			{ID: 2, Name: "1-B", AttendCount: 20, StudentCount: 40},
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetAllClasses", 1).Return(classes, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/statistics/attendance", nil)

		h.GetOverallAttendanceRate(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		// (30+20) / (40+40) * 100 = 62.5
		assert.InDelta(t, 62.5, resp["attendance_rate"], 0.001)
		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
	})

	t.Run("クラスが0人のとき出席率0を返す", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournRepo := new(MockTournamentRepository)

		h := handler.NewStatisticsHandler(mockClassRepo, mockEventRepo, mockSportRepo, mockTournRepo)

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetAllClasses", 1).Return([]*models.Class{}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/statistics/attendance", nil)

		h.GetOverallAttendanceRate(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 0.0, resp["attendance_rate"])
		mockEventRepo.AssertExpectations(t)
	})
}
