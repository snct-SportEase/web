package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

	selectedMatchIDs := normalizeMatchIDs(req.MatchID, req.MatchIDs)
	if len(selectedMatchIDs) == 0 {
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

	match, selectionValid, err := h.findMatchingSelectedMatch(selectedMatchIDs, req.EventID, req.SportID, team.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify selected match"})
		return
	}
	if !selectionValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "選択した試合がこの競技に存在しません"})
		return
	}
	if match == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "まだあなたのクラスはこの試合にチェックインできません"})
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

	if err := h.teamRepo.CheckInRound(team.ID, user.ID, req.EventID, req.SportID, match.ID, round); err != nil {
		if errors.Is(err, repository.ErrRoundAlreadyCheckedIn) {
			c.JSON(http.StatusConflict, gin.H{
				"error":              "チェックイン済みです",
				"already_checked_in": true,
			})
			return
		}
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
		"match_id":       match.ID,
		"match_ids":      selectedMatchIDs,
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

// GetMatchCheckInsHandler returns students checked in for a selected match.
func (h *BarcodeHandler) GetMatchCheckInsHandler(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil || matchID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match_id"})
		return
	}

	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil || eventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "イベントIDが必要です"})
		return
	}

	sportID, err := strconv.Atoi(c.Query("sport_id"))
	if err != nil || sportID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "競技IDが必要です"})
		return
	}

	selectedMatchIDs := parseMatchIDsQuery(matchID, c.Query("match_ids"))
	members := make([]*models.MatchCheckInMember, 0)
	matchTeamIDs := make([]int, 0)
	seenTeamIDs := make(map[int]bool)
	for _, selectedMatchID := range selectedMatchIDs {
		match, err := h.tournRepo.GetMatchForEventSport(selectedMatchID, eventID, sportID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify selected match"})
			return
		}
		if match == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "選択した試合がこの競技に存在しません"})
			return
		}
		for _, teamID := range getMatchTeamIDs(match) {
			if seenTeamIDs[teamID] {
				continue
			}
			seenTeamIDs[teamID] = true
			matchTeamIDs = append(matchTeamIDs, teamID)
		}

		matchMembers, err := h.teamRepo.GetMatchCheckIns(eventID, sportID, selectedMatchID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match check-ins"})
			return
		}
		members = append(members, matchMembers...)
	}

	uncheckedMembers, err := h.buildUncheckedMatchMembers(eventID, sportID, matchTeamIDs, members)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members":            members,
		"count":              len(members),
		"checked_in_members": members,
		"checked_in_count":   len(members),
		"unchecked_members":  uncheckedMembers,
		"unchecked_count":    len(uncheckedMembers),
	})
}

func (h *BarcodeHandler) buildUncheckedMatchMembers(eventID int, sportID int, teamIDs []int, checkedMembers []*models.MatchCheckInMember) ([]*models.MatchCheckInMember, error) {
	uncheckedMembers := make([]*models.MatchCheckInMember, 0)
	if len(teamIDs) == 0 {
		return uncheckedMembers, nil
	}

	teamMembersByTeamID, err := h.teamRepo.GetTeamMembersByTeamIDs(teamIDs)
	if err != nil {
		return nil, err
	}

	checkedUserIDs := make(map[string]bool, len(checkedMembers))
	for _, member := range checkedMembers {
		if member != nil {
			checkedUserIDs[member.UserID] = true
		}
	}

	for _, teamID := range teamIDs {
		for _, user := range teamMembersByTeamID[teamID] {
			if user == nil || checkedUserIDs[user.ID] {
				continue
			}

			classID := 0
			if user.ClassID != nil {
				classID = *user.ClassID
			}
			uncheckedMembers = append(uncheckedMembers, &models.MatchCheckInMember{
				UserID:      user.ID,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				ClassID:     classID,
				TeamID:      teamID,
				EventID:     eventID,
				SportID:     sportID,
			})
		}
	}

	return uncheckedMembers, nil
}

func getMatchTeamIDs(match *models.MatchDB) []int {
	teamIDs := make([]int, 0, 2)
	if match.Team1ID.Valid {
		teamIDs = append(teamIDs, int(match.Team1ID.Int64))
	}
	if match.Team2ID.Valid {
		teamIDs = append(teamIDs, int(match.Team2ID.Int64))
	}
	return teamIDs
}

func normalizeMatchIDs(primaryMatchID int, matchIDs []int) []int {
	seen := make(map[int]bool)
	normalized := make([]int, 0, len(matchIDs)+1)
	for _, matchID := range append([]int{primaryMatchID}, matchIDs...) {
		if matchID <= 0 || seen[matchID] {
			continue
		}
		seen[matchID] = true
		normalized = append(normalized, matchID)
	}
	return normalized
}

func parseMatchIDsQuery(primaryMatchID int, rawMatchIDs string) []int {
	if rawMatchIDs == "" {
		return []int{primaryMatchID}
	}

	matchIDs := make([]int, 0)
	for _, rawID := range strings.Split(rawMatchIDs, ",") {
		matchID, err := strconv.Atoi(strings.TrimSpace(rawID))
		if err == nil {
			matchIDs = append(matchIDs, matchID)
		}
	}
	return normalizeMatchIDs(primaryMatchID, matchIDs)
}

func parseMyIDBarcode(barcodeData string) (string, error) {
	barcode := strings.TrimSpace(barcodeData)
	if len(barcode) >= 2 && strings.HasPrefix(barcode, "*") && strings.HasSuffix(barcode, "*") {
		barcode = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(barcode, "*"), "*"))
	}
	barcode = strings.ToUpper(barcode)

	if !strings.HasPrefix(barcode, myIDBarcodePrefix) {
		return "", fmt.Errorf("invalid MyID barcode prefix")
	}

	studentNumberWithCheckDigit := strings.TrimPrefix(barcode, myIDBarcodePrefix)
	if len(studentNumberWithCheckDigit) != studentNumberLength+1 {
		return "", fmt.Errorf("student number must be %d digits plus a check digit", studentNumberLength)
	}

	for _, char := range studentNumberWithCheckDigit {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("student number must be numeric")
		}
	}

	return studentNumberWithCheckDigit[:studentNumberLength], nil
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

func (h *BarcodeHandler) findMatchingSelectedMatch(matchIDs []int, eventID int, sportID int, teamID int) (*models.MatchDB, bool, error) {
	for _, matchID := range matchIDs {
		match, err := h.tournRepo.GetMatchForEventSport(matchID, eventID, sportID)
		if err != nil {
			return nil, true, err
		}
		if match == nil {
			return nil, false, nil
		}
		if match.Round < 0 {
			return nil, true, fmt.Errorf("invalid match round")
		}
		if isTeamInMatch(teamID, match) {
			return match, true, nil
		}
	}

	return nil, true, nil
}

func isTeamInMatch(teamID int, match *models.MatchDB) bool {
	if match == nil {
		return false
	}
	if match.Team1ID.Valid && int(match.Team1ID.Int64) == teamID {
		return true
	}
	if match.Team2ID.Valid && int(match.Team2ID.Int64) == teamID {
		return true
	}
	return false
}
