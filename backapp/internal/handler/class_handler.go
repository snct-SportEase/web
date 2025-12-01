package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClassHandler struct {
	classRepo      repository.ClassRepository
	eventRepo      repository.EventRepository
	teamRepo       repository.TeamRepository
	tournamentRepo repository.TournamentRepository
}

func NewClassHandler(classRepo repository.ClassRepository, eventRepo repository.EventRepository, teamRepo repository.TeamRepository, tournamentRepo repository.TournamentRepository) *ClassHandler {
	return &ClassHandler{
		classRepo:      classRepo,
		eventRepo:      eventRepo,
		teamRepo:       teamRepo,
		tournamentRepo: tournamentRepo,
	}
}

func (h *ClassHandler) GetAllClasses(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	fmt.Printf("activeEventID: %d\n", activeEventID)
	if activeEventID == 0 {
		// アクティブなイベントがない場合は空の配列を返す
		// 初回ログイン時など、イベントがまだ設定されていない場合に対応
		c.JSON(http.StatusOK, []*models.Class{})
		return
	}

	classes, err := h.classRepo.GetAllClasses(activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

type UpdateStudentCountRequest struct {
	ClassID      int `json:"class_id"`
	StudentCount int `json:"student_count"`
}

func (h *ClassHandler) UpdateStudentCountsHandler(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	var req []UpdateStudentCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	counts := make(map[int]int)
	for _, item := range req {
		counts[item.ClassID] = item.StudentCount
	}

	if err := h.classRepo.UpdateStudentCounts(activeEventID, counts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student counts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student counts updated successfully"})
}

func (h *ClassHandler) UpdateStudentCountsFromCSVHandler(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	file, _, err := c.Request.FormFile("csv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not provided"})
		return
	}
	defer file.Close()

	// Get all classes for the active event to map names to IDs
	allClasses, err := h.classRepo.GetAllClasses(activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get classes"})
		return
	}
	classNameToID := make(map[string]int)
	for _, class := range allClasses {
		classNameToID[class.Name] = class.ID
	}

	reader := csv.NewReader(file)
	counts := make(map[int]int)

	// Skip header row
	if _, err := reader.Read(); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV header: " + err.Error()})
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CSV record: " + err.Error()})
			return
		}

		if len(record) < 2 {
			continue // Skip empty or invalid rows
		}

		className := record[0]
		studentCountStr := record[1]

		classID, ok := classNameToID[className]
		if !ok {
			// If class name is not found, you might want to log this or handle it
			continue
		}

		studentCount, err := strconv.Atoi(studentCountStr)
		if err != nil {
			// Handle error for invalid number format
			continue
		}

		counts[classID] = studentCount
	}

	if len(counts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid data found in CSV file"})
		return
	}

	if err := h.classRepo.UpdateStudentCounts(activeEventID, counts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student counts from CSV: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student counts updated successfully from CSV"})
}

func (h *ClassHandler) GetClassScores(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	scores, err := h.classRepo.GetClassScoresByEvent(activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scores)
}

func (h *ClassHandler) GetClassProgress(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user"})
		return
	}

	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	class, err := h.classRepo.GetClassByRepRole(user.ID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class information"})
		return
	}
	if class == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Class representative role is required"})
		return
	}

	teams, err := h.teamRepo.GetTeamsByClassID(class.ID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get teams"})
		return
	}

	members, err := h.classRepo.GetClassMembers(class.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class members"})
		return
	}

	memberLookup := make(map[string]*models.ClassMemberView)
	memberList := make([]*models.ClassMemberView, 0, len(members))
	for _, m := range members {
		view := &models.ClassMemberView{
			ID:          m.ID,
			Email:       m.Email,
			DisplayName: m.DisplayName,
			Assignments: []models.ClassMemberAssignment{},
		}
		memberLookup[m.ID] = view
		memberList = append(memberList, view)
	}

	var progress []models.ClassProgress
	for _, team := range teams {
		matchDetails, err := h.tournamentRepo.GetMatchesForTeam(activeEventID, team.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match details"})
			return
		}
		teamMembers, err := h.teamRepo.GetTeamMembers(team.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members"})
			return
		}
		for _, tm := range teamMembers {
			view, exists := memberLookup[tm.ID]
			if !exists {
				view = &models.ClassMemberView{
					ID:          tm.ID,
					Email:       tm.Email,
					DisplayName: tm.DisplayName,
					Assignments: []models.ClassMemberAssignment{},
				}
				memberLookup[tm.ID] = view
				memberList = append(memberList, view)
			}
			view.Assignments = append(view.Assignments, models.ClassMemberAssignment{
				SportName: team.SportName,
				TeamName:  team.Name,
			})
		}

		entry := buildClassProgress(team, matchDetails)
		progress = append(progress, entry)
	}

	sort.SliceStable(memberList, func(i, j int) bool {
		var left, right string
		if memberList[i].DisplayName != nil && *memberList[i].DisplayName != "" {
			left = *memberList[i].DisplayName
		} else {
			left = memberList[i].Email
		}
		if memberList[j].DisplayName != nil && *memberList[j].DisplayName != "" {
			right = *memberList[j].DisplayName
		} else {
			right = memberList[j].Email
		}
		return left < right
	})

	c.JSON(http.StatusOK, gin.H{
		"class_id":   class.ID,
		"class_name": class.Name,
		"class_info": gin.H{
			"name":          class.Name,
			"student_count": class.StudentCount,
			"attend_count":  class.AttendCount,
		},
		"members":  memberList,
		"progress": progress,
	})
}

func buildClassProgress(team *models.TeamWithSport, matches []*models.MatchDetail) models.ClassProgress {
	entry := models.ClassProgress{
		SportName:      team.SportName,
		TeamName:       team.Name,
		TournamentName: "",
		Status:         "試合情報なし",
		CurrentRound:   "-",
	}

	if len(matches) == 0 {
		return entry
	}

	entry.TournamentName = matches[0].TournamentName

	maxRound := matches[0].MaxRound

	var lastMatch *models.MatchDetail
	var nextMatch *models.MatchDetail

	for _, m := range matches {
		if m.WinnerTeamID.Valid {
			if lastMatch == nil || m.Round > lastMatch.Round || (m.Round == lastMatch.Round && m.MatchNumber > lastMatch.MatchNumber) {
				copy := *m
				lastMatch = &copy
			}
		} else {
			if nextMatch == nil {
				copy := *m
				nextMatch = &copy
			}
		}
	}

	if nextMatch != nil {
		entry.Status = stageLabel(nextMatch.Round, maxRound, nextMatch.IsBronzeMatch, nextMatch.NextMatchID.Valid) + "進出"
		entry.CurrentRound = stageLabel(nextMatch.Round, maxRound, nextMatch.IsBronzeMatch, nextMatch.NextMatchID.Valid)
		entry.NextMatch = buildProgressMatch(nextMatch, team.ID, "予定", maxRound)
		if nextMatch.StartTime.Valid {
			start := nextMatch.StartTime.String
			entry.NextMatch.StartTime = &start
		}
		entry.NextMatch.MatchStatus = nextMatch.Status
	}

	if lastMatch != nil {
		result := "敗戦"
		if lastMatch.WinnerTeamID.Valid && int(lastMatch.WinnerTeamID.Int64) == team.ID {
			result = "勝利"
		}
		entry.LastMatch = buildProgressMatch(lastMatch, team.ID, result, maxRound)
		score := buildScoreString(lastMatch, team.ID)
		if score != nil {
			entry.LastMatch.Score = score
		}
		entry.LastMatch.MatchStatus = "終了"
	}

	if nextMatch == nil {
		if lastMatch != nil {
			if lastMatch.WinnerTeamID.Valid && int(lastMatch.WinnerTeamID.Int64) == team.ID {
				if !lastMatch.NextMatchID.Valid && !lastMatch.IsBronzeMatch {
					entry.Status = "優勝"
					entry.CurrentRound = "決勝"
				} else if lastMatch.IsBronzeMatch {
					entry.Status = "3位決定戦勝利"
					entry.CurrentRound = "3位決定戦"
				} else {
					entry.Status = stageLabel(lastMatch.Round, maxRound, lastMatch.IsBronzeMatch, lastMatch.NextMatchID.Valid) + "勝利"
					entry.CurrentRound = stageLabel(lastMatch.Round, maxRound, lastMatch.IsBronzeMatch, lastMatch.NextMatchID.Valid)
				}
			} else {
				entry.Status = stageLabel(lastMatch.Round, maxRound, lastMatch.IsBronzeMatch, lastMatch.NextMatchID.Valid) + "敗退"
				entry.CurrentRound = stageLabel(lastMatch.Round, maxRound, lastMatch.IsBronzeMatch, lastMatch.NextMatchID.Valid)
			}
		}
	} else {
		if lastMatch != nil && lastMatch.WinnerTeamID.Valid && int(lastMatch.WinnerTeamID.Int64) == team.ID {
			entry.Status = stageLabel(nextMatch.Round, maxRound, nextMatch.IsBronzeMatch, nextMatch.NextMatchID.Valid) + "進出"
		} else if lastMatch == nil {
			entry.Status = stageLabel(nextMatch.Round, maxRound, nextMatch.IsBronzeMatch, nextMatch.NextMatchID.Valid) + "予定"
		}
	}

	return entry
}

func buildProgressMatch(detail *models.MatchDetail, teamID int, result string, maxRound int) *models.ClassProgressMatch {
	match := &models.ClassProgressMatch{
		MatchID:    detail.MatchID,
		Round:      detail.Round,
		RoundLabel: stageLabel(detail.Round, maxRound, detail.IsBronzeMatch, detail.NextMatchID.Valid),
		Result:     result,
	}

	if detail.Team1ID.Valid && int(detail.Team1ID.Int64) == teamID {
		if detail.Team2Name.Valid {
			match.OpponentName = detail.Team2Name.String
		} else if detail.Team2ID.Valid {
			match.OpponentName = fmt.Sprintf("チーム%d", detail.Team2ID.Int64)
		} else {
			match.OpponentName = "未定"
		}
	} else if detail.Team2ID.Valid && int(detail.Team2ID.Int64) == teamID {
		if detail.Team1Name.Valid {
			match.OpponentName = detail.Team1Name.String
		} else if detail.Team1ID.Valid {
			match.OpponentName = fmt.Sprintf("チーム%d", detail.Team1ID.Int64)
		} else {
			match.OpponentName = "未定"
		}
	} else {
		match.OpponentName = "未定"
	}

	return match
}

func buildScoreString(detail *models.MatchDetail, teamID int) *string {
	if !detail.Team1Score.Valid || !detail.Team2Score.Valid {
		return nil
	}
	var teamScore int32
	var opponentScore int32
	if detail.Team1ID.Valid && int(detail.Team1ID.Int64) == teamID {
		teamScore = detail.Team1Score.Int32
		opponentScore = detail.Team2Score.Int32
	} else {
		teamScore = detail.Team2Score.Int32
		opponentScore = detail.Team1Score.Int32
	}
	score := fmt.Sprintf("%d - %d", teamScore, opponentScore)
	return &score
}

func stageLabel(round, maxRound int, isBronze bool, hasNext bool) string {
	if isBronze {
		return "3位決定戦"
	}
	if round == maxRound && !hasNext {
		return "決勝"
	}
	if round == maxRound-1 {
		return "準決勝"
	}
	if round == maxRound-2 {
		return "準々決勝"
	}
	return fmt.Sprintf("%d回戦", round+1)
}
