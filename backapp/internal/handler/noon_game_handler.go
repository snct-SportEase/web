package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type NoonGameHandler struct {
	noonRepo  repository.NoonGameRepository
	classRepo repository.ClassRepository
	eventRepo repository.EventRepository
}

func NewNoonGameHandler(noonRepo repository.NoonGameRepository, classRepo repository.ClassRepository, eventRepo repository.EventRepository) *NoonGameHandler {
	return &NoonGameHandler{
		noonRepo:  noonRepo,
		classRepo: classRepo,
		eventRepo: eventRepo,
	}
}

type upsertNoonSessionRequest struct {
	Name                string  `json:"name" binding:"required"`
	Description         *string `json:"description"`
	Mode                string  `json:"mode"`
	WinPoints           int     `json:"win_points"`
	LossPoints          int     `json:"loss_points"`
	DrawPoints          int     `json:"draw_points"`
	ParticipationPoints int     `json:"participation_points"`
	AllowManualPoints   *bool   `json:"allow_manual_points"`
}

type upsertNoonGroupRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	ClassIDs    []int   `json:"class_ids"`
}

type upsertNoonMatchRequest struct {
	Title       *string `json:"title"`
	ScheduledAt *string `json:"scheduled_at"`
	Location    *string `json:"location"`
	Format      *string `json:"format"`
	Memo        *string `json:"memo"`
	Status      string  `json:"status"`
	AllowDraw   bool    `json:"allow_draw"`
	HomeSide    struct {
		Type    string `json:"type" binding:"required"` // class or group
		ClassID *int   `json:"class_id"`
		GroupID *int   `json:"group_id"`
	} `json:"home_side" binding:"required"`
	AwaySide struct {
		Type    string `json:"type" binding:"required"`
		ClassID *int   `json:"class_id"`
		GroupID *int   `json:"group_id"`
	} `json:"away_side" binding:"required"`
	Participants []struct {
		ID          *int    `json:"id"`
		Type        string  `json:"type"`
		ClassID     *int    `json:"class_id"`
		GroupID     *int    `json:"group_id"`
		DisplayName *string `json:"display_name"`
	} `json:"participants"`
}

type recordNoonMatchResultRequest struct {
	Winner   string  `json:"winner"` // home, away, draw
	Note     *string `json:"note"`
	Rankings []struct {
		EntryID int     `json:"entry_id"`
		Rank    *int    `json:"rank"`
		Points  int     `json:"points"`
		Note    *string `json:"note"`
	} `json:"rankings"`
}

type manualPointRequest struct {
	ClassID int     `json:"class_id" binding:"required"`
	Points  int     `json:"points" binding:"required"`
	Reason  *string `json:"reason"`
}

func (h *NoonGameHandler) GetSession(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id"})
		return
	}

	if err := h.ensureEventExists(eventID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, errNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	session, err := h.noonRepo.GetSessionByEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve noon game session", "details": err.Error()})
		return
	}

	if session == nil {
		classes, clsErr := h.classRepo.GetAllClasses(eventID)
		if clsErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes", "details": clsErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"session":        nil,
			"groups":         []interface{}{},
			"matches":        []interface{}{},
			"classes":        classes,
			"points_summary": []interface{}{},
		})
		return
	}

	payload, err := h.buildSessionPayload(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build session payload", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payload)
}

func (h *NoonGameHandler) UpsertSession(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id"})
		return
	}

	if err := h.ensureEventExists(eventID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, errNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	var req upsertNoonSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	session := &models.NoonGameSession{
		EventID:             eventID,
		Name:                req.Name,
		Description:         req.Description,
		Mode:                strings.ToLower(strings.TrimSpace(req.Mode)),
		WinPoints:           req.WinPoints,
		LossPoints:          req.LossPoints,
		DrawPoints:          req.DrawPoints,
		ParticipationPoints: req.ParticipationPoints,
		AllowManualPoints:   true,
	}
	if session.Mode == "" {
		session.Mode = "mixed"
	}
	if req.AllowManualPoints != nil {
		session.AllowManualPoints = *req.AllowManualPoints
	}

	updated, err := h.noonRepo.UpsertSession(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save noon game session", "details": err.Error()})
		return
	}

	payload, err := h.buildSessionPayload(updated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build session payload", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payload)
}

func (h *NoonGameHandler) SaveGroup(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}

	session, err := h.noonRepo.GetSessionByID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve noon game session", "details": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Noon game session not found"})
		return
	}

	groupID := 0
	if groupParam := c.Param("group_id"); groupParam != "" {
		groupID, err = strconv.Atoi(groupParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id"})
			return
		}
	}

	var req upsertNoonGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	group := &models.NoonGameGroup{
		ID:          groupID,
		SessionID:   sessionID,
		Name:        req.Name,
		Description: req.Description,
	}

	updated, err := h.noonRepo.SaveGroup(group, req.ClassIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save group", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group": updated})
}

func (h *NoonGameHandler) DeleteGroup(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}
	groupID, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id"})
		return
	}

	if err := h.noonRepo.DeleteGroup(sessionID, groupID); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

func (h *NoonGameHandler) SaveMatch(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}

	session, err := h.noonRepo.GetSessionByID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve noon game session", "details": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Noon game session not found"})
		return
	}

	matchID := 0
	if matchParam := c.Param("match_id"); matchParam != "" {
		matchID, err = strconv.Atoi(matchParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match_id"})
			return
		}
	}

	var req upsertNoonMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	copyIntPtr := func(src *int) *int {
		if src == nil {
			return nil
		}
		val := *src
		return &val
	}

	entries := make([]*models.NoonGameMatchEntry, 0)
	if len(req.Participants) > 0 {
		for idx, p := range req.Participants {
			sideType := strings.ToLower(strings.TrimSpace(p.Type))
			if sideType == "" {
				sideType = "class"
			}
			entry := &models.NoonGameMatchEntry{
				SideType: sideType,
			}
			switch sideType {
			case "class":
				if p.ClassID == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("participants[%d]: class_id is required", idx)})
					return
				}
				entry.ClassID = copyIntPtr(p.ClassID)
				entry.GroupID = nil
			case "group":
				if p.GroupID == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("participants[%d]: group_id is required", idx)})
					return
				}
				entry.GroupID = copyIntPtr(p.GroupID)
				entry.ClassID = nil
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("participants[%d]: type must be 'class' or 'group'", idx)})
				return
			}
			if p.DisplayName != nil {
				name := strings.TrimSpace(*p.DisplayName)
				if name != "" {
					entry.DisplayName = &name
				}
			}
			entries = append(entries, entry)
		}
	} else {
		if err := validateSide(req.HomeSide); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("home_side is invalid: %s", err.Error())})
			return
		}
		if err := validateSide(req.AwaySide); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("away_side is invalid: %s", err.Error())})
			return
		}
		homeEntry := &models.NoonGameMatchEntry{
			SideType: strings.ToLower(strings.TrimSpace(req.HomeSide.Type)),
			ClassID:  copyIntPtr(req.HomeSide.ClassID),
			GroupID:  copyIntPtr(req.HomeSide.GroupID),
		}
		if homeEntry.SideType == "group" {
			homeEntry.ClassID = nil
		} else {
			homeEntry.GroupID = nil
		}
		entries = append(entries, homeEntry)

		awayEntry := &models.NoonGameMatchEntry{
			SideType: strings.ToLower(strings.TrimSpace(req.AwaySide.Type)),
			ClassID:  copyIntPtr(req.AwaySide.ClassID),
			GroupID:  copyIntPtr(req.AwaySide.GroupID),
		}
		if awayEntry.SideType == "group" {
			awayEntry.ClassID = nil
		} else {
			awayEntry.GroupID = nil
		}
		entries = append(entries, awayEntry)
	}

	if len(entries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one participant is required"})
		return
	}

	match := &models.NoonGameMatch{
		ID:        matchID,
		SessionID: sessionID,
		Title:     req.Title,
		Location:  req.Location,
		Format:    req.Format,
		Memo:      req.Memo,
		Status:    strings.ToLower(strings.TrimSpace(req.Status)),
		AllowDraw: req.AllowDraw,
		Entries:   entries,
	}
	if match.Status == "" {
		match.Status = "scheduled"
	}

	if req.ScheduledAt != nil && strings.TrimSpace(*req.ScheduledAt) != "" {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at format. Use RFC3339"})
			return
		}
		match.ScheduledAt = &t
	}

	updated, err := h.noonRepo.SaveMatch(match)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save match", "details": err.Error()})
		return
	}

	full, err := h.noonRepo.GetMatchByID(updated.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve match", "details": err.Error()})
		return
	}
	if full == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match saved but not found"})
		return
	}

	if err := h.decorateMatches([]*models.NoonGameMatchWithResult{full}, nil, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enrich match data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"match": full})
}

func (h *NoonGameHandler) DeleteMatch(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match_id"})
		return
	}

	if err := h.noonRepo.DeleteMatch(sessionID, matchID); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "match deleted"})
}

func (h *NoonGameHandler) RecordMatchResult(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match_id"})
		return
	}

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	}

	match, err := h.noonRepo.GetMatchByID(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match", "details": err.Error()})
		return
	}
	if match == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	session, err := h.noonRepo.GetSessionByID(match.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session not found for match"})
		return
	}

	// 雨天時モードのチェック: 昼競技をブロック
	event, err := h.eventRepo.GetEventByID(session.EventID)
	if err == nil && event != nil && event.IsRainyMode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "雨天時モードでは、昼競技の試合結果を記録できません"})
		return
	}

	var req recordNoonMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	useRankings := len(req.Rankings) > 0
	entryOrder := make(map[int]int)
	entryLookup := make(map[int]*models.NoonGameMatchEntry)
	for idx, entry := range match.Entries {
		if entry == nil {
			continue
		}
		entryOrder[entry.ID] = idx
		entryLookup[entry.ID] = entry
	}

	if useRankings && len(entryLookup) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参加者が設定されていない試合には順位入力できません"})
		return
	}

	if err := h.noonRepo.ClearPointsForMatch(matchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear existing points", "details": err.Error()})
		return
	}

	pointsEntries := make([]*models.NoonGamePoint, 0)
	resultDetails := make([]*models.NoonGameResultDetail, 0)
	matchTitle := fmt.Sprintf("試合 #%d", match.ID)
	if match.Title != nil && strings.TrimSpace(*match.Title) != "" {
		matchTitle = strings.TrimSpace(*match.Title)
	}

	winner := "draw"

	if useRankings {
		bestEntryID := 0
		bestRank := math.MaxInt
		bestPoints := math.MinInt
		tie := false

		for idx, ranking := range req.Rankings {
			entry := entryLookup[ranking.EntryID]
			if entry == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("rankings[%d]: 指定された参加者が存在しません", idx)})
				return
			}

			classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("rankings[%d]: %s", idx, err.Error())})
				return
			}
			if len(classIDs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("rankings[%d]: 該当するクラスが見つかりません", idx)})
				return
			}

			reason := fmt.Sprintf("昼競技ポイント (%s)", matchTitle)
			if ranking.Rank != nil {
				reason = fmt.Sprintf("昼競技順位%d位 (%s)", *ranking.Rank, matchTitle)
			}

			for _, classID := range classIDs {
				pointsEntries = append(pointsEntries, &models.NoonGamePoint{
					SessionID: session.ID,
					MatchID:   &matchID,
					ClassID:   classID,
					Points:    ranking.Points,
					Reason:    &reason,
					Source:    "result",
					CreatedBy: user.ID,
				})
			}

			detail := &models.NoonGameResultDetail{
				EntryID: ranking.EntryID,
				Points:  ranking.Points,
			}
			if ranking.Rank != nil {
				rankVal := *ranking.Rank
				detail.Rank = &rankVal
			}
			if ranking.Note != nil {
				note := strings.TrimSpace(*ranking.Note)
				if note != "" {
					detail.Note = &note
				}
			}
			resultDetails = append(resultDetails, detail)

			currentRank := math.MaxInt / 2
			if ranking.Rank != nil {
				currentRank = *ranking.Rank
			}
			if currentRank < bestRank || (currentRank == bestRank && ranking.Points > bestPoints) {
				bestRank = currentRank
				bestPoints = ranking.Points
				bestEntryID = ranking.EntryID
				tie = false
			} else if currentRank == bestRank && ranking.Points == bestPoints {
				tie = true
			}
		}

		if !tie && bestEntryID != 0 {
			if idx, ok := entryOrder[bestEntryID]; ok {
				if idx == 0 {
					winner = "home"
				} else if idx == 1 {
					winner = "away"
				}
			}
		}
	} else {
		winner = strings.ToLower(strings.TrimSpace(req.Winner))
		if winner != "home" && winner != "away" && winner != "draw" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "winner must be one of 'home', 'away', or 'draw'"})
			return
		}
		if winner == "draw" && !match.AllowDraw {
			c.JSON(http.StatusBadRequest, gin.H{"error": "draw is not allowed for this match"})
			return
		}

		homeClassIDs, err := h.resolveClassIDs(match.HomeSideType, match.HomeClassID, match.HomeGroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to resolve home side: %s", err.Error())})
			return
		}
		awayClassIDs, err := h.resolveClassIDs(match.AwaySideType, match.AwayClassID, match.AwayGroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to resolve away side: %s", err.Error())})
			return
		}
		if len(homeClassIDs) == 0 || len(awayClassIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both sides must have at least one class"})
			return
		}

		pointsEntries = h.calculatePointsEntries(session, match, winner, homeClassIDs, awayClassIDs, user)
		resultDetails = []*models.NoonGameResultDetail{}
	}

	if err := h.noonRepo.InsertPoints(pointsEntries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store points", "details": err.Error()})
		return
	}

	if _, err := h.noonRepo.SaveResult(&models.NoonGameResult{
		MatchID:    matchID,
		Winner:     winner,
		RecordedBy: user.ID,
		Note:       req.Note,
		Details:    resultDetails,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store match result", "details": err.Error()})
		return
	}

	match.Status = "completed"
	if _, err := h.noonRepo.SaveMatch(match.NoonGameMatch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match status", "details": err.Error()})
		return
	}

	summary, err := h.noonRepo.SumPointsByClass(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to aggregate points", "details": err.Error()})
		return
	}

	if err := h.classRepo.SetNoonGamePoints(session.EventID, summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update class scores", "details": err.Error()})
		return
	}

	fullMatch, err := h.noonRepo.GetMatchByID(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match", "details": err.Error()})
		return
	}
	if fullMatch == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match not found after update"})
		return
	}

	if err := h.decorateMatches([]*models.NoonGameMatchWithResult{fullMatch}, nil, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enrich match data", "details": err.Error()})
		return
	}

	payload, err := h.buildSessionPayload(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build session payload", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match":          fullMatch,
		"session":        payload["session"],
		"groups":         payload["groups"],
		"matches":        payload["matches"],
		"classes":        payload["classes"],
		"points_summary": payload["points_summary"],
	})
}

func (h *NoonGameHandler) AddManualPoint(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}

	session, err := h.noonRepo.GetSessionByID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !session.AllowManualPoints {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manual points are not allowed for this session"})
		return
	}

	var req manualPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	}

	if _, err := h.noonRepo.InsertPoint(&models.NoonGamePoint{
		SessionID: sessionID,
		ClassID:   req.ClassID,
		Points:    req.Points,
		Reason:    req.Reason,
		Source:    "manual",
		CreatedBy: user.ID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store manual points", "details": err.Error()})
		return
	}

	summary, err := h.noonRepo.SumPointsByClass(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to aggregate points", "details": err.Error()})
		return
	}

	if err := h.classRepo.SetNoonGamePoints(session.EventID, summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update class scores", "details": err.Error()})
		return
	}

	payload, err := h.buildSessionPayload(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build session payload", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payload)
}

func (h *NoonGameHandler) calculatePointsEntries(session *models.NoonGameSession, match *models.NoonGameMatchWithResult, winner string, homeClassIDs, awayClassIDs []int, user *models.User) []*models.NoonGamePoint {
	var entries []*models.NoonGamePoint
	matchID := match.ID
	matchTitle := fmt.Sprintf("試合 #%d", matchID)
	if match.Title != nil && strings.TrimSpace(*match.Title) != "" {
		matchTitle = strings.TrimSpace(*match.Title)
	}

	addEntries := func(classIDs []int, points int, reason string) {
		if points == 0 {
			return
		}
		for _, classID := range classIDs {
			reasonCopy := reason
			point := &models.NoonGamePoint{
				SessionID: session.ID,
				MatchID:   &matchID,
				ClassID:   classID,
				Points:    points,
				Reason:    &reasonCopy,
				Source:    "result",
				CreatedBy: user.ID,
			}
			entries = append(entries, point)
		}
	}

	if session.ParticipationPoints != 0 {
		reason := fmt.Sprintf("昼競技参加 (%s)", matchTitle)
		addEntries(homeClassIDs, session.ParticipationPoints, reason)
		addEntries(awayClassIDs, session.ParticipationPoints, reason)
	}

	switch winner {
	case "home":
		addEntries(homeClassIDs, session.WinPoints, fmt.Sprintf("昼競技勝利 (%s)", matchTitle))
		addEntries(awayClassIDs, session.LossPoints, fmt.Sprintf("昼競技敗北 (%s)", matchTitle))
	case "away":
		addEntries(awayClassIDs, session.WinPoints, fmt.Sprintf("昼競技勝利 (%s)", matchTitle))
		addEntries(homeClassIDs, session.LossPoints, fmt.Sprintf("昼競技敗北 (%s)", matchTitle))
	case "draw":
		addEntries(homeClassIDs, session.DrawPoints, fmt.Sprintf("昼競技引き分け (%s)", matchTitle))
		addEntries(awayClassIDs, session.DrawPoints, fmt.Sprintf("昼競技引き分け (%s)", matchTitle))
	}

	return entries
}

func (h *NoonGameHandler) resolveClassIDs(sideType string, classID, groupID *int) ([]int, error) {
	switch sideType {
	case "class":
		if classID == nil {
			return nil, fmt.Errorf("class_id is required for class side")
		}
		return []int{*classID}, nil
	case "group":
		if groupID == nil {
			return nil, fmt.Errorf("group_id is required for group side")
		}
		members, err := h.noonRepo.GetGroupMembers(*groupID)
		if err != nil {
			return nil, err
		}
		if len(members) == 0 {
			return nil, fmt.Errorf("group has no members")
		}
		classIDs := make([]int, 0, len(members))
		for _, member := range members {
			classIDs = append(classIDs, member.ClassID)
		}
		return classIDs, nil
	default:
		return nil, fmt.Errorf("unknown side type: %s", sideType)
	}
}

func (h *NoonGameHandler) buildSessionPayload(session *models.NoonGameSession) (gin.H, error) {
	classes, err := h.classRepo.GetAllClasses(session.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classes: %w", err)
	}

	classMap := make(map[int]*models.Class)
	for _, class := range classes {
		classMap[class.ID] = class
	}

	groups, err := h.noonRepo.GetGroupsWithMembers(session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groups: %w", err)
	}

	groupMap := make(map[int]*models.NoonGameGroupWithMembers)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	matches, err := h.noonRepo.GetMatchesWithResults(session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch matches: %w", err)
	}

	if err := h.decorateMatches(matches, classMap, groupMap); err != nil {
		return nil, err
	}

	pointsMap, err := h.noonRepo.SumPointsByClass(session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate points: %w", err)
	}

	var summary []*models.NoonGamePointsSummary
	for _, class := range classes {
		points := pointsMap[class.ID]
		summary = append(summary, &models.NoonGamePointsSummary{
			ClassID:   class.ID,
			ClassName: class.Name,
			Points:    points,
		})
	}

	session.Groups = groups
	session.Matches = matches
	session.PointsSummary = summary

	return gin.H{
		"session":        session,
		"groups":         groups,
		"matches":        matches,
		"classes":        classes,
		"points_summary": summary,
	}, nil
}

func (h *NoonGameHandler) decorateMatches(matches []*models.NoonGameMatchWithResult, classMap map[int]*models.Class, groupMap map[int]*models.NoonGameGroupWithMembers) error {
	for _, match := range matches {
		entryLookup := make(map[int]*models.NoonGameMatchEntry)
		entryNameMap := make(map[int]string)

		if len(match.Entries) > 0 {
			for _, entry := range match.Entries {
				if entry == nil {
					continue
				}
				name, classIDs, err := h.resolveEntryDisplay(match.SessionID, entry, classMap, groupMap)
				if err != nil {
					return err
				}
				entry.ResolvedName = name
				entry.ClassIDs = classIDs
				entryLookup[entry.ID] = entry
				entryNameMap[entry.ID] = name
			}

			if len(match.Entries) >= 1 {
				match.HomeDisplayName = match.Entries[0].ResolvedName
				match.HomeClassIDs = match.Entries[0].ClassIDs
			}
			if len(match.Entries) >= 2 {
				match.AwayDisplayName = match.Entries[1].ResolvedName
				match.AwayClassIDs = match.Entries[1].ClassIDs
			}
		}

		if match.HomeDisplayName == "" {
			homeName, homeIDs, err := h.displayNameWithCache(match.SessionID, match.HomeSideType, match.HomeClassID, match.HomeGroupID, classMap, groupMap)
			if err != nil {
				return err
			}
			match.HomeDisplayName = homeName
			match.HomeClassIDs = homeIDs
		}

		if match.AwayDisplayName == "" {
			awayName, awayIDs, err := h.displayNameWithCache(match.SessionID, match.AwaySideType, match.AwayClassID, match.AwayGroupID, classMap, groupMap)
			if err != nil {
				return err
			}
			match.AwayDisplayName = awayName
			match.AwayClassIDs = awayIDs
		}

		if match.Result != nil {
			if len(match.Result.Details) > 0 {
				bestName := ""
				bestRank := math.MaxInt
				bestPoints := math.MinInt
				tie := false

				for _, detail := range match.Result.Details {
					if detail == nil {
						continue
					}
					if name, ok := entryNameMap[detail.EntryID]; ok && name != "" {
						detail.EntryResolvedName = name
					}
					if detail.EntryResolvedName == "" {
						if entry, ok := entryLookup[detail.EntryID]; ok && entry != nil {
							detail.EntryResolvedName = entry.ResolvedName
						}
					}

					currentRank := math.MaxInt / 2
					if detail.Rank != nil {
						currentRank = *detail.Rank
					}

					if detail.EntryResolvedName != "" {
						if currentRank < bestRank || (currentRank == bestRank && detail.Points > bestPoints) {
							bestRank = currentRank
							bestPoints = detail.Points
							bestName = detail.EntryResolvedName
							tie = false
						} else if currentRank == bestRank && detail.Points == bestPoints {
							tie = true
						}
					}
				}

				if bestName != "" && !tie {
					match.WinnerDisplay = &bestName
				} else {
					draw := "引き分け"
					match.WinnerDisplay = &draw
				}
			} else {
				switch match.Result.Winner {
				case "home":
					name := match.HomeDisplayName
					match.WinnerDisplay = &name
				case "away":
					name := match.AwayDisplayName
					match.WinnerDisplay = &name
				case "draw":
					disp := "引き分け"
					match.WinnerDisplay = &disp
				default:
					match.WinnerDisplay = nil
				}
			}
		}
	}
	return nil
}

func (h *NoonGameHandler) resolveEntryDisplay(sessionID int, entry *models.NoonGameMatchEntry, classMap map[int]*models.Class, groupMap map[int]*models.NoonGameGroupWithMembers) (string, []int, error) {
	if entry == nil {
		return "", nil, fmt.Errorf("entry is nil")
	}

	sideType := strings.ToLower(strings.TrimSpace(entry.SideType))
	switch sideType {
	case "class":
		if entry.ClassID == nil {
			return "", nil, fmt.Errorf("class_id is required for class entry")
		}
		class := classMap[*entry.ClassID]
		if class == nil {
			var err error
			class, err = h.classRepo.GetClassByID(*entry.ClassID)
			if err != nil {
				return "", nil, err
			}
			if class == nil {
				return "", nil, fmt.Errorf("class not found")
			}
			if classMap != nil {
				classMap[class.ID] = class
			}
		}
		name := class.Name
		if entry.DisplayName != nil {
			custom := strings.TrimSpace(*entry.DisplayName)
			if custom != "" {
				name = custom
			}
		}
		return name, []int{class.ID}, nil
	case "group":
		if entry.GroupID == nil {
			return "", nil, fmt.Errorf("group_id is required for group entry")
		}
		group := groupMap[*entry.GroupID]
		if group == nil {
			var err error
			group, err = h.noonRepo.GetGroupWithMembers(sessionID, *entry.GroupID)
			if err != nil {
				return "", nil, err
			}
			if group == nil {
				return "", nil, fmt.Errorf("group not found")
			}
			if groupMap != nil {
				groupMap[group.ID] = group
			}
		}
		name := group.Name
		if entry.DisplayName != nil {
			custom := strings.TrimSpace(*entry.DisplayName)
			if custom != "" {
				name = custom
			}
		}
		classIDs := make([]int, 0, len(group.Members))
		for _, member := range group.Members {
			classIDs = append(classIDs, member.ClassID)
		}
		return name, classIDs, nil
	default:
		return "", nil, fmt.Errorf("unknown side type: %s", entry.SideType)
	}
}

func (h *NoonGameHandler) displayNameWithCache(sessionID int, sideType string, classID, groupID *int, classMap map[int]*models.Class, groupMap map[int]*models.NoonGameGroupWithMembers) (string, []int, error) {
	tempEntry := &models.NoonGameMatchEntry{
		SideType: strings.ToLower(strings.TrimSpace(sideType)),
		ClassID:  classID,
		GroupID:  groupID,
	}
	return h.resolveEntryDisplay(sessionID, tempEntry, classMap, groupMap)
}

var errNotFound = errors.New("event not found")

func (h *NoonGameHandler) ensureEventExists(eventID int) error {
	event, err := h.eventRepo.GetEventByID(eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return errNotFound
	}
	return nil
}

func validateSide(side struct {
	Type    string `json:"type" binding:"required"`
	ClassID *int   `json:"class_id"`
	GroupID *int   `json:"group_id"`
}) error {
	sideType := strings.ToLower(strings.TrimSpace(side.Type))
	switch sideType {
	case "class":
		if side.ClassID == nil {
			return fmt.Errorf("class_id is required when type is 'class'")
		}
	case "group":
		if side.GroupID == nil {
			return fmt.Errorf("group_id is required when type is 'group'")
		}
	default:
		return fmt.Errorf("type must be 'class' or 'group'")
	}
	return nil
}
