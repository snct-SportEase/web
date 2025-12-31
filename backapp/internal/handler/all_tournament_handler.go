package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"backapp/internal/websocket"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TournamentHandler struct {
	tournRepo  repository.TournamentRepository
	sportRepo  repository.SportRepository
	teamRepo   repository.TeamRepository
	classRepo  repository.ClassRepository
	eventRepo  repository.EventRepository
	hubManager *websocket.HubManager
}

func NewTournamentHandler(tournRepo repository.TournamentRepository, sportRepo repository.SportRepository, teamRepo repository.TeamRepository, classRepo repository.ClassRepository, eventRepo repository.EventRepository, hubManager *websocket.HubManager) *TournamentHandler {
	return &TournamentHandler{
		tournRepo:  tournRepo,
		sportRepo:  sportRepo,
		teamRepo:   teamRepo,
		classRepo:  classRepo,
		eventRepo:  eventRepo,
		hubManager: hubManager,
	}
}

func (h *TournamentHandler) GetTournamentsByEventHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	tournaments, err := h.tournRepo.GetTournamentsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tournaments", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tournaments)
}

func (h *TournamentHandler) GenerateAllTournamentsPreviewHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	sports, err := h.sportRepo.GetSportsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports for the event", "details": err.Error()})
		return
	}

	generatedTournaments := make([]models.GeneratedTournament, 0)

	for _, eventSport := range sports {
		// 昼競技(noon_game)は /noon-game で別管理するため、トーナメント生成対象から除外
		if eventSport.Location == "noon_game" {
			continue
		}
		roundBusyClasses := make(map[int]map[int]bool)
		sport, err := h.sportRepo.GetSportByID(eventSport.SportID)
		if err != nil {
			continue // Log this error
		}

		teams, err := h.sportRepo.GetTeamsBySportID(sport.ID)
		if err != nil {
			continue // Log this error
		}

		tournamentData, shuffledTeams, err := generateTournamentStructure(teams, roundBusyClasses)
		if err != nil {
			continue
		}

		// Convert []*models.Team to []models.Team
		teamsSlice := make([]models.Team, len(shuffledTeams))
		for i, t := range shuffledTeams {
			if t != nil {
				teamsSlice[i] = *t
			}
		}

		// 本戦トーナメントを追加
		generatedTournaments = append(generatedTournaments, models.GeneratedTournament{
			EventID:        eventID,
			SportID:        sport.ID,
			SportName:      sport.Name,
			TournamentData: *tournamentData,
			ShuffledTeams:  teamsSlice,
		})

		// gym2の場合は敗者戦A/Bブロックトーナメントも生成
		if eventSport.Location == "gym2" {
			// 本戦の一回戦の試合を確認（8試合である必要がある）
			var firstRoundMatches []models.Match
			for _, match := range tournamentData.Matches {
				if match.RoundIndex == 0 && !match.IsBronzeMatch {
					firstRoundMatches = append(firstRoundMatches, match)
				}
			}

			// 一回戦は8試合（16チーム）である必要がある
			if len(firstRoundMatches) == 8 {
				// 敗者戦Aブロックトーナメント
				loserBracketA := generateLoserBracketTournament("A")
				generatedTournaments = append(generatedTournaments, models.GeneratedTournament{
					EventID:        eventID,
					SportID:        sport.ID,
					SportName:      fmt.Sprintf("%s Tournament - 敗者戦Aブロック", sport.Name),
					TournamentData: *loserBracketA,
					ShuffledTeams:  []models.Team{}, // 敗者戦は試合後に決定
				})

				// 敗者戦Bブロックトーナメント
				loserBracketB := generateLoserBracketTournament("B")
				generatedTournaments = append(generatedTournaments, models.GeneratedTournament{
					EventID:        eventID,
					SportID:        sport.ID,
					SportName:      fmt.Sprintf("%s Tournament - 敗者戦Bブロック", sport.Name),
					TournamentData: *loserBracketB,
					ShuffledTeams:  []models.Team{}, // 敗者戦は試合後に決定
				})
			}
		}
	}

	c.JSON(http.StatusOK, generatedTournaments)
}

func (h *TournamentHandler) BulkCreateTournamentsHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var tournamentsToCreate []models.GeneratedTournament
	if err := c.ShouldBindJSON(&tournamentsToCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 昼競技(noon_game)のトーナメントは作成させない（フロント改変や直叩き対策）
	eventSports, err := h.sportRepo.GetSportsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports for the event", "details": err.Error()})
		return
	}
	locationBySportID := make(map[int]string, len(eventSports))
	for _, es := range eventSports {
		locationBySportID[es.SportID] = es.Location
	}
	for _, t := range tournamentsToCreate {
		if t.EventID != 0 && t.EventID != eventID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "EventID mismatch"})
			return
		}
		if locationBySportID[t.SportID] == "noon_game" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "昼競技(noon_game)はトーナメント生成の対象外です"})
			return
		}
	}

	// First, delete existing tournaments for this event
	if err := h.tournRepo.DeleteTournamentsByEventID(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing tournaments", "details": err.Error()})
		return
	}

	// Then, save the new tournaments
	for _, t := range tournamentsToCreate {
		// Convert []models.Team to []*models.Team
		shuffledTeamsPtr := make([]*models.Team, len(t.ShuffledTeams))
		for i := range t.ShuffledTeams {
			shuffledTeamsPtr[i] = &t.ShuffledTeams[i]
		}
		log.Printf("[BulkCreateTournaments] Saving tournament for sport: %s (EventID: %d, SportID: %d)", t.SportName, t.EventID, t.SportID)
		err := h.tournRepo.SaveTournament(t.EventID, t.SportID, t.SportName, &t.TournamentData, shuffledTeamsPtr)
		if err != nil {
			log.Printf("[BulkCreateTournaments] ERROR: Failed to save tournament for sport %s: %v", t.SportName, err)
			// Attempt to rollback or handle partial save might be needed here
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tournament for sport " + t.SportName, "details": err.Error()})
			return
		}
		log.Printf("[BulkCreateTournaments] Successfully saved tournament for sport: %s", t.SportName)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournaments created successfully."})
}

func (h *TournamentHandler) GenerateAllTournamentsHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	if eventIDStr == "" {
		eventIDStr = c.Param("id")
	}
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.tournRepo.DeleteTournamentsByEventID(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing tournaments", "details": err.Error()})
		return
	}

	sports, err := h.sportRepo.GetSportsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports for the event", "details": err.Error()})
		return
	}

	for _, eventSport := range sports {
		// 昼競技(noon_game)は /noon-game で別管理するため、トーナメント生成対象から除外
		if eventSport.Location == "noon_game" {
			continue
		}
		roundBusyClasses := make(map[int]map[int]bool)
		sport, err := h.sportRepo.GetSportByID(eventSport.SportID)
		if err != nil {
			continue
		}

		teams, err := h.sportRepo.GetTeamsBySportID(sport.ID)
		if err != nil {
			continue
		}

		tournamentData, shuffledTeams, err := generateTournamentStructure(teams, roundBusyClasses)
		if err != nil {
			continue
		}

		// 本戦トーナメントを保存
		err = h.tournRepo.SaveTournament(eventID, sport.ID, sport.Name, tournamentData, shuffledTeams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tournament for sport " + sport.Name, "details": err.Error()})
			return
		}

		// gym2の場合は敗者戦A/Bブロックトーナメントも生成
		if eventSport.Location == "gym2" {
			// 本戦の一回戦の試合を確認（8試合である必要がある）
			var firstRoundMatches []models.Match
			for _, match := range tournamentData.Matches {
				if match.RoundIndex == 0 && !match.IsBronzeMatch {
					firstRoundMatches = append(firstRoundMatches, match)
				}
			}

			// 一回戦は8試合（16チーム）である必要がある
			if len(firstRoundMatches) == 8 {
				// 敗者戦Aブロックトーナメント
				loserBracketA := generateLoserBracketTournament("A")
				err = h.tournRepo.SaveTournament(eventID, sport.ID, fmt.Sprintf("%s Tournament - 敗者戦Aブロック", sport.Name), loserBracketA, []*models.Team{})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save loser bracket A tournament for sport " + sport.Name, "details": err.Error()})
					return
				}

				// 敗者戦Bブロックトーナメント
				loserBracketB := generateLoserBracketTournament("B")
				err = h.tournRepo.SaveTournament(eventID, sport.ID, fmt.Sprintf("%s Tournament - 敗者戦Bブロック", sport.Name), loserBracketB, []*models.Team{})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save loser bracket B tournament for sport " + sport.Name, "details": err.Error()})
					return
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournaments generated successfully for all eligible sports."})
}

func generateTournamentStructure(teams []*models.Team, roundBusyClasses map[int]map[int]bool) (*models.TournamentData, []*models.Team, error) {
	availableTeams := make([]*models.Team, 0)
	if roundBusyClasses[0] == nil {
		roundBusyClasses[0] = make(map[int]bool)
	}

	for _, team := range teams {
		if !roundBusyClasses[0][team.ClassID] {
			availableTeams = append(availableTeams, team)
		}
	}

	if len(availableTeams) <= 1 || (len(availableTeams)&(len(availableTeams)-1)) != 0 {
		return nil, nil, fmt.Errorf("invalid number of available teams to form a tournament: %d", len(availableTeams))
	}

	numRounds := int(math.Log2(float64(len(availableTeams))))

	for _, team := range availableTeams {
		for i := 0; i < numRounds; i++ {
			if roundBusyClasses[i] == nil {
				roundBusyClasses[i] = make(map[int]bool)
			}
			roundBusyClasses[i][team.ClassID] = true
		}
	}

	shuffledTeams := make([]*models.Team, len(availableTeams))
	copy(shuffledTeams, availableTeams)

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	r.Shuffle(len(shuffledTeams), func(i, j int) {
		shuffledTeams[i], shuffledTeams[j] = shuffledTeams[j], shuffledTeams[i]
	})

	contestants := make(map[string]models.Contestant)
	for i, team := range shuffledTeams {
		contestantID := "c" + strconv.Itoa(i)
		contestants[contestantID] = models.Contestant{
			Players: []models.Player{{Title: team.Name}},
		}
	}

	rounds := make([]models.Round, numRounds)

	var roundNames []string
	switch numRounds {
	case 1:
		roundNames = []string{"決勝"}
	case 2:
		roundNames = []string{"準決勝", "決勝"}
	case 3:
		roundNames = []string{"一回戦", "準決勝", "決勝"}
	case 4:
		roundNames = []string{"一回戦", "二回戦", "準決勝", "決勝"}
	default:
		roundNames = make([]string, numRounds)
		for i := 0; i < numRounds-2; i++ {
			roundNames[i] = strconv.Itoa(i+1) + "回戦"
		}
		if numRounds > 1 {
			roundNames[numRounds-2] = "準決勝"
		}
		if numRounds > 0 {
			roundNames[numRounds-1] = "決勝"
		}
	}

	for i := 0; i < numRounds; i++ {
		rounds[i] = models.Round{Name: roundNames[i]}
	}

	matches := []models.Match{}

	matchOrder := 0
	for i := 0; i < len(shuffledTeams); i += 2 {
		contestant1ID := "c" + strconv.Itoa(i)
		contestant2ID := "c" + strconv.Itoa(i+1)
		matches = append(matches, models.Match{
			RoundIndex: 0,
			Order:      matchOrder,
			Sides: []models.Side{
				{ContestantID: contestant1ID},
				{ContestantID: contestant2ID},
			},
		})
		matchOrder++
	}

	numMatchesInRound := len(shuffledTeams) / 2
	for roundIndex := 1; roundIndex < numRounds; roundIndex++ {
		matchOrder = 0
		numMatchesInRound /= 2
		for i := 0; i < numMatchesInRound; i++ {
			matches = append(matches, models.Match{
				RoundIndex: roundIndex,
				Order:      matchOrder,
				Sides:      []models.Side{{}, {}},
			})
			matchOrder++
		}
	}

	if numRounds >= 2 {
		matches = append(matches, models.Match{
			RoundIndex:    numRounds - 1,
			Order:         1,
			IsBronzeMatch: true,
			Sides:         []models.Side{{}, {}},
		})
	}

	tournamentData := models.TournamentData{
		Rounds:      rounds,
		Matches:     matches,
		Contestants: contestants,
	}

	return &tournamentData, shuffledTeams, nil
}

// generateLoserBracketTournament は敗者戦トーナメントを生成します
// block: "A" または "B"
// Aブロック: 本戦1-4試合の敗者
// Bブロック: 本戦5-8試合の敗者
func generateLoserBracketTournament(block string) *models.TournamentData {
	// 敗者戦トーナメントのラウンド
	rounds := []models.Round{
		{Name: "一回戦"},
		{Name: "二回戦"},
	}

	loserRound1Round := 1
	loserRound2Round := 2

	// 敗者戦一回戦（2試合）
	loserRound1Matches := []models.Match{
		{
			RoundIndex:          0,
			Order:               0,
			IsLoserBracketMatch: true,
			LoserBracketRound:   &loserRound1Round,
			LoserBracketBlock:   block,
			Sides:               []models.Side{{}, {}}, // 敗者は試合後に決定
		},
		{
			RoundIndex:          0,
			Order:               1,
			IsLoserBracketMatch: true,
			LoserBracketRound:   &loserRound1Round,
			LoserBracketBlock:   block,
			Sides:               []models.Side{{}, {}},
		},
	}

	// 敗者戦二回戦（1試合）- 勝者に10点付与
	loserRound2Matches := []models.Match{
		{
			RoundIndex:          1,
			Order:               0,
			IsLoserBracketMatch: true,
			LoserBracketRound:   &loserRound2Round,
			LoserBracketBlock:   block,
			Sides:               []models.Side{{}, {}},
		},
	}

	matches := append(loserRound1Matches, loserRound2Matches...)

	return &models.TournamentData{
		Rounds:      rounds,
		Matches:     matches,
		Contestants: map[string]models.Contestant{},
	}
}
