package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GuideDocumentHandler struct {
	repo repository.GuideDocumentRepository
}

func NewGuideDocumentHandler(repo repository.GuideDocumentRepository) *GuideDocumentHandler {
	return &GuideDocumentHandler{repo: repo}
}

func (h *GuideDocumentHandler) ListGuideDocuments(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil || eventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "有効な event_id を指定してください"})
		return
	}

	docs, err := h.repo.ListGuideDocuments(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "資料一覧の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

func (h *GuideDocumentHandler) CreateGuideDocument(c *gin.Context) {
	var payload struct {
		EventID     int     `json:"event_id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		PdfURL      string  `json:"pdf_url"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
		return
	}

	if payload.EventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "有効な event_id は必須です"})
		return
	}

	payload.Title = strings.TrimSpace(payload.Title)
	payload.PdfURL = strings.TrimSpace(payload.PdfURL)
	if payload.Description != nil {
		trimmed := strings.TrimSpace(*payload.Description)
		if trimmed == "" {
			payload.Description = nil
		} else {
			payload.Description = &trimmed
		}
	}

	if payload.Title == "" || payload.PdfURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "タイトルとPDF URLは必須です"})
		return
	}

	doc := &models.GuideDocument{
		EventID:     payload.EventID,
		Title:       payload.Title,
		Description: payload.Description,
		PdfURL:      payload.PdfURL,
	}
	id, err := h.repo.CreateGuideDocument(doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "資料の登録に失敗しました"})
		return
	}
	doc.ID = int(id)

	c.JSON(http.StatusCreated, gin.H{"document": doc})
}

func (h *GuideDocumentHandler) DeleteGuideDocument(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正な資料IDです"})
		return
	}

	if err := h.repo.DeleteGuideDocument(id); err != nil {
		if repository.IsGuideDocumentNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "資料が見つかりません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "資料の削除に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "資料を削除しました"})
}
