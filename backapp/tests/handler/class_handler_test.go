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

func TestClassHandler_GetClassScores(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo)

		expectedScores := []*models.ClassScore{
			{ID: 1, ClassName: "Class A", Season: "spring", TotalPointsCurrentEvent: 100, RankCurrentEvent: 1},
			{ID: 2, ClassName: "Class B", Season: "spring", TotalPointsCurrentEvent: 90, RankCurrentEvent: 2},
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(expectedScores, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualScores []*models.ClassScore
		err := json.Unmarshal(w.Body.Bytes(), &actualScores)
		assert.NoError(t, err)
		assert.Equal(t, expectedScores, actualScores)

		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("No active event", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo)

		mockEventRepo.On("GetActiveEvent").Return(0, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertNotCalled(t, "GetClassScoresByEvent", mock.Anything)
	})
}
