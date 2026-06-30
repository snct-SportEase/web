package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BarcodeHandler handles MyID barcode related API requests.
type BarcodeHandler struct {
	teamRepo  repository.TeamRepository
	sportRepo repository.SportRepository
	userRepo  repository.UserRepository
	eventRepo repository.EventRepository
	classRepo repository.ClassRepository
	tournRepo repository.TournamentRepository
}

const myIDBarcodePrefix = "H10"
const studentNumberLength = 7

// NewBarcodeHandler creates a new instance of BarcodeHandler.
func NewBarcodeHandler(teamRepo repository.TeamRepository, sportRepo repository.SportRepository, userRepo repository.UserRepository, eventRepo repository.EventRepository, classRepo repository.ClassRepository, tournRepo repository.TournamentRepository) *BarcodeHandler {
	return &BarcodeHandler{
		teamRepo:  teamRepo,
		sportRepo: sportRepo,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		classRepo: classRepo,
		tournRepo: tournRepo,
	}
}

// GetUserTeamsHandler returns all teams that the current user is a member of.
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

// CheckInRoundHandler reads a MyID barcode and records round check-in for a pre-entered student.
func (h *BarcodeHandler) CheckInRoundHandler(c *gin.Context) {
	var req models.BarcodeCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.EventID == 0 || req.SportID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "イベントIDと競技IDが必要です"})
		return
	}
	if req.MatchID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "試合を選択してください"})
		return
	}

	trimmedBarcode := strings.TrimSpace(req.BarcodeData)
	studentNumber, err := parseMyIDBarcode(trimmedBarcode)
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

	team, err := h.findPreEnteredTeam(user.ID, req.EventID, req.SportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user assignment"})
		return
	}
	if team == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "このユーザーはこの競技に事前エントリーされていません"})
		return
	}

	match, err := h.tournRepo.GetMatchForEventSport(req.MatchID, req.EventID, req.SportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify selected match"})
		return
	}
	if match == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "選択した試合がこの競技に存在しません"})
		return
	}
	if match.Round < 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify selected match"})
		return
	}
	round := match.Round + 1

	teamDetails, err := h.teamRepo.GetTeamByClassAndSport(team.ClassID, req.SportID, req.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team details"})
		return
	}
	if teamDetails == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Team not found"})
		return
	}

	if err := h.teamRepo.ConfirmTeamMember(teamDetails.ID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm team member"})
		return
	}

	if err := h.teamRepo.CheckInRound(team.ID, user.ID, req.EventID, req.SportID, round); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in round"})
		return
	}

	var capacityWarning *string
	if teamDetails.MinCapacity != nil {
		confirmedCount, err := h.teamRepo.GetConfirmedTeamMembersCount(teamDetails.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team capacity"})
			return
		}

		if confirmedCount < *teamDetails.MinCapacity {
			className := "不明"
			class, err := h.classRepo.GetClassByID(team.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class information"})
				return
			}
			if class != nil {
				className = class.Name
			}

			warningMsg := fmt.Sprintf("%sクラスの%sの参加本登録済みメンバー数（%d人）が最低人数（%d人）に達していません",
				className, team.SportName, confirmedCount, *teamDetails.MinCapacity)
			capacityWarning = &warningMsg
		}
	}

	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}

	response := gin.H{
		"valid":          true,
		"confirmed":      true,
		"checked_in":     true,
		"event_id":       req.EventID,
		"sport_id":       req.SportID,
		"sport_name":     team.SportName,
		"round":          round,
		"match_id":       req.MatchID,
		"team_id":        team.ID,
		"user_id":        user.ID,
		"display_name":   displayName,
		"student_number": studentNumber,
		"barcode_data":   trimmedBarcode,
	}
	if capacityWarning != nil {
		response["capacity_warning"] = *capacityWarning
	}

	c.JSON(http.StatusOK, response)
}

func parseMyIDBarcode(barcodeData string) (string, error) {
	barcode := strings.TrimSpace(barcodeData)
	if !strings.HasPrefix(barcode, myIDBarcodePrefix) {
		return "", fmt.Errorf("invalid MyID barcode prefix")
	}

	studentNumber := strings.TrimPrefix(barcode, myIDBarcodePrefix)
	if len(studentNumber) != studentNumberLength {
		return "", fmt.Errorf("student number must be %d digits", studentNumberLength)
	}

	for _, char := range studentNumber {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("student number must be numeric")
		}
	}

	return studentNumber, nil
}

func (h *BarcodeHandler) findUserByStudentNumber(studentNumber string) (*models.User, error) {
	emailLocalPartForStudentNumber := "s" + studentNumber
	users, err := h.userRepo.FindUsers(emailLocalPartForStudentNumber, "email")
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		emailLocalPart := strings.SplitN(user.Email, "@", 2)[0]
		if strings.EqualFold(emailLocalPart, emailLocalPartForStudentNumber) {
			return user, nil
		}
	}

	return nil, nil
}

func (h *BarcodeHandler) findPreEnteredTeam(userID string, eventID int, sportID int) (*models.TeamWithSport, error) {
	teams, err := h.teamRepo.GetTeamsByUserID(userID)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		if team.EventID == eventID && team.SportID == sportID {
			return team, nil
		}
	}

	return nil, nil
}
