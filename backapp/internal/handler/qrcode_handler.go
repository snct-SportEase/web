package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// QRCodeHandler handles QR code related API requests
type QRCodeHandler struct {
	teamRepo  repository.TeamRepository
	sportRepo repository.SportRepository
	userRepo  repository.UserRepository
	eventRepo repository.EventRepository
}

// NewQRCodeHandler creates a new instance of QRCodeHandler
func NewQRCodeHandler(teamRepo repository.TeamRepository, sportRepo repository.SportRepository, userRepo repository.UserRepository, eventRepo repository.EventRepository) *QRCodeHandler {
	return &QRCodeHandler{
		teamRepo:  teamRepo,
		sportRepo: sportRepo,
		userRepo:  userRepo,
		eventRepo: eventRepo,
	}
}

// GetUserTeamsHandler returns all teams that the current user is a member of
func (h *QRCodeHandler) GetUserTeamsHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	teams, err := h.teamRepo.GetTeamsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve teams"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GenerateQRCodeHandler generates a QR code for a specific event and sport
func (h *QRCodeHandler) GenerateQRCodeHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	var req models.QRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify that the user is assigned to this event and sport
	teams, err := h.teamRepo.GetTeamsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve teams"})
		return
	}

	// Check if user is assigned to the requested event and sport
	isAssigned := false
	var sportName string
	for _, team := range teams {
		if team.EventID == req.EventID && team.SportID == req.SportID {
			isAssigned = true
			sportName = team.SportName
			break
		}
	}

	if !isAssigned {
		c.JSON(http.StatusForbidden, gin.H{"error": "この競技に参加登録されていません"})
		return
	}

	// Get user display name
	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}

	// Generate timestamp and expiration (5 minutes from now)
	timestamp := time.Now().Unix()
	expiresAt := timestamp + (5 * 60) // 5 minutes

	// Create QR code data
	qrData := models.QRCodeData{
		EventID:     req.EventID,
		SportID:     req.SportID,
		SportName:   sportName,
		UserID:      user.ID,
		DisplayName: displayName,
		Timestamp:   timestamp,
		ExpiresAt:   expiresAt,
	}

	// Generate a random token for additional security
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate security token"})
		return
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Combine token and data
	combinedData := map[string]interface{}{
		"token": token,
		"data":  qrData,
	}

	combinedJSON, err := json.Marshal(combinedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode QR code"})
		return
	}

	// Base64 encode the combined data
	qrCodeData := base64.URLEncoding.EncodeToString(combinedJSON)

	response := models.QRCodeResponse{
		QRCodeData: qrCodeData,
		ExpiresAt:  expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyQRCodeHandler verifies a QR code and checks if it's valid (within 5 minutes)
func (h *QRCodeHandler) VerifyQRCodeHandler(c *gin.Context) {
	var req models.QRCodeVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(req.QRCodeData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid QR code format"})
		return
	}

	// Parse JSON
	var combinedData map[string]interface{}
	if err := json.Unmarshal(decoded, &combinedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid QR code data"})
		return
	}

	// Extract data
	dataBytes, err := json.Marshal(combinedData["data"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid QR code data structure"})
		return
	}

	var qrData models.QRCodeData
	if err := json.Unmarshal(dataBytes, &qrData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse QR code data"})
		return
	}

	// Check expiration (5 minutes)
	currentTime := time.Now().Unix()
	if currentTime > qrData.ExpiresAt {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "QRコードの有効期限が切れています",
			"expired":    true,
			"expires_at": qrData.ExpiresAt,
		})
		return
	}

	// Verify user exists and is assigned to the event/sport
	teams, err := h.teamRepo.GetTeamsByUserID(qrData.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user assignment"})
		return
	}

	isAssigned := false
	for _, team := range teams {
		if team.EventID == qrData.EventID && team.SportID == qrData.SportID {
			isAssigned = true
			break
		}
	}

	if !isAssigned {
		c.JSON(http.StatusForbidden, gin.H{"error": "このユーザーはこの競技に参加登録されていません"})
		return
	}

	// Calculate remaining time
	remainingTime := qrData.ExpiresAt - currentTime

	c.JSON(http.StatusOK, gin.H{
		"valid":          true,
		"event_id":       qrData.EventID,
		"sport_id":       qrData.SportID,
		"sport_name":     qrData.SportName,
		"user_id":        qrData.UserID,
		"display_name":   qrData.DisplayName,
		"timestamp":      qrData.Timestamp,
		"expires_at":     qrData.ExpiresAt,
		"remaining_time": remainingTime,
	})
}
