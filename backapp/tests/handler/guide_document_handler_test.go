package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGuideDocumentHandler_ListGuideDocuments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockGuideDocumentRepository)
	doc := &models.GuideDocument{
		ID:        1,
		EventID:   5,
		Title:     "競技ルール集",
		PdfURL:    "https://example.com/rules.pdf",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockRepo.On("ListGuideDocuments", 5).Return([]*models.GuideDocument{doc}, nil).Once()

	h := handler.NewGuideDocumentHandler(mockRepo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/guide-documents?event_id=5", nil)

	h.ListGuideDocuments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"title":"競技ルール集"`)
	mockRepo.AssertExpectations(t)
}

func TestGuideDocumentHandler_CreateGuideDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockGuideDocumentRepository)
	h := handler.NewGuideDocumentHandler(mockRepo)

	mockRepo.On("CreateGuideDocument", mock.MatchedBy(func(doc *models.GuideDocument) bool {
		return doc.EventID == 3 && doc.Title == "会場案内" && doc.PdfURL == "https://example.com/map.pdf" && doc.Description != nil && *doc.Description == "当日の導線資料"
	})).Return(int64(10), nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/guide-documents", bytes.NewBufferString(`{"event_id":3,"title":"会場案内","description":"当日の導線資料","pdf_url":"https://example.com/map.pdf"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateGuideDocument(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":10`)
	mockRepo.AssertExpectations(t)
}

func TestGuideDocumentHandler_DeleteGuideDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockGuideDocumentRepository)
		h := handler.NewGuideDocumentHandler(mockRepo)
		mockRepo.On("DeleteGuideDocument", 3).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "3"}}
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/guide-documents/3", nil)

		h.DeleteGuideDocument(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockGuideDocumentRepository)
		h := handler.NewGuideDocumentHandler(mockRepo)
		mockRepo.On("DeleteGuideDocument", 99).Return(sql.ErrNoRows).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "99"}}
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/guide-documents/99", nil)

		h.DeleteGuideDocument(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
