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

// CreateYearRelayRun は学年対抗リレー(テンプレート)の run を作成し、A/B/総合ボーナス用の試合を生成します。
// 前提: 既に event に対する noon_game_session が存在していること（root が昼競技管理で作成済み）。
func (h *NoonGameHandler) CreateYearRelayRun(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id"})
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

	session, err := h.noonRepo.GetSessionByEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昼競技セッションが未作成です。rootで先にセッションを作成してください。"})
		return
	}

	// 6チーム固定（専教含む）
	teamDefs := []struct {
		Display    string
		ClassNames []string
	}{
		{Display: "1年生", ClassNames: []string{"1-1", "1-2", "1-3"}},
		{Display: "2年生", ClassNames: []string{"IS2", "IT2", "IE2"}},
		{Display: "3年生", ClassNames: []string{"IS3", "IT3", "IE3"}},
		{Display: "4年生", ClassNames: []string{"IS4", "IT4", "IE4"}},
		{Display: "5年生", ClassNames: []string{"IS5", "IT5", "IE5"}},
		{Display: "専教", ClassNames: []string{"専教"}},
	}

	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes", "details": err.Error()})
		return
	}
	classIDByName := make(map[string]int, len(classes))
	for _, cls := range classes {
		classIDByName[cls.Name] = cls.ID
	}

	groupIDs := make([]int, 0, len(teamDefs))
	entryDisplays := make([]string, 0, len(teamDefs))
	for _, def := range teamDefs {
		memberIDs := make([]int, 0, len(def.ClassNames))
		for _, name := range def.ClassNames {
			id, ok := classIDByName[name]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("必要なクラスが見つかりません: %s", name)})
				return
			}
			memberIDs = append(memberIDs, id)
		}
		grp, err := h.noonRepo.SaveGroup(&models.NoonGameGroup{
			SessionID:   session.ID,
			Name:        def.Display,
			Description: nil,
		}, memberIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create year relay group", "details": err.Error()})
			return
		}
		groupIDs = append(groupIDs, grp.ID)
		entryDisplays = append(entryDisplays, def.Display)
	}

	createMatchWithEntries := func(title string) (*models.NoonGameMatchWithResult, error) {
		m := &models.NoonGameMatch{
			SessionID: session.ID,
			Title:     &title,
			Status:    "scheduled",
			AllowDraw: false,
			Entries:   []*models.NoonGameMatchEntry{},
		}
		for i, groupID := range groupIDs {
			display := entryDisplays[i]
			displayCopy := display
			gid := groupID
			m.Entries = append(m.Entries, &models.NoonGameMatchEntry{
				SideType:    "group",
				GroupID:     &gid,
				DisplayName: &displayCopy,
			})
		}
		saved, err := h.noonRepo.SaveMatch(m)
		if err != nil {
			return nil, err
		}
		return h.noonRepo.GetMatchByID(saved.ID)
	}

	matchA, err := createMatchWithEntries("学年対抗リレー Aブロック")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create A block match", "details": err.Error()})
		return
	}
	matchB, err := createMatchWithEntries("学年対抗リレー Bブロック")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create B block match", "details": err.Error()})
		return
	}
	matchBonus, err := createMatchWithEntries("学年対抗リレー 総合ボーナス")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create bonus match", "details": err.Error()})
		return
	}

	run, err := h.noonRepo.CreateTemplateRun(&models.NoonGameTemplateRun{
		SessionID:   session.ID,
		TemplateKey: noonTemplateYearRelay,
		Name:        fmt.Sprintf("学年対抗リレー (event_id=%d)", eventID),
		CreatedBy:   user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template run", "details": err.Error()})
		return
	}

	if _, err := h.noonRepo.LinkTemplateRunMatch(run.ID, matchA.ID, yearRelayMatchA); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link match A", "details": err.Error()})
		return
	}
	if _, err := h.noonRepo.LinkTemplateRunMatch(run.ID, matchB.ID, yearRelayMatchB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link match B", "details": err.Error()})
		return
	}
	if _, err := h.noonRepo.LinkTemplateRunMatch(run.ID, matchBonus.ID, yearRelayMatchBonus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link match bonus", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createYearRelayRunResponse{
		Run: run,
		Matches: map[string]*models.NoonGameMatchWithResult{
			yearRelayMatchA:     matchA,
			yearRelayMatchB:     matchB,
			yearRelayMatchBonus: matchBonus,
		},
	})
}

// CreateCourseRelayRun はコース対抗リレー(テンプレート)の run を作成し、4チームの試合を生成します。
// 前提: 既に event に対する noon_game_session が存在していること（root が昼競技管理で作成済み）。
func (h *NoonGameHandler) CreateCourseRelayRun(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id"})
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

	session, err := h.noonRepo.GetSessionByEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昼競技セッションが未作成です。rootで先にセッションを作成してください。"})
		return
	}

	// 全クラスを取得
	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes", "details": err.Error()})
		return
	}
	if len(classes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "クラスが存在しません"})
		return
	}

	// クラス名から分類
	classIDByName := make(map[string]int, len(classes))
	for _, cls := range classes {
		classIDByName[cls.Name] = cls.ID
	}

	// チーム定義（デフォルト）
	teamDefs := []struct {
		Display    string
		ClassNames []string
	}{
		{Display: "1-1 & IEコース", ClassNames: []string{"1-1", "IE2", "IE3", "IE4", "IE5"}},
		{Display: "1-2 & ISコース", ClassNames: []string{"1-2", "IS2", "IS3", "IS4", "IS5"}},
		{Display: "1-3 & ITコース", ClassNames: []string{"1-3", "IT2", "IT3", "IT4", "IT5"}},
		{Display: "専攻科・教員", ClassNames: []string{"専教"}},
	}

	groupIDs := make([]int, 0, len(teamDefs))
	entryDisplays := make([]string, 0, len(teamDefs))
	for _, def := range teamDefs {
		memberIDs := make([]int, 0, len(def.ClassNames))
		for _, name := range def.ClassNames {
			id, ok := classIDByName[name]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("必要なクラスが見つかりません: %s", name)})
				return
			}
			memberIDs = append(memberIDs, id)
		}
		grp, err := h.noonRepo.SaveGroup(&models.NoonGameGroup{
			SessionID:   session.ID,
			Name:        def.Display,
			Description: nil,
		}, memberIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create course relay group", "details": err.Error()})
			return
		}
		groupIDs = append(groupIDs, grp.ID)
		entryDisplays = append(entryDisplays, def.Display)
	}

	// 試合作成（4チーム）
	title := "コース対抗リレー"
	m := &models.NoonGameMatch{
		SessionID: session.ID,
		Title:     &title,
		Status:    "scheduled",
		AllowDraw: false,
		Entries:   []*models.NoonGameMatchEntry{},
	}
	for i, groupID := range groupIDs {
		display := entryDisplays[i]
		displayCopy := display
		gid := groupID
		m.Entries = append(m.Entries, &models.NoonGameMatchEntry{
			SideType:    "group",
			GroupID:     &gid,
			DisplayName: &displayCopy,
		})
	}
	saved, err := h.noonRepo.SaveMatch(m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create match", "details": err.Error()})
		return
	}
	match, err := h.noonRepo.GetMatchByID(saved.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match", "details": err.Error()})
		return
	}

	run, err := h.noonRepo.CreateTemplateRun(&models.NoonGameTemplateRun{
		SessionID:   session.ID,
		TemplateKey: noonTemplateCourseRelay,
		Name:        fmt.Sprintf("コース対抗リレー (event_id=%d)", eventID),
		CreatedBy:   user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template run", "details": err.Error()})
		return
	}

	if _, err := h.noonRepo.LinkTemplateRunMatch(run.ID, match.ID, courseRelayMatchMain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link match", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createCourseRelayRunResponse{
		Run: run,
		Matches: map[string]*models.NoonGameMatchWithResult{
			courseRelayMatchMain: match,
		},
	})
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

// --- Templates: 学年対抗リレー ---
const (
	noonTemplateYearRelay = "year_relay"
	yearRelayMatchA       = "A"
	yearRelayMatchB       = "B"
	yearRelayMatchBonus   = "BONUS"
)

// --- Templates: コース対抗リレー ---
const (
	noonTemplateCourseRelay = "course_relay"
	courseRelayMatchMain    = "MAIN"
)

type createYearRelayRunResponse struct {
	Run     *models.NoonGameTemplateRun                `json:"run"`
	Matches map[string]*models.NoonGameMatchWithResult `json:"matches"`
}

type createCourseRelayRunResponse struct {
	Run     *models.NoonGameTemplateRun                `json:"run"`
	Matches map[string]*models.NoonGameMatchWithResult `json:"matches"`
}

type yearRelayRankInput struct {
	EntryID int  `json:"entry_id" binding:"required"`
	Rank    int  `json:"rank" binding:"required"`
	Points  *int `json:"points"` // 同順位(同点)の場合のみ必須
}

type yearRelayBlockResultRequest struct {
	Rankings []yearRelayRankInput `json:"rankings" binding:"required"`
	Note     *string              `json:"note"`
}

type yearRelayOverallResultRequest = yearRelayBlockResultRequest

type courseRelayRankInput struct {
	EntryID int  `json:"entry_id" binding:"required"`
	Rank    int  `json:"rank" binding:"required"`
	Points  *int `json:"points"` // 同順位(同点)の場合のみ必須
}

type courseRelayResultRequest struct {
	Rankings     []courseRelayRankInput `json:"rankings" binding:"required"`
	Note         *string                `json:"note"`
	PointsByRank *map[int]int           `json:"points_by_rank"` // オプション: 点数設定を変更可能
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

// RecordYearRelayBlockResult は学年対抗リレーのA/Bブロック順位を登録し、テンプレ点を自動付与します。
// 同順位がある場合は、該当順位の points を必須入力とします。
func (h *NoonGameHandler) RecordYearRelayBlockResult(c *gin.Context) {
	runID, err := strconv.Atoi(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run_id"})
		return
	}
	block := strings.ToUpper(strings.TrimSpace(c.Param("block")))
	if block != yearRelayMatchA && block != yearRelayMatchB {
		c.JSON(http.StatusBadRequest, gin.H{"error": "block must be A or B"})
		return
	}

	var req yearRelayBlockResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	pointsByRank := map[int]int{1: 30, 2: 25, 3: 20, 4: 15, 5: 10, 6: 5}
	h.applyYearRelayRankingsToMatch(c, runID, block, req, pointsByRank)
}

// RecordYearRelayOverallBonus は学年対抗リレーの総合ボーナスを登録します。
// 原則は 1位30/2位20/3位10 を自動付与しますが、同順位がある場合は points 必須入力です。
func (h *NoonGameHandler) RecordYearRelayOverallBonus(c *gin.Context) {
	runID, err := strconv.Atoi(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run_id"})
		return
	}

	var req yearRelayOverallResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 4位以下はボーナス無し
	pointsByRank := map[int]int{1: 30, 2: 20, 3: 10, 4: 0, 5: 0, 6: 0}
	h.applyYearRelayRankingsToMatch(c, runID, yearRelayMatchBonus, req, pointsByRank)
}

// RecordCourseRelayResult はコース対抗リレーの順位を登録し、テンプレ点を自動付与します。
// 点数はオプションで変更可能。同順位がある場合は、該当順位の points を必須入力とします。
func (h *NoonGameHandler) RecordCourseRelayResult(c *gin.Context) {
	runID, err := strconv.Atoi(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run_id"})
		return
	}

	var req courseRelayResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// デフォルトの点数設定（1位40点、2位30点、3位20点、4位10点）
	pointsByRank := map[int]int{1: 40, 2: 30, 3: 20, 4: 10}
	// オプションで点数設定が指定されている場合は上書き
	if req.PointsByRank != nil {
		pointsByRank = *req.PointsByRank
	}

	// yearRelayBlockResultRequestに変換
	rankings := make([]yearRelayRankInput, 0, len(req.Rankings))
	for _, r := range req.Rankings {
		rankings = append(rankings, yearRelayRankInput{
			EntryID: r.EntryID,
			Rank:    r.Rank,
			Points:  r.Points,
		})
	}
	yearRelayReq := yearRelayBlockResultRequest{
		Rankings: rankings,
		Note:     req.Note,
	}

	h.applyCourseRelayRankingsToMatch(c, runID, courseRelayMatchMain, yearRelayReq, pointsByRank)
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

func (h *NoonGameHandler) applyYearRelayRankingsToMatch(
	c *gin.Context,
	runID int,
	matchKey string,
	req yearRelayBlockResultRequest,
	pointsByRank map[int]int,
) {
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

	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch run", "details": err.Error()})
		return
	}
	if run == nil || run.TemplateKey != noonTemplateYearRelay {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	link, err := h.noonRepo.GetTemplateRunMatchByKey(runID, matchKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch run match", "details": err.Error()})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run match not found"})
		return
	}

	match, err := h.noonRepo.GetMatchByID(link.MatchID)
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

	// 入力の検証
	if len(req.Rankings) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings is required"})
		return
	}

	entryLookup := make(map[int]*models.NoonGameMatchEntry)
	for _, e := range match.Entries {
		if e != nil {
			entryLookup[e.ID] = e
		}
	}
	if len(entryLookup) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match entries are not configured"})
		return
	}
	if len(req.Rankings) != len(entryLookup) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings must include all teams"})
		return
	}

	seenEntry := make(map[int]bool)
	rankToCount := make(map[int]int)
	for _, r := range req.Rankings {
		if _, ok := entryLookup[r.EntryID]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rankings contains unknown entry_id"})
			return
		}
		if r.Rank <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rank must be positive"})
			return
		}
		if seenEntry[r.EntryID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate entry_id in rankings"})
			return
		}
		seenEntry[r.EntryID] = true
		rankToCount[r.Rank]++
	}
	for id := range entryLookup {
		if !seenEntry[id] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rankings must include all entries"})
			return
		}
	}

	// 同順位がある rank は points が必須
	for _, r := range req.Rankings {
		if rankToCount[r.Rank] > 1 {
			if r.Points == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "同順位があるため points の指定が必須です"})
				return
			}
		} else {
			// 同順位でない場合は points を受け取らない（テンプレ自動計算）
			if r.Points != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "同順位でない順位には points を指定できません"})
				return
			}
		}
	}

	// 既存ポイントを消す
	if err := h.noonRepo.ClearPointsForMatch(match.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear existing points", "details": err.Error()})
		return
	}

	pointsEntries := make([]*models.NoonGamePoint, 0)
	resultDetails := make([]*models.NoonGameResultDetail, 0)

	matchTitle := fmt.Sprintf("試合 #%d", match.ID)
	if match.Title != nil && strings.TrimSpace(*match.Title) != "" {
		matchTitle = strings.TrimSpace(*match.Title)
	}

	// 最良rank(=1)を winner 扱い（複数なら draw）
	bestRank := math.MaxInt
	bestCount := 0
	bestEntryID := 0

	for _, r := range req.Rankings {
		entry := entryLookup[r.EntryID]
		classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(classIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "対象クラスが解決できません"})
			return
		}

		points := 0
		if rankToCount[r.Rank] > 1 {
			points = *r.Points
		} else {
			p, ok := pointsByRank[r.Rank]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "rank is out of range"})
				return
			}
			points = p
		}

		reason := fmt.Sprintf("昼競技テンプレ(%s) (%s)", run.Name, matchTitle)
		for _, classID := range classIDs {
			reasonCopy := reason
			mid := match.ID
			pointsEntries = append(pointsEntries, &models.NoonGamePoint{
				SessionID: session.ID,
				MatchID:   &mid,
				ClassID:   classID,
				Points:    points,
				Reason:    &reasonCopy,
				Source:    "result",
				CreatedBy: user.ID,
			})
		}

		rankCopy := r.Rank
		resultDetails = append(resultDetails, &models.NoonGameResultDetail{
			EntryID: r.EntryID,
			Rank:    &rankCopy,
			Points:  points,
		})

		if r.Rank < bestRank {
			bestRank = r.Rank
			bestCount = 1
			bestEntryID = r.EntryID
		} else if r.Rank == bestRank {
			bestCount++
		}
	}

	if err := h.noonRepo.InsertPoints(pointsEntries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store points", "details": err.Error()})
		return
	}

	winner := "draw"
	if bestCount == 1 && bestEntryID != 0 && len(match.Entries) >= 2 {
		// entry_index 0/1 を home/away に割り当てている前提
		// それ以外(3人以上)では winner は draw のまま
		if match.Entries[0] != nil && match.Entries[0].ID == bestEntryID {
			winner = "home"
		} else if match.Entries[1] != nil && match.Entries[1].ID == bestEntryID {
			winner = "away"
		}
	}

	if _, err := h.noonRepo.SaveResult(&models.NoonGameResult{
		MatchID:    match.ID,
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

	// クラス得点へ反映
	summary, err := h.noonRepo.SumPointsByClass(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to aggregate points", "details": err.Error()})
		return
	}
	if err := h.classRepo.SetNoonGamePoints(session.EventID, summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update class scores", "details": err.Error()})
		return
	}

	full, err := h.noonRepo.GetMatchByID(match.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match", "details": err.Error()})
		return
	}
	if full == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match not found after update"})
		return
	}
	if err := h.decorateMatches([]*models.NoonGameMatchWithResult{full}, nil, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enrich match data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"match": full})
}

func (h *NoonGameHandler) applyCourseRelayRankingsToMatch(
	c *gin.Context,
	runID int,
	matchKey string,
	req yearRelayBlockResultRequest,
	pointsByRank map[int]int,
) {
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

	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch run", "details": err.Error()})
		return
	}
	if run == nil || run.TemplateKey != noonTemplateCourseRelay {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	link, err := h.noonRepo.GetTemplateRunMatchByKey(runID, matchKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch run match", "details": err.Error()})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run match not found"})
		return
	}

	match, err := h.noonRepo.GetMatchByID(link.MatchID)
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

	// 入力の検証
	if len(req.Rankings) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings is required"})
		return
	}

	entryLookup := make(map[int]*models.NoonGameMatchEntry)
	for _, e := range match.Entries {
		if e != nil {
			entryLookup[e.ID] = e
		}
	}
	if len(entryLookup) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match entries are not configured"})
		return
	}
	if len(req.Rankings) != len(entryLookup) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings must include all teams"})
		return
	}

	seenEntry := make(map[int]bool)
	rankToCount := make(map[int]int)
	for _, r := range req.Rankings {
		if _, ok := entryLookup[r.EntryID]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rankings contains unknown entry_id"})
			return
		}
		if r.Rank <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rank must be positive"})
			return
		}
		if seenEntry[r.EntryID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate entry_id in rankings"})
			return
		}
		seenEntry[r.EntryID] = true
		rankToCount[r.Rank]++
	}
	for id := range entryLookup {
		if !seenEntry[id] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rankings must include all entries"})
			return
		}
	}

	// 同順位がある rank は points が必須
	for _, r := range req.Rankings {
		if rankToCount[r.Rank] > 1 {
			if r.Points == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "同順位があるため points の指定が必須です"})
				return
			}
		} else {
			// 同順位でない場合は points を受け取らない（テンプレ自動計算）
			if r.Points != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "同順位でない順位には points を指定できません"})
				return
			}
		}
	}

	// 既存ポイントを消す
	if err := h.noonRepo.ClearPointsForMatch(match.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear existing points", "details": err.Error()})
		return
	}

	pointsEntries := make([]*models.NoonGamePoint, 0)
	resultDetails := make([]*models.NoonGameResultDetail, 0)

	matchTitle := fmt.Sprintf("試合 #%d", match.ID)
	if match.Title != nil && strings.TrimSpace(*match.Title) != "" {
		matchTitle = strings.TrimSpace(*match.Title)
	}

	// 最良rank(=1)を winner 扱い（複数なら draw）
	bestRank := math.MaxInt
	bestCount := 0
	bestEntryID := 0

	for _, r := range req.Rankings {
		entry := entryLookup[r.EntryID]
		classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(classIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "対象クラスが解決できません"})
			return
		}

		points := 0
		if rankToCount[r.Rank] > 1 {
			points = *r.Points
		} else {
			p, ok := pointsByRank[r.Rank]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "rank is out of range"})
				return
			}
			points = p
		}

		reason := fmt.Sprintf("昼競技テンプレ(%s) (%s)", run.Name, matchTitle)
		for _, classID := range classIDs {
			reasonCopy := reason
			mid := match.ID
			pointsEntries = append(pointsEntries, &models.NoonGamePoint{
				SessionID: session.ID,
				MatchID:   &mid,
				ClassID:   classID,
				Points:    points,
				Reason:    &reasonCopy,
				Source:    "result",
				CreatedBy: user.ID,
			})
		}

		rankCopy := r.Rank
		resultDetails = append(resultDetails, &models.NoonGameResultDetail{
			EntryID: r.EntryID,
			Rank:    &rankCopy,
			Points:  points,
		})

		if r.Rank < bestRank {
			bestRank = r.Rank
			bestCount = 1
			bestEntryID = r.EntryID
		} else if r.Rank == bestRank {
			bestCount++
		}
	}

	if err := h.noonRepo.InsertPoints(pointsEntries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store points", "details": err.Error()})
		return
	}

	winner := "draw"
	if bestCount == 1 && bestEntryID != 0 && len(match.Entries) >= 2 {
		// entry_index 0/1 を home/away に割り当てている前提
		// それ以外(3人以上)では winner は draw のまま
		if match.Entries[0] != nil && match.Entries[0].ID == bestEntryID {
			winner = "home"
		} else if match.Entries[1] != nil && match.Entries[1].ID == bestEntryID {
			winner = "away"
		}
	}

	if _, err := h.noonRepo.SaveResult(&models.NoonGameResult{
		MatchID:    match.ID,
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

	// クラス得点へ反映
	summary, err := h.noonRepo.SumPointsByClass(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to aggregate points", "details": err.Error()})
		return
	}
	if err := h.classRepo.SetNoonGamePoints(session.EventID, summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update class scores", "details": err.Error()})
		return
	}

	full, err := h.noonRepo.GetMatchByID(match.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match", "details": err.Error()})
		return
	}
	if full == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match not found after update"})
		return
	}
	if err := h.decorateMatches([]*models.NoonGameMatchWithResult{full}, nil, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enrich match data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"match": full})
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
