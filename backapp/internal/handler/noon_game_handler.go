package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"errors"
	"fmt"
	"log"
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

	var req createTemplateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// リクエストボディが空でもエラーにしない（後方互換性のため）
		req = createTemplateRunRequest{}
	}

	session, err := h.noonRepo.GetSessionByEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		// セッションが存在しない場合はリクエストから設定を取得、またはデフォルト値を使用
		sessionName := fmt.Sprintf("学年対抗リレー_%d", eventID)
		if req.Session != nil && req.Session.Name != "" {
			sessionName = req.Session.Name
		}
		mode := "group"
		if req.Session != nil && req.Session.Mode != "" {
			mode = strings.ToLower(strings.TrimSpace(req.Session.Mode))
		}
		session = &models.NoonGameSession{
			EventID:             eventID,
			Name:                sessionName,
			Description:         nil,
			Mode:                mode,
			WinPoints:           0,
			LossPoints:          0,
			DrawPoints:          0,
			ParticipationPoints: 0,
			AllowManualPoints:   false,
		}
		if req.Session != nil {
			if req.Session.Description != nil {
				session.Description = req.Session.Description
			}
			session.WinPoints = req.Session.WinPoints
			session.LossPoints = req.Session.LossPoints
			session.DrawPoints = req.Session.DrawPoints
			session.ParticipationPoints = req.Session.ParticipationPoints
			if req.Session.AllowManualPoints != nil {
				session.AllowManualPoints = *req.Session.AllowManualPoints
			}
		}
		session, err = h.noonRepo.UpsertSession(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session", "details": err.Error()})
			return
		}
	} else {
		// 既にセッションが存在する場合は、そのセッションを使用（1つのイベントにつき1つのセッションのみ）
		// セッション設定の更新は行わない（テンプレート作成時は既存セッションを使用）
	}

	// 既にテンプレートランが存在する場合は、既存のテンプレートと関連データを削除してから新しいテンプレートを作成
	existingRuns, err := h.noonRepo.ListTemplateRunsBySession(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing template runs", "details": err.Error()})
		return
	}
	if len(existingRuns) > 0 {
		// 既存のテンプレートランと関連データ（試合、グループなど）を削除
		if err := h.noonRepo.DeleteTemplateRunAndRelatedData(session.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete existing template run", "details": err.Error()})
			return
		}
	}

	// グループ設定を取得（手動設定があればそれを使用、なければデフォルト設定を使用）
	var teamDefs []struct {
		Display    string
		ClassNames []string
	}

	if req.Session != nil && len(req.Session.Groups) > 0 {
		// 手動設定を使用
		for _, g := range req.Session.Groups {
			// バリデーション: グループ名とクラス名が空でないことを確認
			if strings.TrimSpace(g.GroupName) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "グループ名が空です"})
				return
			}
			if len(g.ClassNames) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("グループ「%s」にクラスが設定されていません", g.GroupName)})
				return
			}
			teamDefs = append(teamDefs, struct {
				Display    string
				ClassNames []string
			}{
				Display:    strings.TrimSpace(g.GroupName),
				ClassNames: g.ClassNames,
			})
		}
	} else {
		// デフォルト設定を使用
		defaultGroups, err := h.noonRepo.GetTemplateDefaultGroups(noonTemplateYearRelay)
		if err != nil {
			log.Printf("WARNING: failed to fetch default groups, using hardcoded defaults: %v", err)
			// フォールバック: ハードコードされたデフォルト値
			teamDefs = []struct {
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
		} else {
			// デフォルト設定をソートして使用
			if len(defaultGroups) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "デフォルトグループ設定が見つかりません。グループを手動で設定してください。"})
				return
			}
			for _, g := range defaultGroups {
				if len(g.ClassNames) == 0 {
					log.Printf("WARNING: default group %q has no class names, skipping", g.GroupName)
					continue
				}
				teamDefs = append(teamDefs, struct {
					Display    string
					ClassNames []string
				}{
					Display:    g.GroupName,
					ClassNames: g.ClassNames,
				})
			}
			if len(teamDefs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "有効なデフォルトグループ設定が見つかりません。グループを手動で設定してください。"})
				return
			}
		}
	}

	// グループ設定が空の場合はエラー
	if len(teamDefs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "グループ設定が空です。少なくとも1つのグループを設定してください。"})
		return
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

	if user.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}
	if len(user.ID) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user ID must be a valid UUID (36 characters), got %d characters: %q", len(user.ID), user.ID)})
		return
	}
	// 学年対抗リレーの点数設定
	var pointsByRankJSON interface{}
	if req.Session != nil && req.Session.PointsByRank != nil {
		pointsByRankJSON = req.Session.PointsByRank
	}
	run, err := h.noonRepo.CreateTemplateRunWithPointsByRankJSON(session.ID, noonTemplateYearRelay, fmt.Sprintf("学年対抗リレー (event_id=%d)", eventID), user.ID, pointsByRankJSON)
	if err != nil {
		log.Printf("ERROR: CreateYearRelayRun failed to create template run: session_id=%d, user_id=%q, error=%v", session.ID, user.ID, err)
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
func (h *NoonGameHandler) CreateCourseRelayRun(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
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

	var req createTemplateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// リクエストボディが空でもエラーにしない（後方互換性のため）
		req = createTemplateRunRequest{}
	}

	session, err := h.noonRepo.GetSessionByEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		// セッションが存在しない場合はリクエストから設定を取得、またはデフォルト値を使用
		sessionName := fmt.Sprintf("コース対抗リレー_%d", eventID)
		if req.Session != nil && req.Session.Name != "" {
			sessionName = req.Session.Name
		}
		mode := "group"
		if req.Session != nil && req.Session.Mode != "" {
			mode = strings.ToLower(strings.TrimSpace(req.Session.Mode))
		}
		session = &models.NoonGameSession{
			EventID:             eventID,
			Name:                sessionName,
			Description:         nil,
			Mode:                mode,
			WinPoints:           0,
			LossPoints:          0,
			DrawPoints:          0,
			ParticipationPoints: 0,
			AllowManualPoints:   false,
		}
		if req.Session != nil {
			if req.Session.Description != nil {
				session.Description = req.Session.Description
			}
			session.WinPoints = req.Session.WinPoints
			session.LossPoints = req.Session.LossPoints
			session.DrawPoints = req.Session.DrawPoints
			session.ParticipationPoints = req.Session.ParticipationPoints
			if req.Session.AllowManualPoints != nil {
				session.AllowManualPoints = *req.Session.AllowManualPoints
			}
		}
		session, err = h.noonRepo.UpsertSession(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session", "details": err.Error()})
			return
		}
	} else {
		// 既にセッションが存在する場合は、そのセッションを使用（1つのイベントにつき1つのセッションのみ）
		// セッション設定の更新は行わない（テンプレート作成時は既存セッションを使用）
	}

	// 既にテンプレートランが存在する場合は、既存のテンプレートと関連データを削除してから新しいテンプレートを作成
	existingRuns, err := h.noonRepo.ListTemplateRunsBySession(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing template runs", "details": err.Error()})
		return
	}
	if len(existingRuns) > 0 {
		// 既存のテンプレートランと関連データ（試合、グループなど）を削除
		if err := h.noonRepo.DeleteTemplateRunAndRelatedData(session.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete existing template run", "details": err.Error()})
			return
		}
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

	// グループ設定を取得（手動設定があればそれを使用、なければデフォルト設定を使用）
	var teamDefs []struct {
		Display    string
		ClassNames []string
	}

	if req.Session != nil && len(req.Session.Groups) > 0 {
		// 手動設定を使用
		for _, g := range req.Session.Groups {
			// バリデーション: グループ名とクラス名が空でないことを確認
			if strings.TrimSpace(g.GroupName) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "グループ名が空です"})
				return
			}
			if len(g.ClassNames) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("グループ「%s」にクラスが設定されていません", g.GroupName)})
				return
			}
			teamDefs = append(teamDefs, struct {
				Display    string
				ClassNames []string
			}{
				Display:    strings.TrimSpace(g.GroupName),
				ClassNames: g.ClassNames,
			})
		}
	} else {
		// デフォルト設定を使用
		defaultGroups, err := h.noonRepo.GetTemplateDefaultGroups(noonTemplateCourseRelay)
		if err != nil {
			log.Printf("WARNING: failed to fetch default groups, using hardcoded defaults: %v", err)
			// フォールバック: ハードコードされたデフォルト値
			teamDefs = []struct {
				Display    string
				ClassNames []string
			}{
				{Display: "1-1 & IEコース", ClassNames: []string{"1-1", "IE2", "IE3", "IE4", "IE5"}},
				{Display: "1-2 & ISコース", ClassNames: []string{"1-2", "IS2", "IS3", "IS4", "IS5"}},
				{Display: "1-3 & ITコース", ClassNames: []string{"1-3", "IT2", "IT3", "IT4", "IT5"}},
				{Display: "専攻科・教員", ClassNames: []string{"専教"}},
			}
		} else {
			// デフォルト設定をソートして使用
			if len(defaultGroups) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "デフォルトグループ設定が見つかりません。グループを手動で設定してください。"})
				return
			}
			for _, g := range defaultGroups {
				if len(g.ClassNames) == 0 {
					log.Printf("WARNING: default group %q has no class names, skipping", g.GroupName)
					continue
				}
				teamDefs = append(teamDefs, struct {
					Display    string
					ClassNames []string
				}{
					Display:    g.GroupName,
					ClassNames: g.ClassNames,
				})
			}
			if len(teamDefs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "有効なデフォルトグループ設定が見つかりません。グループを手動で設定してください。"})
				return
			}
		}
	}

	// グループ設定が空の場合はエラー
	if len(teamDefs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "グループ設定が空です。少なくとも1つのグループを設定してください。"})
		return
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

	if user.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}
	if len(user.ID) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user ID must be a valid UUID (36 characters), got %d characters: %q", len(user.ID), user.ID)})
		return
	}
	pointsByRank := map[int]int{1: 40, 2: 30, 3: 20, 4: 10} // デフォルト値
	var pointsByRankInterface interface{}
	if req.Session != nil && req.Session.PointsByRank != nil {
		pointsByRankInterface = req.Session.PointsByRank
	} else {
		pointsByRankInterface = pointsByRank
	}

	run, err := h.noonRepo.CreateTemplateRunWithPointsByRankJSON(session.ID, noonTemplateCourseRelay, fmt.Sprintf("コース対抗リレー (event_id=%d)", eventID), user.ID, pointsByRankInterface)
	if err != nil {
		log.Printf("ERROR: CreateCourseRelayRun failed to create template run: session_id=%d, user_id=%q, error=%v", session.ID, user.ID, err)
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

// CreateTugOfWarRun は綱引き(テンプレート)の run を作成し、4チームの試合を生成します。
func (h *NoonGameHandler) CreateTugOfWarRun(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
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

	var req createTemplateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// リクエストボディが空でもエラーにしない（後方互換性のため）
		req = createTemplateRunRequest{}
	}

	session, err := h.noonRepo.GetSessionByEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session", "details": err.Error()})
		return
	}
	if session == nil {
		// セッションが存在しない場合はリクエストから設定を取得、またはデフォルト値を使用
		sessionName := fmt.Sprintf("綱引き_%d", eventID)
		if req.Session != nil && req.Session.Name != "" {
			sessionName = req.Session.Name
		}
		mode := "group"
		if req.Session != nil && req.Session.Mode != "" {
			mode = strings.ToLower(strings.TrimSpace(req.Session.Mode))
		}
		session = &models.NoonGameSession{
			EventID:             eventID,
			Name:                sessionName,
			Description:         nil,
			Mode:                mode,
			WinPoints:           0,
			LossPoints:          0,
			DrawPoints:          0,
			ParticipationPoints: 0,
			AllowManualPoints:   false,
		}
		if req.Session != nil {
			if req.Session.Description != nil {
				session.Description = req.Session.Description
			}
			session.WinPoints = req.Session.WinPoints
			session.LossPoints = req.Session.LossPoints
			session.DrawPoints = req.Session.DrawPoints
			session.ParticipationPoints = req.Session.ParticipationPoints
			if req.Session.AllowManualPoints != nil {
				session.AllowManualPoints = *req.Session.AllowManualPoints
			}
		}
		session, err = h.noonRepo.UpsertSession(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session", "details": err.Error()})
			return
		}
	} else {
		// 既にセッションが存在する場合は、そのセッションを使用（1つのイベントにつき1つのセッションのみ）
		// セッション設定の更新は行わない（テンプレート作成時は既存セッションを使用）
	}

	// 既にテンプレートランが存在する場合は、既存のテンプレートと関連データを削除してから新しいテンプレートを作成
	existingRuns, err := h.noonRepo.ListTemplateRunsBySession(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing template runs", "details": err.Error()})
		return
	}
	if len(existingRuns) > 0 {
		// 既存のテンプレートランと関連データ（試合、グループなど）を削除
		if err := h.noonRepo.DeleteTemplateRunAndRelatedData(session.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete existing template run", "details": err.Error()})
			return
		}
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

	// グループ設定を取得（手動設定があればそれを使用、なければデフォルト設定を使用）
	var teamDefs []struct {
		Display    string
		ClassNames []string
	}

	if req.Session != nil && len(req.Session.Groups) > 0 {
		// 手動設定を使用
		for _, g := range req.Session.Groups {
			teamDefs = append(teamDefs, struct {
				Display    string
				ClassNames []string
			}{
				Display:    g.GroupName,
				ClassNames: g.ClassNames,
			})
		}
	} else {
		// デフォルト設定を使用
		defaultGroups, err := h.noonRepo.GetTemplateDefaultGroups(noonTemplateTugOfWar)
		if err != nil {
			log.Printf("WARNING: failed to fetch default groups, using hardcoded defaults: %v", err)
			// フォールバック: ハードコードされたデフォルト値
			teamDefs = []struct {
				Display    string
				ClassNames []string
			}{
				{Display: "1-1 & ISコース", ClassNames: []string{"1-1", "IS2", "IS3", "IS4", "IS5"}},
				{Display: "1-2 & ITコース", ClassNames: []string{"1-2", "IT2", "IT3", "IT4", "IT5"}},
				{Display: "1-3 & IEコース", ClassNames: []string{"1-3", "IE2", "IE3", "IE4", "IE5"}},
				{Display: "専攻科・教員", ClassNames: []string{"専教"}},
			}
		} else {
			// デフォルト設定をソートして使用
			for _, g := range defaultGroups {
				teamDefs = append(teamDefs, struct {
					Display    string
					ClassNames []string
				}{
					Display:    g.GroupName,
					ClassNames: g.ClassNames,
				})
			}
		}
	}

	// グループ設定が空の場合はエラー
	if len(teamDefs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "グループ設定が空です。少なくとも1つのグループを設定してください。"})
		return
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tug of war group", "details": err.Error()})
			return
		}
		groupIDs = append(groupIDs, grp.ID)
		entryDisplays = append(entryDisplays, def.Display)
	}

	// 試合作成（4チーム）
	title := "綱引き"
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

	if user.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}
	if len(user.ID) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user ID must be a valid UUID (36 characters), got %d characters: %q", len(user.ID), user.ID)})
		return
	}
	pointsByRank := map[int]int{1: 40, 2: 30, 3: 20, 4: 10} // デフォルト値
	var pointsByRankInterface interface{}
	if req.Session != nil && req.Session.PointsByRank != nil {
		pointsByRankInterface = req.Session.PointsByRank
	} else {
		pointsByRankInterface = pointsByRank
	}

	run, err := h.noonRepo.CreateTemplateRunWithPointsByRankJSON(session.ID, noonTemplateTugOfWar, fmt.Sprintf("綱引き (event_id=%d)", eventID), user.ID, pointsByRankInterface)
	if err != nil {
		log.Printf("ERROR: CreateTugOfWarRun failed to create template run: session_id=%d, user_id=%q, error=%v", session.ID, user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template run", "details": err.Error()})
		return
	}

	if _, err := h.noonRepo.LinkTemplateRunMatch(run.ID, match.ID, tugOfWarMatchMain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link match", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createTugOfWarRunResponse{
		Run: run,
		Matches: map[string]*models.NoonGameMatchWithResult{
			tugOfWarMatchMain: match,
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

// --- Templates: 綱引き ---
const (
	noonTemplateTugOfWar = "tug_of_war"
	tugOfWarMatchMain    = "MAIN"
)

type createYearRelayRunResponse struct {
	Run     *models.NoonGameTemplateRun                `json:"run"`
	Matches map[string]*models.NoonGameMatchWithResult `json:"matches"`
}

type createTugOfWarRunResponse struct {
	Run     *models.NoonGameTemplateRun                `json:"run"`
	Matches map[string]*models.NoonGameMatchWithResult `json:"matches"`
}

type createCourseRelayRunResponse struct {
	Run     *models.NoonGameTemplateRun                `json:"run"`
	Matches map[string]*models.NoonGameMatchWithResult `json:"matches"`
}

type createTemplateRunRequest struct {
	Session *struct {
		Name                string                 `json:"name"`
		Description         *string                `json:"description"`
		Mode                string                 `json:"mode"`
		WinPoints           int                    `json:"win_points"`
		LossPoints          int                    `json:"loss_points"`
		DrawPoints          int                    `json:"draw_points"`
		ParticipationPoints int                    `json:"participation_points"`
		AllowManualPoints   *bool                  `json:"allow_manual_points"`
		PointsByRank        map[string]interface{} `json:"points_by_rank,omitempty"` // 学年対抗リレー、コース対抗リレー、綱引き用
		Groups              []templateGroupConfig  `json:"groups,omitempty"`         // 手動グループ設定（オプション）
	} `json:"session"`
}

type templateGroupConfig struct {
	GroupName  string   `json:"group_name" binding:"required"`
	ClassNames []string `json:"class_names" binding:"required"`
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
	Rankings []courseRelayRankInput `json:"rankings" binding:"required"`
	Note     *string                `json:"note"`
}

type tugOfWarRankInput struct {
	EntryID int  `json:"entry_id" binding:"required"`
	Rank    int  `json:"rank" binding:"required"`
	Points  *int `json:"points"` // 同順位(同点)の場合のみ必須
}

type tugOfWarResultRequest struct {
	Rankings []tugOfWarRankInput `json:"rankings" binding:"required"`
	Note     *string             `json:"note"`
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
			"template_runs":  []interface{}{},
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
		log.Printf("ERROR: UpsertSession failed to build session payload: session_id=%d, error=%v", updated.ID, err)
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

// GetTemplateDefaultGroups は指定されたテンプレートキーのデフォルトグループ設定を取得します。
func (h *NoonGameHandler) GetTemplateDefaultGroups(c *gin.Context) {
	templateKey := c.Param("template_key")
	if templateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template_key is required"})
		return
	}

	groups, err := h.noonRepo.GetTemplateDefaultGroups(templateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch default groups", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// SaveTemplateDefaultGroups は指定されたテンプレートキーのデフォルトグループ設定を保存します。
func (h *NoonGameHandler) SaveTemplateDefaultGroups(c *gin.Context) {
	templateKey := c.Param("template_key")
	if templateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template_key is required"})
		return
	}

	var req struct {
		Groups []struct {
			GroupIndex int      `json:"group_index" binding:"required"`
			GroupName  string   `json:"group_name" binding:"required"`
			ClassNames []string `json:"class_names" binding:"required"`
		} `json:"groups" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// バリデーション
	if len(req.Groups) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "グループが1つ以上必要です"})
		return
	}

	groups := make([]*models.NoonGameTemplateDefaultGroup, len(req.Groups))
	for i, g := range req.Groups {
		// グループ名のバリデーション
		if strings.TrimSpace(g.GroupName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("グループ%dの名前が空です", i+1)})
			return
		}
		// クラス名のバリデーション
		if len(g.ClassNames) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("グループ「%s」にクラスが設定されていません", g.GroupName)})
			return
		}
		// 空のクラス名をフィルタリング
		validClassNames := make([]string, 0, len(g.ClassNames))
		for _, className := range g.ClassNames {
			if strings.TrimSpace(className) != "" {
				validClassNames = append(validClassNames, strings.TrimSpace(className))
			}
		}
		if len(validClassNames) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("グループ「%s」に有効なクラス名がありません", g.GroupName)})
			return
		}

		groups[i] = &models.NoonGameTemplateDefaultGroup{
			TemplateKey: templateKey,
			GroupIndex:  g.GroupIndex,
			GroupName:   strings.TrimSpace(g.GroupName),
			ClassNames:  validClassNames,
		}
	}

	if err := h.noonRepo.SaveTemplateDefaultGroups(templateKey, groups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save default groups", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default groups saved"})
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

			// エントリー名を解決
			entryName := ""
			if entry, ok := entryLookup[ranking.EntryID]; ok && entry != nil {
				// classMap と groupMap を取得
				classes, _ := h.classRepo.GetAllClasses(match.SessionID)
				classMap := make(map[int]*models.Class)
				for _, class := range classes {
					classMap[class.ID] = class
				}
				groups, _ := h.noonRepo.GetGroupsWithMembers(match.SessionID)
				groupMap := make(map[int]*models.NoonGameGroupWithMembers)
				for _, group := range groups {
					groupMap[group.ID] = group
				}
				name, _, err := h.resolveEntryDisplay(match.SessionID, entry, classMap, groupMap)
				if err == nil && name != "" {
					entryName = name
				}
			}

			detail := &models.NoonGameResultDetail{
				EntryID:           ranking.EntryID,
				Points:            ranking.Points,
				EntryResolvedName: entryName,
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

	// テンプレートランから点数設定を取得
	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	// デフォルトの点数設定（1位30点、2位25点、3位20点、4位15点、5位10点、6位5点）
	pointsByRank := map[int]int{1: 30, 2: 25, 3: 20, 4: 15, 5: 10, 6: 5}
	// テンプレートランに保存された点数設定を使用
	if run.PointsByRank != nil {
		if yearRelayPoints, ok := run.PointsByRank.(map[string]interface{}); ok {
			blockKey := "block_a"
			if block == yearRelayMatchB {
				blockKey = "block_b"
			}
			if blockPoints, ok := yearRelayPoints[blockKey].(map[string]interface{}); ok {
				pointsByRank = make(map[int]int)
				for k, v := range blockPoints {
					if rank, err := strconv.Atoi(k); err == nil {
						if points, ok := v.(float64); ok {
							pointsByRank[rank] = int(points)
						}
					}
				}
			}
		}
	}

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

	// テンプレートランから点数設定を取得
	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	// デフォルトの点数設定（1位30点、2位20点、3位10点、4位以下0点）
	pointsByRank := map[int]int{1: 30, 2: 20, 3: 10, 4: 0, 5: 0, 6: 0}
	// テンプレートランに保存された点数設定を使用
	if run.PointsByRank != nil {
		if yearRelayPoints, ok := run.PointsByRank.(map[string]interface{}); ok {
			if overallPoints, ok := yearRelayPoints["overall"].(map[string]interface{}); ok {
				pointsByRank = make(map[int]int)
				for k, v := range overallPoints {
					if rank, err := strconv.Atoi(k); err == nil {
						if points, ok := v.(float64); ok {
							pointsByRank[rank] = int(points)
						}
					}
				}
			}
		}
	}

	h.applyYearRelayRankingsToMatch(c, runID, yearRelayMatchBonus, req, pointsByRank)
}

// RecordCourseRelayResult はコース対抗リレーの順位を登録し、テンプレ点を自動付与します。
// 点数はテンプレートランに保存された設定を使用します。同順位がある場合は、該当順位の points を必須入力とします。
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

	// テンプレートランから点数設定を取得
	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	// デフォルトの点数設定（1位40点、2位30点、3位20点、4位10点）
	pointsByRank := map[int]int{1: 40, 2: 30, 3: 20, 4: 10}
	// テンプレートランに保存された点数設定を使用
	if run.PointsByRank != nil {
		if pointsMap, ok := run.PointsByRank.(map[string]interface{}); ok {
			// コース対抗リレーと綱引きの形式: {"1": 40, "2": 30, ...}
			pointsByRank = make(map[int]int)
			for k, v := range pointsMap {
				if rank, err := strconv.Atoi(k); err == nil {
					if points, ok := v.(float64); ok {
						pointsByRank[rank] = int(points)
					}
				}
			}
		}
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

// RecordTugOfWarResult は綱引きの順位を登録し、テンプレ点を自動付与します。
// 点数はテンプレートランに保存された設定を使用します。同順位がある場合は、該当順位の points を必須入力とします。
func (h *NoonGameHandler) RecordTugOfWarResult(c *gin.Context) {
	runID, err := strconv.Atoi(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run_id"})
		return
	}

	var req tugOfWarResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// テンプレートランから点数設定を取得
	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	// デフォルトの点数設定（1位40点、2位30点、3位20点、4位10点）
	pointsByRank := map[int]int{1: 40, 2: 30, 3: 20, 4: 10}
	// テンプレートランに保存された点数設定を使用
	if run.PointsByRank != nil {
		if pointsMap, ok := run.PointsByRank.(map[string]interface{}); ok {
			// コース対抗リレーと綱引きの形式: {"1": 40, "2": 30, ...}
			pointsByRank = make(map[int]int)
			for k, v := range pointsMap {
				if rank, err := strconv.Atoi(k); err == nil {
					if points, ok := v.(float64); ok {
						pointsByRank[rank] = int(points)
					}
				}
			}
		}
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

	h.applyTugOfWarRankingsToMatch(c, runID, tugOfWarMatchMain, yearRelayReq, pointsByRank)
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

	// classMap と groupMap を取得
	classes, err := h.classRepo.GetAllClasses(session.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes", "details": err.Error()})
		return
	}
	classMap := make(map[int]*models.Class)
	for _, class := range classes {
		classMap[class.ID] = class
	}
	groups, err := h.noonRepo.GetGroupsWithMembers(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch groups", "details": err.Error()})
		return
	}
	groupMap := make(map[int]*models.NoonGameGroupWithMembers)
	for _, group := range groups {
		groupMap[group.ID] = group
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

		// エントリー名を解決
		entryName := ""
		if entry != nil {
			name, _, err := h.resolveEntryDisplay(session.ID, entry, classMap, groupMap)
			if err == nil && name != "" {
				entryName = name
			}
		}

		rankCopy := r.Rank
		resultDetails = append(resultDetails, &models.NoonGameResultDetail{
			EntryID:           r.EntryID,
			Rank:              &rankCopy,
			Points:            points,
			EntryResolvedName: entryName,
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

	// AブロックまたはBブロックの結果が登録された場合、両方のブロックの結果が揃っていれば総合ボーナスを自動計算
	if matchKey == yearRelayMatchA || matchKey == yearRelayMatchB {
		if err := h.calculateAndRecordYearRelayOverallBonus(runID, user); err != nil {
			// エラーが発生しても、ブロック結果の登録は成功しているので、エラーログを出力するだけ
			log.Printf("WARNING: failed to calculate overall bonus: run_id=%d, error=%v", runID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"match": full})
}

// calculateAndRecordYearRelayOverallBonus は学年対抗リレーの総合ボーナスを自動計算して登録します。
// AブロックとBブロックの両方の結果が揃っている場合のみ実行されます。
func (h *NoonGameHandler) calculateAndRecordYearRelayOverallBonus(runID int, user *models.User) error {
	// テンプレートランから点数設定を取得
	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		return fmt.Errorf("failed to fetch template run: %w", err)
	}
	if run == nil || run.TemplateKey != noonTemplateYearRelay {
		return fmt.Errorf("template run not found or invalid")
	}

	// classMap と groupMap を取得
	sessionObj, err := h.noonRepo.GetSessionByID(run.SessionID)
	if err != nil {
		return fmt.Errorf("failed to fetch session: %w", err)
	}
	if sessionObj == nil {
		return fmt.Errorf("session not found")
	}
	classes, err := h.classRepo.GetAllClasses(sessionObj.EventID)
	if err != nil {
		return fmt.Errorf("failed to fetch classes: %w", err)
	}
	classMap := make(map[int]*models.Class)
	for _, class := range classes {
		classMap[class.ID] = class
	}
	groups, err := h.noonRepo.GetGroupsWithMembers(run.SessionID)
	if err != nil {
		return fmt.Errorf("failed to fetch groups: %w", err)
	}
	groupMap := make(map[int]*models.NoonGameGroupWithMembers)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	// session 変数を sessionObj から取得
	session := sessionObj

	// AブロックとBブロックの試合を取得
	linkA, err := h.noonRepo.GetTemplateRunMatchByKey(runID, yearRelayMatchA)
	if err != nil {
		return fmt.Errorf("failed to fetch A block match link: %w", err)
	}
	if linkA == nil {
		return fmt.Errorf("A block match link not found")
	}

	linkB, err := h.noonRepo.GetTemplateRunMatchByKey(runID, yearRelayMatchB)
	if err != nil {
		return fmt.Errorf("failed to fetch B block match link: %w", err)
	}
	if linkB == nil {
		return fmt.Errorf("B block match link not found")
	}

	matchA, err := h.noonRepo.GetMatchByID(linkA.MatchID)
	if err != nil {
		return fmt.Errorf("failed to fetch A block match: %w", err)
	}
	if matchA == nil {
		return fmt.Errorf("A block match not found")
	}

	matchB, err := h.noonRepo.GetMatchByID(linkB.MatchID)
	if err != nil {
		return fmt.Errorf("failed to fetch B block match: %w", err)
	}
	if matchB == nil {
		return fmt.Errorf("B block match not found")
	}

	// AブロックとBブロックの両方に結果が登録されているか確認
	if matchA.Result == nil || matchB.Result == nil {
		log.Printf("INFO: skip overall bonus (results not ready) run_id=%d matchA_result_nil=%t matchB_result_nil=%t", runID, matchA.Result == nil, matchB.Result == nil)
		return nil // まだ両方の結果が揃っていない場合は何もしない
	}

	// 総合ボーナスの試合を取得
	linkBonus, err := h.noonRepo.GetTemplateRunMatchByKey(runID, yearRelayMatchBonus)
	if err != nil {
		return fmt.Errorf("failed to fetch bonus match link: %w", err)
	}
	if linkBonus == nil {
		return fmt.Errorf("bonus match link not found")
	}

	matchBonus, err := h.noonRepo.GetMatchByID(linkBonus.MatchID)
	if err != nil {
		return fmt.Errorf("failed to fetch bonus match: %w", err)
	}
	if matchBonus == nil {
		return fmt.Errorf("bonus match not found")
	}

	// 既存の総合ボーナスがあっても毎回再計算して上書きする
	// （空結果が残っていても必ず計算し直す）

	// session は既に取得済みなので、sessionObj を使用

	// 各チームのAブロックとBブロックの合計点数を計算
	// groupIDをキーとして、AブロックとBブロックの点数を集計（同じ学年のエントリーをマッチング）
	groupTotalPoints := make(map[int]int)  // groupID -> 合計点数
	groupToClassIDs := make(map[int][]int) // groupID -> classIDs
	groupToEntryID := make(map[int]int)    // groupID -> 総合ボーナス試合のentryID

	log.Printf("INFO: start overall bonus calc run_id=%d match_bonus_id=%d entries=%d a_details=%d b_details=%d", runID, matchBonus.ID, len(matchBonus.Entries), len(matchA.Result.Details), len(matchB.Result.Details))

	// 総合ボーナス試合のエントリーをマッピング（groupID -> entryID）
	for _, entry := range matchBonus.Entries {
		if entry != nil && entry.SideType == "group" && entry.GroupID != nil {
			groupToEntryID[*entry.GroupID] = entry.ID
			// classIDsも取得
			classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
			if err == nil && len(classIDs) > 0 {
				groupToClassIDs[*entry.GroupID] = classIDs
			}
		}
	}

	// Aブロックの点数を集計
	if matchA.Result != nil && matchA.Result.Details != nil {
		for _, detail := range matchA.Result.Details {
			if detail.Points > 0 {
				found := false
				var resolvedName string
				if detail.EntryResolvedName != "" {
					resolvedName = strings.TrimSpace(detail.EntryResolvedName)
				}
				// エントリーからgroupIDを取得（マッチ内）
				for _, entry := range matchA.Entries {
					if entry != nil && entry.ID == detail.EntryID && entry.SideType == "group" && entry.GroupID != nil {
						groupID := *entry.GroupID
						groupTotalPoints[groupID] += detail.Points
						if _, ok := groupToClassIDs[groupID]; !ok {
							classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
							if err == nil && len(classIDs) > 0 {
								groupToClassIDs[groupID] = classIDs
							}
						}
						found = true
						break
					}
				}
				// マッチ内に該当エントリーが無い場合、DBまたは名前マッチで補完
				if !found {
					if entry, err := h.noonRepo.GetEntryByID(detail.EntryID); err == nil && entry != nil && entry.SideType == "group" && entry.GroupID != nil {
						groupID := *entry.GroupID
						groupTotalPoints[groupID] += detail.Points
						if _, ok := groupToClassIDs[groupID]; !ok {
							classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
							if err == nil && len(classIDs) > 0 {
								groupToClassIDs[groupID] = classIDs
							}
						}
						if _, ok := groupToEntryID[groupID]; !ok {
							// BONUSエントリーが不一致の場合もあるので、fallbackとしてセット
							groupToEntryID[groupID] = detail.EntryID
						}
						log.Printf("INFO: bonus calc A-block fallback entry lookup by id detail_entry_id=%d group_id=%d", detail.EntryID, groupID)
						found = true
					}
				}
				if !found {
					// 名前でBONUSエントリーとマッチさせて groupID を推定
					groupID := resolveGroupIDByName(resolvedName, matchBonus.Entries, groupMap)
					if groupID != nil {
						gid := *groupID
						groupTotalPoints[gid] += detail.Points
						if _, ok := groupToClassIDs[gid]; !ok {
							classIDs, err := h.resolveClassIDs("group", nil, groupID)
							if err == nil && len(classIDs) > 0 {
								groupToClassIDs[gid] = classIDs
							}
						}
						if _, ok := groupToEntryID[gid]; !ok {
							groupToEntryID[gid] = detail.EntryID
						}
						log.Printf("INFO: bonus calc A-block fallback by name detail_entry_id=%d name=%q group_id=%d", detail.EntryID, resolvedName, gid)
						found = true
					}
				}
				if !found {
					log.Printf("WARNING: bonus calc A-block entry not found detail_entry_id=%d name=%q", detail.EntryID, resolvedName)
				}
			}
		}
	}

	// Bブロックの点数を集計
	if matchB.Result != nil && matchB.Result.Details != nil {
		for _, detail := range matchB.Result.Details {
			if detail.Points > 0 {
				found := false
				var resolvedName string
				if detail.EntryResolvedName != "" {
					resolvedName = strings.TrimSpace(detail.EntryResolvedName)
				}
				// エントリーからgroupIDを取得（マッチ内）
				for _, entry := range matchB.Entries {
					if entry != nil && entry.ID == detail.EntryID && entry.SideType == "group" && entry.GroupID != nil {
						groupID := *entry.GroupID
						groupTotalPoints[groupID] += detail.Points
						if _, ok := groupToClassIDs[groupID]; !ok {
							classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
							if err == nil && len(classIDs) > 0 {
								groupToClassIDs[groupID] = classIDs
							}
						}
						found = true
						break
					}
				}
				// マッチ内に該当エントリーが無い場合、DBまたは名前マッチで補完
				if !found {
					if entry, err := h.noonRepo.GetEntryByID(detail.EntryID); err == nil && entry != nil && entry.SideType == "group" && entry.GroupID != nil {
						groupID := *entry.GroupID
						groupTotalPoints[groupID] += detail.Points
						if _, ok := groupToClassIDs[groupID]; !ok {
							classIDs, err := h.resolveClassIDs(entry.SideType, entry.ClassID, entry.GroupID)
							if err == nil && len(classIDs) > 0 {
								groupToClassIDs[groupID] = classIDs
							}
						}
						if _, ok := groupToEntryID[groupID]; !ok {
							groupToEntryID[groupID] = detail.EntryID
						}
						log.Printf("INFO: bonus calc B-block fallback entry lookup by id detail_entry_id=%d group_id=%d", detail.EntryID, groupID)
						found = true
					}
				}
				if !found {
					groupID := resolveGroupIDByName(resolvedName, matchBonus.Entries, groupMap)
					if groupID != nil {
						gid := *groupID
						groupTotalPoints[gid] += detail.Points
						if _, ok := groupToClassIDs[gid]; !ok {
							classIDs, err := h.resolveClassIDs("group", nil, groupID)
							if err == nil && len(classIDs) > 0 {
								groupToClassIDs[gid] = classIDs
							}
						}
						if _, ok := groupToEntryID[gid]; !ok {
							groupToEntryID[gid] = detail.EntryID
						}
						log.Printf("INFO: bonus calc B-block fallback by name detail_entry_id=%d name=%q group_id=%d", detail.EntryID, resolvedName, gid)
						found = true
					}
				}
				if !found {
					log.Printf("WARNING: bonus calc B-block entry not found detail_entry_id=%d name=%q", detail.EntryID, resolvedName)
				}
			}
		}
	}

	// 合計点数でソートして順位を決定
	type groupScore struct {
		GroupID int
		EntryID int
		Points  int
	}
	scores := make([]groupScore, 0, len(groupTotalPoints))
	for groupID, points := range groupTotalPoints {
		entryID, ok := groupToEntryID[groupID]
		if !ok {
			// 総合ボーナス試合にエントリーがない場合はスキップ
			continue
		}
		scores = append(scores, groupScore{
			GroupID: groupID,
			EntryID: entryID,
			Points:  points,
		})
	}
	// 点数が高い順にソート
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Points < scores[j].Points {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// 順位を決定（同点の場合は同じ順位）
	rankings := make([]yearRelayRankInput, 0, len(scores))
	currentRank := 1
	for i, score := range scores {
		if i > 0 && score.Points < scores[i-1].Points {
			currentRank = i + 1
		}
		rankings = append(rankings, yearRelayRankInput{
			EntryID: score.EntryID,
			Rank:    currentRank,
			Points:  nil, // 総合ボーナスは自動計算なのでpointsは不要
		})
	}

	// テンプレートランから総合ボーナスの点数設定を取得
	pointsByRank := map[int]int{1: 30, 2: 20, 3: 10, 4: 0, 5: 0, 6: 0}
	if run.PointsByRank != nil {
		if yearRelayPoints, ok := run.PointsByRank.(map[string]interface{}); ok {
			if overallPoints, ok := yearRelayPoints["overall"].(map[string]interface{}); ok {
				pointsByRank = make(map[int]int)
				for k, v := range overallPoints {
					if rank, err := strconv.Atoi(k); err == nil {
						if points, ok := v.(float64); ok {
							pointsByRank[rank] = int(points)
						}
					}
				}
			}
		}
	}

	// 総合ボーナスの点数を付与
	pointsEntries := make([]*models.NoonGamePoint, 0)
	resultDetails := make([]*models.NoonGameResultDetail, 0)

	matchTitle := fmt.Sprintf("試合 #%d", matchBonus.ID)
	if matchBonus.Title != nil && strings.TrimSpace(*matchBonus.Title) != "" {
		matchTitle = strings.TrimSpace(*matchBonus.Title)
	}

	for _, ranking := range rankings {
		// ranking.EntryIDからgroupIDを逆引き
		var groupID *int
		for _, entry := range matchBonus.Entries {
			if entry != nil && entry.ID == ranking.EntryID && entry.SideType == "group" && entry.GroupID != nil {
				groupID = entry.GroupID
				break
			}
		}
		if groupID == nil {
			continue
		}

		classIDs := groupToClassIDs[*groupID]
		if len(classIDs) == 0 {
			continue
		}

		points := 0
		if p, ok := pointsByRank[ranking.Rank]; ok {
			points = p
		}

		if points > 0 {
			reason := fmt.Sprintf("昼競技テンプレ(%s) (%s)", run.Name, matchTitle)
			for _, classID := range classIDs {
				reasonCopy := reason
				mid := matchBonus.ID
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
		}

		// エントリー名を解決（総合ボーナス試合のエントリーを使用）
		entryName := ""
		for _, entry := range matchBonus.Entries {
			if entry != nil && entry.ID == ranking.EntryID {
				name, _, err := h.resolveEntryDisplay(sessionObj.ID, entry, classMap, groupMap)
				if err == nil && name != "" {
					entryName = name
				}
				break
			}
		}

		rankCopy := ranking.Rank
		resultDetails = append(resultDetails, &models.NoonGameResultDetail{
			EntryID:           ranking.EntryID,
			Rank:              &rankCopy,
			Points:            points,
			EntryResolvedName: entryName,
		})
	}

	// 既存ポイント・既存結果を消す（空結果が残っているケースに対応）
	if err := h.noonRepo.ClearPointsForMatch(matchBonus.ID); err != nil {
		return fmt.Errorf("failed to clear existing points: %w", err)
	}
	if matchBonus.Result != nil {
		if _, err := h.noonRepo.SaveResult(&models.NoonGameResult{
			MatchID:    matchBonus.ID,
			Winner:     "draw",
			RecordedBy: user.ID,
			Note:       nil,
			Details:    []*models.NoonGameResultDetail{}, // 既存 detail を一旦クリア
		}); err != nil {
			return fmt.Errorf("failed to clear existing bonus result: %w", err)
		}
	}

	// 点数を登録
	if err := h.noonRepo.InsertPoints(pointsEntries); err != nil {
		return fmt.Errorf("failed to store points: %w", err)
	}

	// 試合結果を登録
	if _, err := h.noonRepo.SaveResult(&models.NoonGameResult{
		MatchID:    matchBonus.ID,
		Winner:     "draw", // 総合ボーナスは順位ベースなのでwinnerはdraw
		RecordedBy: user.ID,
		Note:       nil,
		Details:    resultDetails,
	}); err != nil {
		return fmt.Errorf("failed to store match result: %w", err)
	}

	// 試合ステータスを更新
	matchBonus.Status = "completed"
	if _, err := h.noonRepo.SaveMatch(matchBonus.NoonGameMatch); err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}

	// クラス得点へ反映
	summary, err := h.noonRepo.SumPointsByClass(session.ID)
	if err != nil {
		return fmt.Errorf("failed to aggregate points: %w", err)
	}
	if err := h.classRepo.SetNoonGamePoints(session.EventID, summary); err != nil {
		return fmt.Errorf("failed to update class scores: %w", err)
	}

	log.Printf("INFO: overall bonus saved run_id=%d match_bonus_id=%d details=%d points_entries=%d", runID, matchBonus.ID, len(resultDetails), len(pointsEntries))

	return nil
}

// resolveGroupIDByName は、detail.EntryResolvedName などから groupID を推定するためのフォールバック。
// ・ボーナス試合のエントリー表示名/ResolvedName と突き合わせ
// ・groupMap に名称があれば名称一致で解決
func resolveGroupIDByName(name string, entries []*models.NoonGameMatchEntry, groupMap map[int]*models.NoonGameGroupWithMembers) *int {
	if strings.TrimSpace(name) == "" {
		return nil
	}
	n := strings.TrimSpace(name)
	// bonus match entries 優先
	for _, e := range entries {
		if e != nil && e.SideType == "group" && e.GroupID != nil {
			display := ""
			if e.DisplayName != nil {
				display = strings.TrimSpace(*e.DisplayName)
			}
			if strings.EqualFold(display, n) {
				return e.GroupID
			}
		}
	}
	// groupMap の名称マッチ
	for gid, g := range groupMap {
		if g != nil && strings.EqualFold(strings.TrimSpace(g.Name), n) {
			id := gid
			return &id
		}
	}
	return nil
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

	// classMap と groupMap を取得
	classes, err := h.classRepo.GetAllClasses(session.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes", "details": err.Error()})
		return
	}
	classMap := make(map[int]*models.Class)
	for _, class := range classes {
		classMap[class.ID] = class
	}
	groups, err := h.noonRepo.GetGroupsWithMembers(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch groups", "details": err.Error()})
		return
	}
	groupMap := make(map[int]*models.NoonGameGroupWithMembers)
	for _, group := range groups {
		groupMap[group.ID] = group
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

		// エントリー名を解決
		entryName := ""
		if entry != nil {
			name, _, err := h.resolveEntryDisplay(session.ID, entry, classMap, groupMap)
			if err == nil && name != "" {
				entryName = name
			}
		}

		rankCopy := r.Rank
		resultDetails = append(resultDetails, &models.NoonGameResultDetail{
			EntryID:           r.EntryID,
			Rank:              &rankCopy,
			Points:            points,
			EntryResolvedName: entryName,
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

func (h *NoonGameHandler) applyTugOfWarRankingsToMatch(
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
	if run == nil || run.TemplateKey != noonTemplateTugOfWar {
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

	// classMap と groupMap を取得
	classes, err := h.classRepo.GetAllClasses(session.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes", "details": err.Error()})
		return
	}
	classMap := make(map[int]*models.Class)
	for _, class := range classes {
		classMap[class.ID] = class
	}
	groups, err := h.noonRepo.GetGroupsWithMembers(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch groups", "details": err.Error()})
		return
	}
	groupMap := make(map[int]*models.NoonGameGroupWithMembers)
	for _, group := range groups {
		groupMap[group.ID] = group
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

		// エントリー名を解決
		entryName := ""
		if entry != nil {
			name, _, err := h.resolveEntryDisplay(session.ID, entry, classMap, groupMap)
			if err == nil && name != "" {
				entryName = name
			}
		}

		rankCopy := r.Rank
		resultDetails = append(resultDetails, &models.NoonGameResultDetail{
			EntryID:           r.EntryID,
			Rank:              &rankCopy,
			Points:            points,
			EntryResolvedName: entryName,
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

	templateRuns, err := h.noonRepo.ListTemplateRunsBySession(session.ID)
	if err != nil {
		log.Printf("ERROR: buildSessionPayload failed to fetch template runs: session_id=%d, error=%v", session.ID, err)
		return nil, fmt.Errorf("failed to fetch template runs: %w", err)
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
		"template_runs":  templateRuns,
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
					// まず entryNameMap から取得を試みる
					if name, ok := entryNameMap[detail.EntryID]; ok && name != "" {
						detail.EntryResolvedName = name
					}
					// entryNameMap にない場合は entryLookup から取得
					if detail.EntryResolvedName == "" {
						if entry, ok := entryLookup[detail.EntryID]; ok && entry != nil {
							detail.EntryResolvedName = entry.ResolvedName
						}
					}
					// まだ設定されていない場合、match.Entries から直接探す
					if detail.EntryResolvedName == "" && len(match.Entries) > 0 {
						for _, entry := range match.Entries {
							if entry != nil && entry.ID == detail.EntryID {
								name, classIDs, err := h.resolveEntryDisplay(match.SessionID, entry, classMap, groupMap)
								if err == nil && name != "" {
									entry.ResolvedName = name
									entry.ClassIDs = classIDs
									detail.EntryResolvedName = name
									// 次回のために entryNameMap にも追加
									entryNameMap[entry.ID] = name
									entryLookup[entry.ID] = entry
								}
								break
							}
						}
					}
					// まだ設定されていない場合、データベースから直接エントリー情報を取得
					if detail.EntryResolvedName == "" {
						entry, err := h.noonRepo.GetEntryByID(detail.EntryID)
						if err == nil && entry != nil {
							name, classIDs, err := h.resolveEntryDisplay(match.SessionID, entry, classMap, groupMap)
							if err == nil && name != "" {
								entry.ResolvedName = name
								entry.ClassIDs = classIDs
								detail.EntryResolvedName = name
								// 次回のために entryNameMap にも追加
								entryNameMap[entry.ID] = name
								entryLookup[entry.ID] = entry
							}
						} else {
							// エントリーが存在しない場合、現在のエントリーから順位と点数で照合を試みる
							// これはエントリーが削除・再作成された場合のフォールバック
							if len(match.Entries) > 0 && detail.Rank != nil {
								// 順位に基づいてエントリーを推測（最後の手段）
								rankIndex := *detail.Rank - 1
								if rankIndex >= 0 && rankIndex < len(match.Entries) {
									entryByRank := match.Entries[rankIndex]
									if entryByRank != nil {
										name, _, err := h.resolveEntryDisplay(match.SessionID, entryByRank, classMap, groupMap)
										if err == nil && name != "" {
											detail.EntryResolvedName = name
										}
									}
								}
							}
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

// GetTemplateRunByMatchID は試合IDからテンプレートラン情報を取得します
func (h *NoonGameHandler) GetTemplateRunByMatchID(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match_id"})
		return
	}

	link, err := h.noonRepo.GetTemplateRunMatchByMatchID(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run match", "details": err.Error()})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run match not found"})
		return
	}

	run, err := h.noonRepo.GetTemplateRunByID(link.RunID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"run":       run,
		"match_key": link.MatchKey,
	})
}
