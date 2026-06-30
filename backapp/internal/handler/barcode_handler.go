package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// BarcodeHandler handles barcode related API requests
type BarcodeHandler struct {
	teamRepo   repository.TeamRepository
	sportRepo  repository.SportRepository
	userRepo   repository.UserRepository
	eventRepo  repository.EventRepository
	classRepo  repository.ClassRepository
	tokenStore BarcodeTokenStore
}

const barcodeTTL = 10 * time.Second
const myIDBarcodePrefix = "H10"

// NewBarcodeHandler creates a new instance of BarcodeHandler
func NewBarcodeHandler(teamRepo repository.TeamRepository, sportRepo repository.SportRepository, userRepo repository.UserRepository, eventRepo repository.EventRepository, classRepo repository.ClassRepository) *BarcodeHandler {
	return &BarcodeHandler{
		teamRepo:   teamRepo,
		sportRepo:  sportRepo,
		userRepo:   userRepo,
		eventRepo:  eventRepo,
		classRepo:  classRepo,
		tokenStore: NewRedisBarcodeTokenStore(),
	}
}

func (h *BarcodeHandler) WithTokenStore(tokenStore BarcodeTokenStore) *BarcodeHandler {
	h.tokenStore = tokenStore
	return h
}

// GetUserTeamsHandler returns all teams that the current user is a member of
func (h *BarcodeHandler) GetUserTeamsHandler(c *gin.Context) {
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

// GenerateBarcodeHandler generates a barcode for a specific event and sport
func (h *BarcodeHandler) GenerateBarcodeHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	var req models.BarcodeRequest
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

	// Generate timestamp and expiration (10 seconds from now)
	timestamp := time.Now().Unix()
	expiresAt := timestamp + int64(barcodeTTL/time.Second)

	// Create barcode data
	barcodeTokenData := models.BarcodeData{
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

	if err := h.tokenStore.SaveActiveToken(user.ID, req.EventID, req.SportID, token, barcodeTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store barcode token"})
		return
	}

	// Combine token and data
	combinedData := map[string]interface{}{
		"token": token,
		"data":  barcodeTokenData,
	}

	combinedJSON, err := json.Marshal(combinedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode barcode"})
		return
	}

	// Base64 encode the combined data
	barcodeData := base64.URLEncoding.EncodeToString(combinedJSON)

	response := models.BarcodeResponse{
		BarcodeData: barcodeData,
		ExpiresAt:   expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyBarcodeHandler verifies a barcode and checks if it's valid (within 5 minutes)
func (h *BarcodeHandler) VerifyBarcodeHandler(c *gin.Context) {
	var req models.BarcodeVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	barcodeData := strings.TrimSpace(req.BarcodeData)
	if isMyIDBarcodeRequest(req, barcodeData) {
		h.verifyMyIDBarcode(c, req)
		return
	}

	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(barcodeData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barcode format"})
		return
	}

	// Parse JSON
	var combinedData map[string]interface{}
	if err := json.Unmarshal(decoded, &combinedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barcode data"})
		return
	}

	// Extract data
	dataBytes, err := json.Marshal(combinedData["data"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barcode data structure"})
		return
	}

	var barcodeTokenData models.BarcodeData
	if err := json.Unmarshal(dataBytes, &barcodeTokenData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse barcode data"})
		return
	}

	token, ok := combinedData["token"].(string)
	if !ok || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barcode token"})
		return
	}

	activeToken, err := h.tokenStore.GetActiveToken(barcodeTokenData.UserID, barcodeTokenData.EventID, barcodeTokenData.SportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate barcode token"})
		return
	}
	if activeToken == "" || activeToken != token {
		c.JSON(http.StatusBadRequest, gin.H{"error": "バーコードは無効または期限切れです"})
		return
	}

	// Check expiration (10 seconds)
	currentTime := time.Now().Unix()
	if currentTime > barcodeTokenData.ExpiresAt {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "バーコードの有効期限が切れています",
			"expired":    true,
			"expires_at": barcodeTokenData.ExpiresAt,
		})
		return
	}

	// Verify user exists and is assigned to the event/sport
	teams, err := h.teamRepo.GetTeamsByUserID(barcodeTokenData.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user assignment"})
		return
	}

	var targetTeam *models.TeamWithSport
	for _, team := range teams {
		if team.EventID == barcodeTokenData.EventID && team.SportID == barcodeTokenData.SportID {
			targetTeam = team
			break
		}
	}

	if targetTeam == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "このユーザーはこの競技に参加登録されていません"})
		return
	}

	// Get full team details including capacity
	team, err := h.teamRepo.GetTeamByClassAndSport(targetTeam.ClassID, barcodeTokenData.SportID, barcodeTokenData.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team details"})
		return
	}

	if team == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Team not found"})
		return
	}

	consumed, err := h.tokenStore.ConsumeActiveToken(barcodeTokenData.UserID, barcodeTokenData.EventID, barcodeTokenData.SportID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consume barcode token"})
		return
	}
	if !consumed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "バーコードは無効または期限切れです"})
		return
	}

	// Confirm team member (参加本登録)
	err = h.teamRepo.ConfirmTeamMember(team.ID, barcodeTokenData.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm team member"})
		return
	}

	// Check minimum capacity requirement
	var capacityWarning *string
	if team.MinCapacity != nil {
		confirmedCount, err := h.teamRepo.GetConfirmedTeamMembersCount(team.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team capacity"})
			return
		}

		if confirmedCount < *team.MinCapacity {
			// Get class name for warning message
			class, err := h.classRepo.GetClassByID(targetTeam.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class information"})
				return
			}
			className := "不明"
			if class != nil {
				className = class.Name
			}
			warningMsg := fmt.Sprintf("%sクラスの%sの参加本登録済みメンバー数（%d人）が最低人数（%d人）に達していません",
				className, barcodeTokenData.SportName, confirmedCount, *team.MinCapacity)
			capacityWarning = &warningMsg
		}
	}

	// Calculate remaining time
	remainingTime := barcodeTokenData.ExpiresAt - currentTime

	response := gin.H{
		"valid":          true,
		"event_id":       barcodeTokenData.EventID,
		"sport_id":       barcodeTokenData.SportID,
		"sport_name":     barcodeTokenData.SportName,
		"user_id":        barcodeTokenData.UserID,
		"display_name":   barcodeTokenData.DisplayName,
		"timestamp":      barcodeTokenData.Timestamp,
		"expires_at":     barcodeTokenData.ExpiresAt,
		"remaining_time": remainingTime,
	}

	if capacityWarning != nil {
		response["capacity_warning"] = *capacityWarning
	}

	c.JSON(http.StatusOK, response)
}

func isMyIDBarcodeRequest(req models.BarcodeVerifyRequest, barcodeData string) bool {
	return req.EventID != 0 || req.SportID != 0 || strings.HasPrefix(barcodeData, myIDBarcodePrefix)
}

func (h *BarcodeHandler) verifyMyIDBarcode(c *gin.Context, req models.BarcodeVerifyRequest) {
	if req.EventID == 0 || req.SportID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "イベントIDと競技IDが必要です"})
		return
	}

	studentNumber, err := parseMyIDBarcode(req.BarcodeData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "バーコード形式が不正です"})
		return
	}

	user, err := h.findUserByStudentNumber(studentNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user by student number"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "該当する学生が見つかりません"})
		return
	}

	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}

	response, err := h.confirmParticipation(user.ID, req.EventID, req.SportID, "", displayName)
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()
		if message == "not assigned" {
			status = http.StatusForbidden
			message = "このユーザーはこの競技に参加登録されていません"
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	response["student_number"] = studentNumber
	response["barcode_data"] = strings.TrimSpace(req.BarcodeData)
	c.JSON(http.StatusOK, response)
}

func parseMyIDBarcode(barcodeData string) (string, error) {
	barcode := strings.TrimSpace(barcodeData)
	if !strings.HasPrefix(barcode, myIDBarcodePrefix) {
		return "", fmt.Errorf("invalid MyID barcode prefix")
	}

	studentNumber := strings.TrimPrefix(barcode, myIDBarcodePrefix)
	if studentNumber == "" {
		return "", fmt.Errorf("student number is empty")
	}

	for _, char := range studentNumber {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("student number must be numeric")
		}
	}

	return studentNumber, nil
}

func (h *BarcodeHandler) findUserByStudentNumber(studentNumber string) (*models.User, error) {
	users, err := h.userRepo.FindUsers(studentNumber, "email")
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		emailLocalPart := strings.SplitN(user.Email, "@", 2)[0]
		if strings.EqualFold(emailLocalPart, studentNumber) {
			return user, nil
		}
	}

	return nil, nil
}

func (h *BarcodeHandler) confirmParticipation(userID string, eventID int, sportID int, sportName string, displayName string) (gin.H, error) {
	teams, err := h.teamRepo.GetTeamsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("Failed to verify user assignment")
	}

	var targetTeam *models.TeamWithSport
	for _, team := range teams {
		if team.EventID == eventID && team.SportID == sportID {
			targetTeam = team
			break
		}
	}

	if targetTeam == nil {
		return nil, fmt.Errorf("not assigned")
	}

	if sportName == "" {
		sportName = targetTeam.SportName
	}

	team, err := h.teamRepo.GetTeamByClassAndSport(targetTeam.ClassID, sportID, eventID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get team details")
	}

	if team == nil {
		return nil, fmt.Errorf("Team not found")
	}

	if err := h.teamRepo.ConfirmTeamMember(team.ID, userID); err != nil {
		return nil, fmt.Errorf("Failed to confirm team member")
	}

	var capacityWarning *string
	if team.MinCapacity != nil {
		confirmedCount, err := h.teamRepo.GetConfirmedTeamMembersCount(team.ID)
		if err != nil {
			return nil, fmt.Errorf("Failed to check team capacity")
		}

		if confirmedCount < *team.MinCapacity {
			class, err := h.classRepo.GetClassByID(targetTeam.ClassID)
			if err != nil {
				return nil, fmt.Errorf("Failed to get class information")
			}
			className := "不明"
			if class != nil {
				className = class.Name
			}
			warningMsg := fmt.Sprintf("%sクラスの%sの参加本登録済みメンバー数（%d人）が最低人数（%d人）に達していません",
				className, sportName, confirmedCount, *team.MinCapacity)
			capacityWarning = &warningMsg
		}
	}

	response := gin.H{
		"valid":        true,
		"event_id":     eventID,
		"sport_id":     sportID,
		"sport_name":   sportName,
		"user_id":      userID,
		"display_name": displayName,
	}

	if capacityWarning != nil {
		response["capacity_warning"] = *capacityWarning
	}

	return response, nil
}
