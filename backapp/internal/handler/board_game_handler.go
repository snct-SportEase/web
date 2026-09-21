package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type BoardGameHandler struct {
	boardGameRepo repository.BoardGameRepository
	classRepo     repository.ClassRepository
}

func NewBoardGameHandler(boardGameRepo repository.BoardGameRepository, classRepo repository.ClassRepository) *BoardGameHandler {
	return &BoardGameHandler{boardGameRepo: boardGameRepo, classRepo: classRepo}
}

type boardGameParticipant struct {
	ClassID int
}

type boardGameRunRequest struct {
	GameType       string           `json:"game_type" binding:"required"`
	Name           string           `json:"name"`
	Description    *string          `json:"description"`
	Location       string           `json:"location"`
	RulesPDFURL    *string          `json:"rules_pdf_url"`
	ScheduledDate  *string          `json:"scheduled_date"`
	WinPoints      *int             `json:"win_points"`
	RankPoints     map[string]int   `json:"rank_points"`
	RegularMinutes *int             `json:"regular_minutes"`
	FinalMinutes   *int             `json:"final_minutes"`
	Status         string           `json:"status"`
	SeedOrders     map[string][]int `json:"seed_orders"`
}

type boardGamePreset struct {
	Name                string
	Slots               []string
	WinPoints           int
	RankPoints          map[string]int
	RegularMinutes      int
	FinalMinutes        int
	PlayersPerClass     int
	SubstitutesPerClass int
}

func boardGamePresetFor(gameType string) (boardGamePreset, bool) {
	switch gameType {
	case "shogi":
		return boardGamePreset{Name: "将棋", Slots: []string{"A", "B"}, WinPoints: 5, RankPoints: map[string]int{"1": 40, "2": 30, "3": 20, "4": 10}, RegularMinutes: 15, FinalMinutes: 30, PlayersPerClass: 2, SubstitutesPerClass: 1}, true
	case "othello":
		return boardGamePreset{Name: "オセロ", Slots: []string{"MAIN"}, WinPoints: 10, RankPoints: map[string]int{"1": 50, "2": 40, "3": 30, "4": 20}, RegularMinutes: 10, FinalMinutes: 20, PlayersPerClass: 3, SubstitutesPerClass: 0}, true
	default:
		return boardGamePreset{}, false
	}
}

func (h *BoardGameHandler) CreateRun(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil || eventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	var req boardGameRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	userValue, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := userValue.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	}
	input, err := h.buildRunCreate(eventID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	run, err := h.boardGameRepo.CreateRun(input)
	if errors.Is(err, repository.ErrBoardGameRunHasResults) {
		c.JSON(http.StatusConflict, gin.H{"error": "試合結果の登録後は競技種別や組み合わせを変更できません"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create board-game tournament"})
		return
	}
	c.JSON(http.StatusCreated, run)
}

func (h *BoardGameHandler) ListRuns(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil || eventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	publishedOnly := strings.HasPrefix(c.FullPath(), "/api/student/")
	runs, err := h.boardGameRepo.ListRuns(eventID, publishedOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get board-game tournaments"})
		return
	}
	c.JSON(http.StatusOK, runs)
}

func (h *BoardGameHandler) ListEligibleClasses(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil || eventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event classes"})
		return
	}
	type eligibleClass struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	result := make([]eligibleClass, 0, len(classes))
	for _, class := range classes {
		result = append(result, eligibleClass{ID: class.ID, Name: class.Name})
	}
	c.JSON(http.StatusOK, result)
}

func (h *BoardGameHandler) SaveRankings(c *gin.Context) {
	runID, runErr := strconv.Atoi(c.Param("run_id"))
	tournamentID, tournamentErr := strconv.Atoi(c.Param("tournament_id"))
	if runErr != nil || tournamentErr != nil || runID <= 0 || tournamentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}
	activeEventValue, ok := c.Get("active_event_id")
	activeEventID, validActiveEvent := activeEventValue.(int)
	if !ok || !validActiveEvent {
		c.JSON(http.StatusForbidden, gin.H{"error": "No active event found"})
		return
	}
	existingRun, err := h.boardGameRepo.GetRunByID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get board-game tournament"})
		return
	}
	if existingRun == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board-game tournament not found"})
		return
	}
	if existingRun.EventID != activeEventID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tournament does not belong to the active event"})
		return
	}
	var req struct {
		Rankings []models.BoardGameRankingInput `json:"rankings" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	userValue, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := userValue.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	}
	run, err := h.boardGameRepo.SaveRankings(runID, tournamentID, req.Rankings, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *BoardGameHandler) buildRunCreate(eventID int, createdBy string, req *boardGameRunRequest) (*models.BoardGameRunCreate, error) {
	preset, ok := boardGamePresetFor(strings.TrimSpace(req.GameType))
	if !ok {
		return nil, fmt.Errorf("game_type must be shogi or othello")
	}
	if req.Status == "" {
		req.Status = "published"
	}
	if req.Status != "published" {
		return nil, fmt.Errorf("status must be published")
	}
	if req.ScheduledDate != nil && strings.TrimSpace(*req.ScheduledDate) != "" {
		if _, err := time.Parse("2006-01-02", *req.ScheduledDate); err != nil {
			return nil, fmt.Errorf("scheduled_date must be YYYY-MM-DD")
		}
	}
	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event classes: %w", err)
	}
	participants, err := fixedBoardGameParticipants(classes)
	if err != nil {
		return nil, err
	}
	classByID := make(map[int]*models.Class, len(classes))
	for _, class := range classes {
		classByID[class.ID] = class
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = preset.Name
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		location = "ICTメディア室"
	}
	winPoints := preset.WinPoints
	if req.WinPoints != nil {
		winPoints = *req.WinPoints
	}
	rankPoints := preset.RankPoints
	if req.RankPoints != nil {
		rankPoints = req.RankPoints
	}
	regularMinutes := preset.RegularMinutes
	if req.RegularMinutes != nil {
		regularMinutes = *req.RegularMinutes
	}
	finalMinutes := preset.FinalMinutes
	if req.FinalMinutes != nil {
		finalMinutes = *req.FinalMinutes
	}
	if winPoints < 0 || regularMinutes <= 0 || finalMinutes <= 0 {
		return nil, fmt.Errorf("得点と試合時間の設定が不正です")
	}
	for rank := 1; rank <= 4; rank++ {
		if rankPoints[strconv.Itoa(rank)] < 0 {
			return nil, fmt.Errorf("順位点は0以上にしてください")
		}
	}

	tournaments := make([]models.BoardGameTournamentCreate, 0, len(preset.Slots))
	for _, slot := range preset.Slots {
		order, err := boardGameSeedOrder(slot, participants, req.SeedOrders)
		if err != nil {
			return nil, err
		}
		entries := make([]models.BoardGameEntryCreate, 0, len(order))
		for index, classID := range order {
			class := classByID[classID]
			minCapacity, maxCapacity := 1, preset.PlayersPerClass
			teamName := class.Name
			if req.GameType == "shogi" {
				minCapacity, maxCapacity = 1, 1+preset.SubstitutesPerClass
				teamName += " " + slot
			}
			entries = append(entries, models.BoardGameEntryCreate{ClassID: classID, ClassName: class.Name, TeamName: teamName, EntryKey: strings.ToLower(req.GameType + "_" + slot), SeedNumber: index + 1, MinCapacity: minCapacity, MaxCapacity: maxCapacity})
		}
		data := buildBoardGameBracket(entries, req.ScheduledDate)
		tournamentName := name
		if req.GameType == "shogi" {
			tournamentName += " " + slot + "ブロック"
		}
		tournaments = append(tournaments, models.BoardGameTournamentCreate{Name: tournamentName, SlotKey: slot, Entries: entries, Data: data})
	}

	return &models.BoardGameRunCreate{EventID: eventID, GameType: req.GameType, SportName: preset.Name, Name: name, Description: req.Description, Location: location, RulesPDFURL: req.RulesPDFURL, ScheduledDate: req.ScheduledDate, WinPoints: winPoints, RankPoints: rankPoints, RegularMinutes: regularMinutes, FinalMinutes: finalMinutes, PlayersPerClass: preset.PlayersPerClass, SubstitutesPerClass: preset.SubstitutesPerClass, Status: req.Status, CreatedBy: createdBy, Tournaments: tournaments}, nil
}

func fixedBoardGameParticipants(classes []*models.Class) ([]boardGameParticipant, error) {
	if len(classes) != 16 {
		return nil, fmt.Errorf("盤上競技は全16クラスが必要です（現在%dクラス）", len(classes))
	}
	participants := make([]boardGameParticipant, 0, len(classes))
	for _, class := range classes {
		participants = append(participants, boardGameParticipant{ClassID: class.ID})
	}
	return participants, nil
}

func boardGameSeedOrder(slot string, participants []boardGameParticipant, configured map[string][]int) ([]int, error) {
	order := make([]int, 0, len(participants))
	if configured != nil && len(configured[slot]) > 0 {
		order = append(order, configured[slot]...)
	} else {
		for _, participant := range participants {
			order = append(order, participant.ClassID)
		}
	}
	if len(order) != len(participants) {
		return nil, fmt.Errorf("%s のシード順には参加クラスをすべて指定してください", slot)
	}
	want := make(map[int]bool, len(participants))
	for _, participant := range participants {
		want[participant.ClassID] = true
	}
	seen := make(map[int]bool, len(order))
	for _, classID := range order {
		if !want[classID] || seen[classID] {
			return nil, fmt.Errorf("%s のシード順が参加クラスと一致しません", slot)
		}
		seen[classID] = true
	}
	return order, nil
}

func buildBoardGameBracket(entries []models.BoardGameEntryCreate, scheduledDate *string) models.TournamentData {
	roundCount := int(math.Log2(float64(len(entries))))
	roundNames := make([]string, roundCount)
	for round := 0; round < roundCount; round++ {
		switch roundCount - round {
		case 1:
			roundNames[round] = "決勝"
		case 2:
			roundNames[round] = "準決勝"
		default:
			roundNames[round] = strconv.Itoa(round+1) + "回戦"
		}
	}
	rounds := make([]models.Round, roundCount)
	for index, name := range roundNames {
		rounds[index] = models.Round{Name: name}
	}
	contestants := make(map[string]models.Contestant, len(entries))
	for index, entry := range entries {
		contestants["c"+strconv.Itoa(index)] = models.Contestant{Players: []models.Player{{Title: entry.TeamName}}}
	}
	matches := make([]models.Match, 0, len(entries))
	for round := 0; round < roundCount; round++ {
		matchCount := len(entries) >> (round + 1)
		for order := 0; order < matchCount; order++ {
			match := models.Match{RoundIndex: round, Order: order, Sides: []models.Side{{}, {}}, StartTime: boardGameDefaultStartTime(scheduledDate, round, roundCount, order, matchCount)}
			if round == 0 {
				match.Sides[0].ContestantID = "c" + strconv.Itoa(order*2)
				match.Sides[1].ContestantID = "c" + strconv.Itoa(order*2+1)
			}
			matches = append(matches, match)
		}
	}
	if roundCount >= 2 {
		matches = append(matches, models.Match{RoundIndex: roundCount - 1, Order: 1, Sides: []models.Side{{}, {}}, IsBronzeMatch: true, StartTime: boardGameDefaultStartTime(scheduledDate, roundCount-1, roundCount, 1, 2)})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].RoundIndex == matches[j].RoundIndex {
			return matches[i].Order < matches[j].Order
		}
		return matches[i].RoundIndex < matches[j].RoundIndex
	})
	return models.TournamentData{Rounds: rounds, Matches: matches, Contestants: contestants}
}

func boardGameDefaultStartTime(scheduledDate *string, round, roundCount, order, matchCount int) string {
	if scheduledDate == nil || strings.TrimSpace(*scheduledDate) == "" {
		return ""
	}
	hour, minute := 13, 0
	remaining := roundCount - round
	switch remaining {
	case 1:
		hour = 15
	case 2:
		hour = 14
	default:
		if round == 0 {
			hour, minute = 9, 45
			if matchCount > 1 && order >= (matchCount+1)/2 {
				hour, minute = 10, 45
			}
		}
	}
	location := time.FixedZone("JST", 9*60*60)
	date, _ := time.ParseInLocation("2006-01-02", *scheduledDate, location)
	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, location).Format(time.RFC3339)
}
