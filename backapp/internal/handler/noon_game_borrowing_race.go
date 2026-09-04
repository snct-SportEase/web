package handler

import (
	"backapp/internal/models"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	noonTemplateBorrowingRace       = "borrowing_race"
	borrowingRaceMatchMain          = "MAIN"
	defaultBorrowingRaceSessionName = "借り物競争"
)

var defaultBorrowingRacePointsByRank = map[string]int{
	"1": 100, "2": 90, "3": 80, "4": 70, "5": 60,
	"6": 50, "7": 40, "8": 30, "9": 20, "10": 10,
}

type borrowingRaceRunRequest struct {
	Session struct {
		Name                string         `json:"name"`
		Description         *string        `json:"description"`
		ScheduledAt         *string        `json:"scheduled_at"`
		Location            *string        `json:"location"`
		DurationMinutes     int            `json:"duration_minutes"`
		ParticipantClassIDs []int          `json:"participant_class_ids"`
		PointsByRank        map[string]int `json:"points_by_rank"`
	} `json:"session"`
}

type borrowingRaceConfig struct {
	DurationMinutes int            `json:"duration_minutes"`
	ItemPoints      []int          `json:"item_points"`
	RankPoints      map[string]int `json:"rank_points"`
}

type borrowingRaceScoreInput struct {
	EntryID          int `json:"entry_id" binding:"required"`
	CompetitionScore int `json:"competition_score"`
}

type borrowingRaceResultRequest struct {
	Scores   []borrowingRaceScoreInput `json:"scores" binding:"required"`
	Finalize bool                      `json:"finalize"`
	Note     *string                   `json:"note"`
}

func (h *NoonGameHandler) CreateBorrowingRaceRun(c *gin.Context) {
	eventIDRaw := c.Param("event_id")
	if eventIDRaw == "" {
		eventIDRaw = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := userValue.(*models.User)
	if !ok || len(user.ID) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid user ID is required"})
		return
	}

	var req borrowingRaceRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	durationMinutes := req.Session.DurationMinutes
	if durationMinutes == 0 {
		durationMinutes = 15
	}
	if durationMinutes < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "競技時間は1分以上で指定してください"})
		return
	}
	var scheduledAt *time.Time
	if req.Session.ScheduledAt != nil && strings.TrimSpace(*req.Session.ScheduledAt) != "" {
		parsed, err := time.Parse(time.RFC3339, *req.Session.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at"})
			return
		}
		scheduledAt = &parsed
	}

	rankPoints := req.Session.PointsByRank
	if len(rankPoints) == 0 {
		rankPoints = cloneStringIntMap(defaultBorrowingRacePointsByRank)
	}
	for rank, points := range rankPoints {
		rankNumber, err := strconv.Atoi(rank)
		if err != nil || rankNumber < 1 || points < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "順位点は1位以上・0点以上で指定してください"})
			return
		}
	}

	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classes"})
		return
	}
	if len(classes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "クラスが存在しません"})
		return
	}
	classByID := make(map[int]*models.Class, len(classes))
	for _, class := range classes {
		classByID[class.ID] = class
	}

	participantIDs := req.Session.ParticipantClassIDs
	if len(participantIDs) == 0 {
		participantIDs = make([]int, 0, len(classes))
		for _, class := range classes {
			participantIDs = append(participantIDs, class.ID)
		}
	}
	seenClass := make(map[int]bool, len(participantIDs))
	participants := make([]*models.Class, 0, len(participantIDs))
	for _, classID := range participantIDs {
		class := classByID[classID]
		if class == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("参加クラスが大会に存在しません: %d", classID)})
			return
		}
		if seenClass[classID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参加クラスが重複しています"})
			return
		}
		seenClass[classID] = true
		participants = append(participants, class)
	}
	if len(participants) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参加クラスは2クラス以上選択してください"})
		return
	}

	session, err := h.getSessionForTemplate(eventID, noonTemplateBorrowingRace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session"})
		return
	}
	if session == nil {
		session = &models.NoonGameSession{EventID: eventID, TemplateKey: noonTemplateBorrowingRace, Status: "draft"}
	} else {
		if err := h.noonRepo.DeleteTemplateRunAndRelatedData(session.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace existing template run"})
			return
		}
		if err := h.rebuildNoonGameScores(eventID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rebuild class scores"})
			return
		}
	}
	session.Name = strings.TrimSpace(req.Session.Name)
	if session.Name == "" {
		session.Name = defaultBorrowingRaceSessionName
	}
	session.Description = req.Session.Description
	session.Location = req.Session.Location
	session.Mode = "class"
	session.AllowManualPoints = false
	session.ScheduledAt = scheduledAt
	session, err = h.noonRepo.UpsertSession(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session"})
		return
	}
	if err := h.syncNoonGameSport(eventID, session.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync noon game sport"})
		return
	}

	title := session.Name
	format := fmt.Sprintf("順位戦・%d分", durationMinutes)
	match := &models.NoonGameMatch{
		SessionID: session.ID,
		Title:     &title,
		Location:  session.Location,
		Format:    &format,
		Status:    "scheduled",
		AllowDraw: true,
		Entries:   make([]*models.NoonGameMatchEntry, 0, len(participants)),
	}
	for _, class := range participants {
		classID := class.ID
		displayName := class.Name
		match.Entries = append(match.Entries, &models.NoonGameMatchEntry{
			SideType: "class", ClassID: &classID, DisplayName: &displayName,
		})
	}
	savedMatch, err := h.noonRepo.SaveMatch(match)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create borrowing race match"})
		return
	}
	fullMatch, err := h.noonRepo.GetMatchByID(savedMatch.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch borrowing race match"})
		return
	}

	config := borrowingRaceConfig{DurationMinutes: durationMinutes, ItemPoints: []int{1, 2, 4}, RankPoints: rankPoints}
	run, err := h.noonRepo.CreateTemplateRunWithPointsByRankJSON(session.ID, noonTemplateBorrowingRace, session.Name, user.ID, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template run"})
		return
	}
	if _, err := h.noonRepo.LinkTemplateRunMatch(run.ID, savedMatch.ID, borrowingRaceMatchMain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link borrowing race match"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"session": session, "run": run, "matches": gin.H{borrowingRaceMatchMain: fullMatch}})
}

func (h *NoonGameHandler) RecordBorrowingRaceResult(c *gin.Context) {
	runID, err := strconv.Atoi(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run_id"})
		return
	}
	var req borrowingRaceResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	userValue, exists := c.Get("user")
	user, ok := userValue.(*models.User)
	if !exists || !ok || user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	run, err := h.noonRepo.GetTemplateRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template run"})
		return
	}
	if run == nil || run.TemplateKey != noonTemplateBorrowingRace {
		c.JSON(http.StatusNotFound, gin.H{"error": "template run not found"})
		return
	}
	link, err := h.noonRepo.GetTemplateRunMatchByKey(runID, borrowingRaceMatchMain)
	if err != nil || link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run match not found"})
		return
	}
	match, err := h.noonRepo.GetMatchByID(link.MatchID)
	if err != nil || match == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	session, err := h.noonRepo.GetSessionByID(match.SessionID)
	if err != nil || session == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session not found"})
		return
	}

	entryByID := make(map[int]*models.NoonGameMatchEntry, len(match.Entries))
	for _, entry := range match.Entries {
		if entry != nil {
			entryByID[entry.ID] = entry
		}
	}
	if len(req.Scores) != len(entryByID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "全参加クラスの獲得点を入力してください"})
		return
	}
	seenEntry := make(map[int]bool, len(req.Scores))
	for _, score := range req.Scores {
		if entryByID[score.EntryID] == nil || seenEntry[score.EntryID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不明または重複したentry_idです"})
			return
		}
		if score.CompetitionScore < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "獲得点は0以上で入力してください"})
			return
		}
		seenEntry[score.EntryID] = true
	}

	rankedScores := rankBorrowingRaceScores(req.Scores)
	rankPoints := borrowingRaceRankPoints(run.PointsByRank)
	details := make([]*models.NoonGameResultDetail, 0, len(rankedScores))
	points := make([]*models.NoonGamePoint, 0, len(rankedScores))
	bestEntryID := 0
	bestCount := 0
	for _, score := range rankedScores {
		entry := entryByID[score.EntryID]
		competitionScore := score.CompetitionScore
		rank := score.Rank
		awardPoints := 0
		if req.Finalize {
			awardPoints = rankPoints[rank]
		}
		details = append(details, &models.NoonGameResultDetail{
			EntryID: score.EntryID, Rank: &rank, CompetitionScore: &competitionScore,
			Points: awardPoints, EntryResolvedName: strings.TrimSpace(valueOrEmpty(entry.DisplayName)),
		})
		if rank == 1 {
			bestCount++
			bestEntryID = score.EntryID
		}
		if req.Finalize && awardPoints != 0 && entry.ClassID != nil {
			reason := fmt.Sprintf("借り物競争 %d位", rank)
			matchID := match.ID
			points = append(points, &models.NoonGamePoint{
				SessionID: session.ID, MatchID: &matchID, ClassID: *entry.ClassID,
				Points: awardPoints, Reason: &reason, Source: "result", CreatedBy: user.ID,
			})
		}
	}

	if err := h.noonRepo.ClearPointsForMatch(match.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear existing points"})
		return
	}
	if req.Finalize && len(points) > 0 {
		if err := h.noonRepo.InsertPoints(points); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store points"})
			return
		}
	}

	winner := "draw"
	if bestCount == 1 && len(match.Entries) >= 2 {
		if match.Entries[0].ID == bestEntryID {
			winner = "home"
		} else if match.Entries[1].ID == bestEntryID {
			winner = "away"
		}
	}
	if _, err := h.noonRepo.SaveResult(&models.NoonGameResult{
		MatchID: match.ID, Winner: winner, RecordedBy: user.ID, Note: req.Note, Details: details,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save result"})
		return
	}
	if req.Finalize {
		match.Status = "completed"
	} else {
		match.Status = "in_progress"
	}
	if _, err := h.noonRepo.SaveMatch(match.NoonGameMatch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match status"})
		return
	}
	if err := h.rebuildNoonGameScores(session.EventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update class scores"})
		return
	}
	fullMatch, err := h.noonRepo.GetMatchByID(match.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match"})
		return
	}
	h.normalizeMatchesToJST([]*models.NoonGameMatchWithResult{fullMatch})
	c.JSON(http.StatusOK, gin.H{"match": fullMatch, "finalized": req.Finalize})
}

type rankedBorrowingRaceScore struct {
	borrowingRaceScoreInput
	Rank int
}

func rankBorrowingRaceScores(scores []borrowingRaceScoreInput) []rankedBorrowingRaceScore {
	ordered := make([]borrowingRaceScoreInput, len(scores))
	copy(ordered, scores)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].CompetitionScore == ordered[j].CompetitionScore {
			return ordered[i].EntryID < ordered[j].EntryID
		}
		return ordered[i].CompetitionScore > ordered[j].CompetitionScore
	})
	ranked := make([]rankedBorrowingRaceScore, 0, len(ordered))
	for index, score := range ordered {
		rank := 1
		if index > 0 {
			rank = ranked[index-1].Rank
			if score.CompetitionScore < ordered[index-1].CompetitionScore {
				rank = index + 1
			}
		}
		ranked = append(ranked, rankedBorrowingRaceScore{borrowingRaceScoreInput: score, Rank: rank})
	}
	return ranked
}

func borrowingRaceRankPoints(value interface{}) map[int]int {
	points := make(map[int]int)
	config, ok := value.(map[string]interface{})
	if !ok {
		for rank, score := range defaultBorrowingRacePointsByRank {
			rankNumber, _ := strconv.Atoi(rank)
			points[rankNumber] = score
		}
		return points
	}
	rawRankPoints, ok := config["rank_points"].(map[string]interface{})
	if !ok {
		return points
	}
	for rank, value := range rawRankPoints {
		rankNumber, err := strconv.Atoi(rank)
		if err != nil {
			continue
		}
		switch score := value.(type) {
		case float64:
			points[rankNumber] = int(score)
		case int:
			points[rankNumber] = score
		}
	}
	return points
}

func cloneStringIntMap(source map[string]int) map[string]int {
	cloned := make(map[string]int, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
