package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMICHandler_GetEligibleClasses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		classes := []models.Class{{ID: 1, Name: "1-1"}}
		mockRepo.On("GetEligibleClasses", 1).Return(classes, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/mic/eligible-classes?event_id=1", nil)

		h.GetEligibleClasses(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []models.Class
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
		assert.Equal(t, "1-1", resp[0].Name)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		h := handler.NewMICHandler(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/mic/eligible-classes?event_id=abc", nil)
		h.GetEligibleClasses(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestMICHandler_VoteMIC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		user := &models.User{ID: "user1"}
		req := handler.MICVoteRequest{VotedForClassID: 2, EventID: 1, Reason: "nice"}

		mockRepo.On("VoteMIC", "user1", 2, 1, "nice").Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		jsonBody, _ := json.Marshal(req)
		c.Request, _ = http.NewRequest("POST", "/api/mic/vote", bytes.NewBuffer(jsonBody))

		h.VoteMIC(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		h := handler.NewMICHandler(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody, _ := json.Marshal(handler.MICVoteRequest{})
		c.Request, _ = http.NewRequest("POST", "/api/mic/vote", bytes.NewBuffer(jsonBody))

		h.VoteMIC(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestMICHandler_GetMICVotes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		votes := []models.MICVote{{VotedForClassID: 1, Points: 10}}
		mockRepo.On("GetMICVotes", 1).Return(votes, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/mic/votes?event_id=1", nil)

		h.GetMICVotes(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMICHandler_GetUserVote(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Voted", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		user := &models.User{ID: "user1"}
		vote := &models.MICVote{VotedForClassID: 2}
		mockRepo.On("GetVoteByUserID", "user1", 1).Return(vote, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/mic/user-vote?event_id=1", nil)

		h.GetUserVote(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["voted"].(bool))
	})

	t.Run("Not Voted", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		user := &models.User{ID: "user1"}
		mockRepo.On("GetVoteByUserID", "user1", 1).Return(nil, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest("GET", "/api/mic/user-vote?event_id=1", nil)

		h.GetUserVote(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["voted"].(bool))
	})
}

func TestMICHandler_GetMICClass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		res := &models.MICResult{ClassName: "1-1", TotalPoints: 100}
		mockRepo.On("GetMICClass", 1).Return(res, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/mic/result?event_id=1", nil)

		h.GetMICClass(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockMICRepository)
		h := handler.NewMICHandler(mockRepo)

		mockRepo.On("GetMICClass", 1).Return(nil, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/mic/result?event_id=1", nil)

		h.GetMICClass(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "No MIC class found yet", resp["message"])
	})
}
