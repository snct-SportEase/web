package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestClassHandler_ExportClassScoresCSVHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedScores := []*models.ClassScore{
		{
			ID:                      1,
			EventID:                 1,
			ClassID:                 1,
			ClassName:               "1A",
			RankOverall:             1,
			TotalPointsOverall:      100,
			RankCurrentEvent:        2,
			TotalPointsCurrentEvent: 50,
			SurveyPoints:            10,
			AttendancePoints:        20,
			InitialPoints:           20,
		},
		{
			ID:                      2,
			EventID:                 1,
			ClassID:                 2,
			ClassName:               "1B",
			RankOverall:             2,
			TotalPointsOverall:      80,
			RankCurrentEvent:        1,
			TotalPointsCurrentEvent: 60,
			SurveyPoints:            5,
			AttendancePoints:        5,
			InitialPoints:           10,
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(expectedScores, nil).Once()

		h := handler.NewClassHandler(mockClassRepo, nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		h.ExportClassScoresCSVHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, "attachment; filename=\"event_1_class_scores.csv\"", w.Header().Get("Content-Disposition"))

		body := w.Body.Bytes()
		// Start with BOM
		assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, body[:3])

		csvReader := csv.NewReader(strings.NewReader(string(body[3:])))
		records, err := csvReader.ReadAll()
		assert.NoError(t, err)

		// Header + 2 rows
		assert.Len(t, records, 3)
		assert.Equal(t, []string{"順位 (総合)", "クラス名", "総合スコア", "順位 (今大会)", "今大会スコア", "事前回答点", "出席点", "初期点"}, records[0])

		// 1A row
		assert.Equal(t, []string{"1", "1A", "100", "2", "50", "10", "20", "20"}, records[1])
		// 1B row
		assert.Equal(t, []string{"2", "1B", "80", "1", "60", "5", "5", "10"}, records[2])
	})

	t.Run("Invalid Event ID", func(t *testing.T) {
		h := handler.NewClassHandler(nil, nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "invalid"}}

		h.ExportClassScoresCSVHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Database Error", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(nil, assert.AnError).Once()

		h := handler.NewClassHandler(mockClassRepo, nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		h.ExportClassScoresCSVHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
